# Bug Report: Tab Refresh ("r" hotkey) Doesn't Restore Tab State

## Summary
When in a tab that supports grouped views (e.g., the "Project" tab), drilling into a specific group and then pressing "r" to refresh does not restore the original tab view. Instead, it shows ALL pending tasks, losing the current view context.

## Steps to Reproduce

1. Navigate to the **Project** tab
2. Observe the list of projects displayed with completion percentages
3. Press **Enter** on a project to drill into it (e.g., "Home" project)
4. Observe that only tasks for that specific project are shown (expected behavior)
5. Press **"r"** to refresh the view

## Expected Behavior

When pressing "r" to refresh, one of the following should occur:
- **Option A**: Restore the original Project tab view (showing the list of projects) - similar to pressing Esc
- **Option B**: Maintain the drilled-in view (showing only the selected project's tasks) with refreshed data

## Actual Behavior

Pressing "r" shows **ALL pending tasks** regardless of the currently selected project, effectively losing the tab's context and state.

## Root Cause Analysis

### Code Location: `/home/user/wui/internal/tui/model.go`

#### 1. Refresh Handler (lines 481-485)
```go
if m.keyMatches(keyPressed, "refresh") {
    m.isLoading = true
    isSearchTab := m.currentSection != nil && m.currentSection.Name == "Search"
    return m, loadTasksCmd(m.service, m.activeFilter, isSearchTab)
}
```
The refresh handler loads all tasks matching the current filter, but doesn't consider the `selectedGroup` context.

#### 2. TasksLoadedMsg Handler (lines 314-318)
```go
} else {
    // Normal view or drilling into a group
    // Update task list component with actual tasks
    m.taskList.SetTasks(m.tasks)
}
```

When `inGroupView = false` (which is the case when drilled into a project), the handler simply sets **all loaded tasks** to the task list, completely ignoring the `selectedGroup` state.

### State Variables Involved

- **`inGroupView`** (bool): `false` when drilled into a specific project/group
- **`selectedGroup`** (*core.TaskGroup): Contains the selected project but is **ignored** during refresh
- **`groups`** ([]core.TaskGroup): The project list, only refreshed when `inGroupView = true`

## Impact

- **Severity**: Medium
- **User Experience**: Confusing and inconsistent behavior
- **Workaround**: Press **Esc** to return to the project list view, then press **"r"** to refresh

## Comparison: Escape Key Behavior

The **Esc** key correctly handles returning to the group list view:

**File**: `/home/user/wui/internal/tui/model.go` (lines 446-466)
```go
if keyPressed == "esc" {
    // Go back to group list if we drilled into a group
    if !m.inGroupView && m.selectedGroup != nil &&
       (m.sections.IsProjectsView() || m.sections.IsTagsView()) {
        m.inGroupView = true
        m.selectedGroup = nil
        // Recompute groups from all tasks
        if m.sections.IsProjectsView() {
            m.groups = core.GroupProjectsByHierarchy(m.projectSummaries, m.tasks)
        } else if m.sections.IsTagsView() {
            m.groups = core.GroupByTag(m.tasks)
        }
        m.taskList.SetGroups(m.groups)
    }
}
```

## Suggested Fix

The refresh handler should check if we're in a drilled-in group view and either:

**Option A (Recommended)**: Reset to group list view (consistent with "refresh means restore original view")
```go
if m.keyMatches(keyPressed, "refresh") {
    m.isLoading = true

    // If drilled into a group, reset to group list view on refresh
    if !m.inGroupView && m.selectedGroup != nil &&
       (m.sections.IsProjectsView() || m.sections.IsTagsView()) {
        m.inGroupView = true
        m.selectedGroup = nil
    }

    isSearchTab := m.currentSection != nil && m.currentSection.Name == "Search"
    return m, loadTasksCmd(m.service, m.activeFilter, isSearchTab)
}
```

**Option B**: Filter tasks by selected group after refresh
```go
// In TasksLoadedMsg handler, when not in group view
} else {
    // If we have a selected group, filter tasks to that group
    tasksToDisplay := m.tasks
    if m.selectedGroup != nil {
        tasksToDisplay = filterTasksByGroup(m.tasks, m.selectedGroup)
    }
    m.taskList.SetTasks(tasksToDisplay)
}
```

## Affected Files

- `/home/user/wui/internal/tui/model.go` (lines 481-485, 314-318)
- Potentially affects both **Projects** and **Tags** tabs

## Additional Notes

This bug is consistent across any tab that supports grouped views (Projects, Tags), not just the Project tab specifically.

---

**Date Reported**: 2025-11-04
**Reported By**: User testing
**Priority**: Medium
