'use strict';

/* ── Tab switching ───────────────────────────────────────────────────────── */
const _SEARCH_FILTER_KEY = 'wui-search-filter';

function wuiActivateTab(btn, tabName) {
  // Persist the Search tab filter before leaving it.
  const currentTab = document.querySelector('[name="tab"]')?.value;
  if (currentTab === 'Search') {
    const f = document.querySelector('[name="filter"]')?.value || '';
    sessionStorage.setItem(_SEARCH_FILTER_KEY, f);
  }

  document.querySelectorAll('.tab-btn').forEach(b => b.classList.remove('active'));
  btn.classList.add('active');
  const tabInput = document.querySelector('[name="tab"]');
  if (tabInput) tabInput.value = tabName;

  const filterInput = document.querySelector('[name="filter"]');
  if (tabName === 'Search') {
    // Restore saved Search filter (if any).
    if (filterInput) filterInput.value = sessionStorage.getItem(_SEARCH_FILTER_KEY) || '';
  } else {
    if (filterInput) filterInput.value = '';
  }
}

/* ── Group list (Projects / Tags) ───────────────────────────────────────── */
function wuiOpenGroup(filter) {
  sessionStorage.setItem(_SEARCH_FILTER_KEY, filter);
  window.location = '/?tab=Search&filter=' + encodeURIComponent(filter);
}

/* ── Multi-select state ─────────────────────────────────────────────────── */
let _selected = new Set();
let _multiselectActive = false;

function wuiToggleSelect(checkbox) {
  const uuid = checkbox.dataset.uuid;
  if (checkbox.checked) {
    _selected.add(uuid);
  } else {
    _selected.delete(uuid);
  }
  _updateMultiselectBar();
}

function _enterMultiselect(uuid) {
  _multiselectActive = true;
  _selected.add(uuid);
  const cb = document.querySelector(`.task-checkbox[data-uuid="${uuid}"]`);
  if (cb) cb.checked = true;
  document.querySelectorAll('.task-checkbox').forEach(c => c.closest('td').classList.remove('hidden'));
  _updateMultiselectBar();
}

function wuiExitMultiselect() {
  _multiselectActive = false;
  _selected.clear();
  document.querySelectorAll('.task-checkbox').forEach(c => { c.checked = false; });
  document.getElementById('multiselect-bar').classList.add('hidden');
}

function _updateMultiselectBar() {
  const bar = document.getElementById('multiselect-bar');
  if (_selected.size === 0) {
    wuiExitMultiselect();
    return;
  }
  document.getElementById('multiselect-count').textContent = `${_selected.size} selected`;
  bar.classList.remove('hidden');
}

/* ── Row click / navigation ──────────────────────────────────────────────── */
function wuiRowClick(event, row) {
  if (event.target.closest('.task-checkbox')) return; // handled by toggle
  if (_multiselectActive) {
    const cb = row.querySelector('.task-checkbox');
    cb.checked = !cb.checked;
    wuiToggleSelect(cb);
    return;
  }
  // Navigate to detail page via the link inside the row.
  const link = row.querySelector('.task-link');
  if (link) window.location = link.href;
}

/* ── Swipe gesture detection ─────────────────────────────────────────────── */
let _touchStartX = 0;
let _touchStartY = 0;

function wuiTouchStart(event, row) {
  _touchStartX = event.touches[0].clientX;
  _touchStartY = event.touches[0].clientY;
}

function wuiTouchMove(event) {
  // prevent scroll hijack only when clearly horizontal
  const dx = Math.abs(event.touches[0].clientX - _touchStartX);
  const dy = Math.abs(event.touches[0].clientY - _touchStartY);
  if (dx > dy && dx > 10) event.preventDefault();
}

function wuiTouchEnd(event, row) {
  const dx = event.changedTouches[0].clientX - _touchStartX;
  const dy = event.changedTouches[0].clientY - _touchStartY;

  if (Math.abs(dy) > Math.abs(dx)) return; // vertical scroll, ignore
  if (Math.abs(dx) < 80) return;           // below threshold

  const uuid = row.dataset.uuid;
  if (dx > 0) {
    // Swipe right → mark done
    wuiMarkDone(uuid);
  } else {
    // Swipe left → confirm delete
    if (confirm('Delete this task?')) {
      fetch(`/api/v1/tasks/${uuid}`, { method: 'DELETE' })
        .then(() => { wuiShowToast('Task deleted — Undo', [uuid], 'delete'); _syncAfterAction(); _refreshTaskList(); });
    }
  }
}

/* ── Long-press (mobile multi-select) ───────────────────────────────────── */
let _longPressTimer = null;
let _longPressStartX = 0;
let _longPressStartY = 0;
const _LONGPRESS_MOVE_THRESHOLD = 8; // px — more than this cancels the timer

document.addEventListener('touchstart', e => {
  const row = e.target.closest('.task-row');
  if (!row) return;
  _longPressStartX = e.touches[0].clientX;
  _longPressStartY = e.touches[0].clientY;
  _longPressTimer = setTimeout(() => {
    _enterMultiselect(row.dataset.uuid);
  }, 500);
}, { passive: true });

document.addEventListener('touchmove', e => {
  if (_longPressTimer === null) return;
  const dx = Math.abs(e.touches[0].clientX - _longPressStartX);
  const dy = Math.abs(e.touches[0].clientY - _longPressStartY);
  if (dx > _LONGPRESS_MOVE_THRESHOLD || dy > _LONGPRESS_MOVE_THRESHOLD) {
    clearTimeout(_longPressTimer);
    _longPressTimer = null;
  }
}, { passive: true });

document.addEventListener('touchend', () => {
  clearTimeout(_longPressTimer);
  _longPressTimer = null;
});

/* ── Bulk actions ────────────────────────────────────────────────────────── */
function wuiBulkDone() {
  const uuids = [..._selected];
  Promise.all(uuids.map(u => fetch(`/api/v1/tasks/${u}/done`, { method: 'POST' })))
    .then(() => {
      wuiShowToast(`${uuids.length} task${uuids.length > 1 ? 's' : ''} marked done — Undo`, uuids, 'done');
      _syncAfterAction();
      wuiExitMultiselect();
      _refreshTaskList();
    });
}

function wuiBulkDelete() {
  if (!confirm(`Delete ${_selected.size} task(s)?`)) return;
  const uuids = [..._selected];
  Promise.all(uuids.map(u => fetch(`/api/v1/tasks/${u}`, { method: 'DELETE' })))
    .then(() => {
      wuiShowToast(`${uuids.length} task${uuids.length > 1 ? 's' : ''} deleted — Undo`, uuids, 'delete');
      _syncAfterAction();
      wuiExitMultiselect();
      _refreshTaskList();
    });
}

/* ── Sync (fire-and-forget after every modification) ─────────────────────── */
function _syncAfterAction() {
  fetch('/api/v1/sync', { method: 'POST' }).catch(() => {});
}

/* ── Mark done / delete (single task) ───────────────────────────────────── */
function wuiMarkDone(uuid) {
  fetch(`/api/v1/tasks/${uuid}/done`, { method: 'POST' })
    .then(r => {
      if (r.ok) {
        wuiShowToast('Task marked done — Undo', [uuid], 'done');
        _syncAfterAction();
        _refreshTaskList();
      }
    });
}

function wuiConfirmDelete(uuid) {
  if (!confirm('Delete this task?')) return;
  fetch(`/api/v1/tasks/${uuid}`, { method: 'DELETE' })
    .then(r => {
      if (r.ok) {
        wuiShowToast('Task deleted — Undo', [uuid], 'delete');
        _syncAfterAction();
        history.back();
      }
    });
}

function wuiAddAnnotation(uuid) {
  const input = document.getElementById('new-annotation');
  const text = input.value.trim();
  if (!text) return;
  fetch(`/api/v1/tasks/${uuid}/annotate`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ text }),
  }).then(r => { if (r.ok) { _syncAfterAction(); window.location.reload(); } });
}

function wuiRemoveAnnotation(uuid, description) {
  fetch(`/api/v1/tasks/${uuid}/annotate`, {
    method: 'DELETE',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ description }),
  }).then(r => { if (r.ok) { _syncAfterAction(); window.location.reload(); } });
}

/* ── Undo toast ──────────────────────────────────────────────────────────── */
let _toastTimer = null;
let _lastUndoAction = null;

function wuiShowToast(msg, uuids, action) {
  _lastUndoAction = { uuids, action };
  document.getElementById('undo-toast-msg').textContent = msg;
  document.getElementById('undo-toast').classList.remove('hidden');
  clearTimeout(_toastTimer);
  _toastTimer = setTimeout(wuiDismissToast, 5000);
}

function wuiDismissToast() {
  document.getElementById('undo-toast').classList.add('hidden');
  clearTimeout(_toastTimer);
  _lastUndoAction = null;
}

function wuiUndo() {
  fetch('/api/v1/undo', { method: 'POST' })
    .then(() => { wuiDismissToast(); _syncAfterAction(); _refreshTaskList(); });
}

/* ── Filter panel ────────────────────────────────────────────────────────── */
function wuiToggleFilter() {
  const panel = document.getElementById('filter-panel');
  panel.classList.toggle('hidden');
}

function wuiApplyFilter() {
  const filter = document.getElementById('filter-input').value.trim();
  if (!filter) return;
  fetch('/api/gui/filter-history', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ filter }),
  }).then(() => {
    const tab = document.querySelector('[name="tab"]')?.value || '';
    window.location = `/?tab=${encodeURIComponent(tab)}&filter=${encodeURIComponent(filter)}`;
  });
}

function wuiClearFilter() {
  const tab = document.querySelector('[name="tab"]')?.value || '';
  window.location = `/?tab=${encodeURIComponent(tab)}`;
}

function wuiFillFilter(btn) {
  document.getElementById('filter-input').value = btn.dataset.filter;
}

function wuiDeleteFilter(btn) {
  const filter = btn.dataset.filter;
  fetch('/api/gui/filter-history', {
    method: 'DELETE',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ filter }),
  }).then(() => window.location.reload());
}

function wuiClearAllFilters() {
  fetch('/api/gui/filter-history/clear', { method: 'POST' })
    .then(() => window.location.reload());
}

/* ── Tag chip input ──────────────────────────────────────────────────────── */
function wuiTagKeydown(event) {
  if (event.key !== ' ' && event.key !== 'Enter') return;
  event.preventDefault();
  const input = event.target;
  const tag = input.value.trim().replace(/\s+/g, '_');
  if (!tag) return;

  const container = document.getElementById('tag-chip-container');
  // Dedup
  if (container.querySelector(`input[value="${tag}"]`)) { input.value = ''; return; }

  const chip = document.createElement('span');
  chip.className = 'tag-chip';
  chip.innerHTML = `${tag}<button type="button" onclick="wuiRemoveTag(this)">✕</button><input type="hidden" name="tags" value="${tag}">`;
  container.insertBefore(chip, input);
  input.value = '';
}

function wuiRemoveTag(btn) {
  btn.closest('.tag-chip').remove();
}

/* ── Depends field ───────────────────────────────────────────────────────── */
function wuiRemoveDep(btn) {
  btn.closest('.dep-chip').remove();
}

/* ── Task form mode toggle ───────────────────────────────────────────────── */
function wuiSetMode(mode) {
  const formDiv = document.getElementById('mode-form');
  const rawDiv  = document.getElementById('mode-raw');
  const btnForm = document.getElementById('btn-form');
  const btnRaw  = document.getElementById('btn-raw');
  const descInput = document.getElementById('f-desc');

  if (mode === 'raw') {
    formDiv.classList.add('hidden');
    rawDiv.classList.remove('hidden');
    btnRaw.classList.add('active');
    btnForm.classList.remove('active');
    // Remove required so the hidden description field doesn't block submission.
    if (descInput) descInput.removeAttribute('required');
  } else {
    rawDiv.classList.add('hidden');
    formDiv.classList.remove('hidden');
    btnForm.classList.add('active');
    btnRaw.classList.remove('active');
    if (descInput) descInput.setAttribute('required', '');
  }
}

/* ── Auto-commit pending tag on form submit ──────────────────────────────── */
document.addEventListener('submit', function() {
  const input = document.getElementById('tag-input');
  if (!input) return;
  const tag = input.value.trim().replace(/\s+/g, '_');
  if (!tag) return;
  const container = document.getElementById('tag-chip-container');
  if (!container.querySelector(`input[value="${CSS.escape(tag)}"]`)) {
    const chip = document.createElement('span');
    chip.className = 'tag-chip';
    chip.innerHTML = `${tag}<button type="button" onclick="wuiRemoveTag(this)">✕</button><input type="hidden" name="tags" value="${tag}">`;
    container.insertBefore(chip, input);
  }
  input.value = '';
});

/* ── Task list refresh ───────────────────────────────────────────────────── */
function _refreshTaskList() {
  const container = document.getElementById('task-list-container');
  if (!container) return;
  const tab    = document.querySelector('[name="tab"]')?.value || '';
  const filter = document.querySelector('[name="filter"]')?.value || '';
  const q      = document.getElementById('search-input')?.value || '';
  const url    = `/api/gui/tasks?tab=${encodeURIComponent(tab)}&filter=${encodeURIComponent(filter)}&q=${encodeURIComponent(q)}`;
  htmx.ajax('GET', url, { target: '#task-list-container', swap: 'innerHTML' });
}

/* ── Reconnection overlay ────────────────────────────────────────────────── */
let _reconnectPolling = false;

document.addEventListener('htmx:responseError', _onConnectionError);
document.addEventListener('htmx:sendError',     _onConnectionError);

function _onConnectionError(evt) {
  const status = evt.detail?.xhr?.status;
  // Show overlay on network errors or 5xx
  if (!status || status >= 500) {
    _showReconnectOverlay();
  }
}

function _showReconnectOverlay() {
  document.getElementById('reconnect-overlay').classList.remove('hidden');
  if (_reconnectPolling) return;
  _reconnectPolling = true;
  _pollAPI();
}

function _pollAPI() {
  fetch('/api/v1/version')
    .then(r => {
      if (r.ok) {
        document.getElementById('reconnect-overlay').classList.add('hidden');
        _reconnectPolling = false;
        _refreshTaskList();
      } else {
        setTimeout(_pollAPI, 2000);
      }
    })
    .catch(() => setTimeout(_pollAPI, 2000));
}
