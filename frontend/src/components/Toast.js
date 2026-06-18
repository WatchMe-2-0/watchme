// ═══════════════════════════════════════════════════════════════════
// WATCHME — Toast Notification Component
// ═══════════════════════════════════════════════════════════════════

const TOAST_DURATION = 4000;

export function showToast(message, type = 'info') {
  const container = document.getElementById('toast-container');
  if (!container) return;

  const toast = document.createElement('div');
  toast.className = `toast toast-${type} glass`;
  toast.textContent = message;

  container.appendChild(toast);

  // Auto-remove after duration
  setTimeout(() => {
    toast.style.animation = 'fadeIn 300ms ease reverse both';
    setTimeout(() => toast.remove(), 300);
  }, TOAST_DURATION);

  // Click to dismiss
  toast.addEventListener('click', () => {
    toast.style.animation = 'fadeIn 200ms ease reverse both';
    setTimeout(() => toast.remove(), 200);
  });
}

export function toastSuccess(message) { showToast(message, 'success'); }
export function toastError(message) { showToast(message, 'error'); }
export function toastInfo(message) { showToast(message, 'info'); }
