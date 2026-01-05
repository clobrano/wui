package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ListPicker is a component for selecting from a filtered list of items
type ListPicker struct {
	allItems      []string // All available items
	filteredItems []string // Items matching current filter
	filter        string   // Current filter text
	selectedIndex int      // Currently highlighted item index
	title         string   // Title to display (e.g., "Projects", "Tags")
	maxVisible    int      // Maximum items to show at once
	scrollOffset  int      // Offset for scrolling through items
}

// NewListPicker creates a new list picker with the given items
func NewListPicker(title string, items []string, filter string) ListPicker {
	lp := ListPicker{
		allItems:      items,
		filter:        filter,
		title:         title,
		selectedIndex: 0,
		maxVisible:    10,
		scrollOffset:  0,
	}
	lp.updateFilteredItems()
	return lp
}

// updateFilteredItems filters the items based on current filter
func (lp *ListPicker) updateFilteredItems() {
	if lp.filter == "" {
		lp.filteredItems = lp.allItems
	} else {
		lp.filteredItems = []string{}
		filterLower := strings.ToLower(lp.filter)
		for _, item := range lp.allItems {
			if strings.HasPrefix(strings.ToLower(item), filterLower) {
				lp.filteredItems = append(lp.filteredItems, item)
			}
		}
	}

	// Reset selection if current index is out of bounds
	if lp.selectedIndex >= len(lp.filteredItems) {
		lp.selectedIndex = 0
		lp.scrollOffset = 0
	}

	// Adjust scroll offset if needed
	if lp.selectedIndex < lp.scrollOffset {
		lp.scrollOffset = lp.selectedIndex
	} else if lp.selectedIndex >= lp.scrollOffset+lp.maxVisible {
		lp.scrollOffset = lp.selectedIndex - lp.maxVisible + 1
	}
}

// Update handles key presses for the list picker
func (lp ListPicker) Update(msg tea.Msg) (ListPicker, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
			if lp.selectedIndex > 0 {
				lp.selectedIndex--
				if lp.selectedIndex < lp.scrollOffset {
					lp.scrollOffset = lp.selectedIndex
				}
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
			if lp.selectedIndex < len(lp.filteredItems)-1 {
				lp.selectedIndex++
				if lp.selectedIndex >= lp.scrollOffset+lp.maxVisible {
					lp.scrollOffset = lp.selectedIndex - lp.maxVisible + 1
				}
			}
		}
	}
	return lp, nil
}

// View renders the list picker
func (lp ListPicker) View() string {
	var b strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("212")).
		Padding(0, 1)
	b.WriteString(titleStyle.Render(lp.title))
	b.WriteString("\n\n")

	// Filter display and item count
	infoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Padding(0, 1)

	if lp.filter != "" {
		b.WriteString(infoStyle.Render(fmt.Sprintf("Filter: %s", lp.filter)))
		b.WriteString("\n")
	}

	// Show count info
	b.WriteString(infoStyle.Render(fmt.Sprintf("Showing %d of %d items", len(lp.filteredItems), len(lp.allItems))))
	b.WriteString("\n\n")

	// Items list
	if len(lp.filteredItems) == 0 {
		noItemsStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Padding(0, 1)
		b.WriteString(noItemsStyle.Render("No matches found"))
		b.WriteString("\n")
	} else {
		// Calculate visible range
		endIdx := lp.scrollOffset + lp.maxVisible
		if endIdx > len(lp.filteredItems) {
			endIdx = len(lp.filteredItems)
		}

		// Show scroll indicator if needed
		if lp.scrollOffset > 0 {
			scrollStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).
				Padding(0, 1)
			b.WriteString(scrollStyle.Render("↑ more"))
			b.WriteString("\n")
		}

		// Render visible items
		for i := lp.scrollOffset; i < endIdx; i++ {
			item := lp.filteredItems[i]
			if i == lp.selectedIndex {
				selectedStyle := lipgloss.NewStyle().
					Background(lipgloss.Color("62")).
					Foreground(lipgloss.Color("230")).
					Padding(0, 1)
				b.WriteString(selectedStyle.Render("► " + item))
			} else {
				itemStyle := lipgloss.NewStyle().
					Padding(0, 1)
				b.WriteString(itemStyle.Render("  " + item))
			}
			b.WriteString("\n")
		}

		// Show scroll indicator if there are more items below
		if endIdx < len(lp.filteredItems) {
			scrollStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).
				Padding(0, 1)
			b.WriteString(scrollStyle.Render("↓ more"))
			b.WriteString("\n")
		}
	}

	// Help text
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Padding(1, 1, 0, 1)
	b.WriteString(helpStyle.Render("↑/↓: navigate • enter: select • esc: cancel"))

	// Container style
	containerStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2)

	return containerStyle.Render(b.String())
}

// SelectedItem returns the currently selected item, or empty string if none
func (lp ListPicker) SelectedItem() string {
	if len(lp.filteredItems) == 0 || lp.selectedIndex >= len(lp.filteredItems) {
		return ""
	}
	return lp.filteredItems[lp.selectedIndex]
}

// HasItems returns true if there are any filtered items to select from
func (lp ListPicker) HasItems() bool {
	return len(lp.filteredItems) > 0
}
