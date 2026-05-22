import { api } from './api.js';
import { router } from './router.js';

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

export function renderNavbar() {
  const user = api.getUser();
  const isAuth = api.isAuthenticated();

  return `<nav class="navbar"><div class="navbar-inner">
    <a href="#/" class="navbar-logo">FreelanceHub</a>
    <div class="navbar-links">
      <a href="#/jobs" class="${isActiveRoute('/jobs')}">Browse Jobs</a>
      ${isAuth ? `
        <a href="#/messages" class="${isActiveRoute('/messages')}" id="nav-messages-link" style="display:flex;align-items:center;">
          Messages
          <span id="nav-unread-badge" style="display:none; background:var(--danger); color:white; font-size:0.7rem; font-weight:bold; border-radius:9999px; padding:2px 6px; margin-left:6px; line-height:1;">0</span>
        </a>
        <a href="#/profile" class="${isActiveRoute('/profile')}">${user?.name || 'Profile'}</a>
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

  if (api.isAuthenticated()) {
    api.getDialogs().then(dialogs => {
      if (!dialogs) return;
      let totalUnread = 0;
      const seenUsers = new Set();
      // Calculate unread using local storage since backend doesn't support it yet
      for (const d of dialogs) {
        const uid = d.other_user_id || d.otherUserId;
        if (!seenUsers.has(uid)) {
          seenUsers.add(uid);
          
          const lastMsg = d.last_message || d.lastMessage;
          const msgTime = new Date(lastMsg?.timestamp || lastMsg?.created_at || lastMsg?.createdAt || 0).getTime();
          const readTime = Number(localStorage.getItem(`chat_read_${uid}`) || 0);
          
          // If the last message is newer than our read time AND we are not the sender
          const senderId = lastMsg?.sender_id || lastMsg?.senderId;
          if (senderId !== api.userId && msgTime > readTime) {
             totalUnread += 1;
          }
        }
      }
      const badge = document.getElementById('nav-unread-badge');
      if (badge && totalUnread > 0) {
        badge.textContent = totalUnread > 9 ? '9+' : totalUnread;
        badge.style.display = 'inline-block';
      } else if (badge) {
        badge.style.display = 'none';
      }
    }).catch(e => console.error("Failed to load unread count", e));
  }
}

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
