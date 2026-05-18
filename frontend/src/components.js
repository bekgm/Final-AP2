import { api } from './api.js';
import { router } from './router.js';

// ── Toast ────────────────────────────────────────
let toastContainer;
export function initToast() {
  toastContainer = document.createElement('div');
  toastContainer.className = 'toast-container';
  document.body.appendChild(toastContainer);
}
export function showToast(message, type = 'info') {
  const el = document.createElement('div');
  el.className = `toast toast-${type}`;
  el.textContent = message;
  toastContainer.appendChild(el);
  setTimeout(() => { el.style.opacity = '0'; setTimeout(() => el.remove(), 300); }, 3000);
}

// ── Helpers ──────────────────────────────────────
export function formatDate(dateStr) {
  if (!dateStr) return '';
  const d = new Date(dateStr);
  return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
}

export function formatBudget(b) {
  return `$${Number(b).toLocaleString()}`;
}

export function statusBadge(status) {
  const s = (status || '').replace('JOB_STATUS_', '').replace('APPLICATION_STATUS_', '').toLowerCase();
  const map = { open: 'open', in_progress: 'progress', closed: 'closed', pending: 'pending', accepted: 'accepted', rejected: 'rejected' };
  return `<span class="badge badge-${map[s] || 'closed'}">${s.replace('_',' ')}</span>`;
}

export function getInitials(name) {
  if (!name) return '?';
  return name.split(' ').map(w => w[0]).join('').toUpperCase().slice(0, 2);
}

// ── Navbar ───────────────────────────────────────
export function renderNavbar() {
  const user = api.getUser();
  const isAuth = api.isAuthenticated();

  return `<nav class="navbar"><div class="navbar-inner">
    <a href="#/" class="navbar-logo">FreelanceHub</a>
    <div class="navbar-links">
      <a href="#/jobs" class="${isActiveRoute('/jobs')}">Browse Jobs</a>
      ${isAuth ? `
        <a href="#/messages" class="${isActiveRoute('/messages')}">Messages</a>
        <a href="#/profile" class="${isActiveRoute('/profile')}"><span class="hide-mobile">👤 </span>${user?.name || 'Profile'}</a>
        <button id="logout-btn" class="btn btn-ghost btn-sm">Logout</button>
      ` : `
        <a href="#/login" class="btn btn-ghost btn-sm">Sign In</a>
        <a href="#/register" class="btn btn-primary btn-sm">Get Started</a>
      `}
    </div>
  </div></nav>`;
}

function isActiveRoute(path) {
  const hash = window.location.hash.slice(1) || '/';
  return hash.startsWith(path) ? 'active' : '';
}

export function bindNavbar() {
  const btn = document.getElementById('logout-btn');
  if (btn) {
    btn.addEventListener('click', () => {
      api.clearAuth();
      showToast('Logged out successfully', 'info');
      router.navigate('/');
    });
  }
}

// ── Modal ────────────────────────────────────────
export function showModal(content) {
  const overlay = document.createElement('div');
  overlay.className = 'modal-overlay';
  overlay.innerHTML = `<div class="modal">${content}</div>`;
  overlay.addEventListener('click', (e) => { if (e.target === overlay) overlay.remove(); });
  document.body.appendChild(overlay);
  return overlay;
}

export function closeModal() {
  document.querySelector('.modal-overlay')?.remove();
}
