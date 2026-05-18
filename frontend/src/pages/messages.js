import { api } from '../api.js';
import { router } from '../router.js';
import { renderNavbar, bindNavbar, showToast, getInitials } from '../components.js';

export function renderMessages(app) {
  if (!api.isAuthenticated()) {
    router.navigate('/login');
    return;
  }

  app.innerHTML = `
    ${renderNavbar()}
    <div class="messages-page">
      <div class="dialogs-sidebar">
        <div class="dialogs-header">
          <h2>Messages</h2>
        </div>
        <div class="dialogs-list" id="dialogs-list">
          <div class="loading-center"><div class="spinner"></div></div>
        </div>
      </div>
      <div class="chat-area" id="chat-area">
        <div class="chat-empty">
          <div style="text-align:center">
            <div style="font-size:3rem;margin-bottom:16px;">💬</div>
            <p>Select a conversation to start messaging</p>
          </div>
        </div>
      </div>
    </div>
  `;

  bindNavbar();

  let activeDialog = null;
  const userCache = {};

  async function getUserName(userId) {
    if (userCache[userId]) return userCache[userId];
    try {
      const u = await api.getUserById(userId);
      userCache[userId] = u.name || u.email || userId.slice(0, 8);
      return userCache[userId];
    } catch {
      return userId.slice(0, 8);
    }
  }

  async function loadDialogs() {
    const list = document.getElementById('dialogs-list');
    try {
      const dialogs = await api.getDialogs();
      if (!dialogs || dialogs.length === 0) {
        list.innerHTML = `<div class="empty-state" style="padding:40px 20px"><div class="empty-state-icon">📭</div><h3 style="font-size:1rem">No conversations</h3><p style="font-size:0.85rem">Start a conversation from a job page</p></div>`;
        return;
      }

      // Merge dialogs by otherUserId
      const uniqueDialogs = new Map();
      for (const d of dialogs) {
        const otherId = d.other_user_id || d.otherUserId;
        const msgTime = new Date(d.last_message?.created_at || d.lastMessage?.createdAt || 0).getTime();
        
        if (!uniqueDialogs.has(otherId)) {
          uniqueDialogs.set(otherId, { ...d, time: msgTime });
        } else {
          const existing = uniqueDialogs.get(otherId);
          if (msgTime > existing.time) {
            uniqueDialogs.set(otherId, { ...d, time: msgTime });
          }
        }
      }
      const mergedDialogs = Array.from(uniqueDialogs.values()).sort((a, b) => b.time - a.time);

      const names = await Promise.all(mergedDialogs.map(d => getUserName(d.other_user_id || d.otherUserId)));

      list.innerHTML = mergedDialogs.map((d, i) => {
        const otherUserId = d.other_user_id || d.otherUserId;
        const lastMsg = d.last_message || d.lastMessage;
        const preview = lastMsg?.content || 'No messages yet';
        const unread = d.unread_count || d.unreadCount || 0;
        return `
          <div class="dialog-item" data-user-id="${otherUserId}">
            <div class="dialog-avatar">${getInitials(names[i])}</div>
            <div class="dialog-info">
              <div class="dialog-name">${escapeHtml(names[i])}</div>
              <div class="dialog-preview">${escapeHtml(preview)}</div>
            </div>
            ${unread > 0 ? `<div class="dialog-unread">${unread}</div>` : ''}
          </div>
        `;
      }).join('');

      list.querySelectorAll('.dialog-item').forEach(item => {
        item.addEventListener('click', () => {
          list.querySelectorAll('.dialog-item').forEach(i => i.classList.remove('active'));
          item.classList.add('active');
          openChat(item.dataset.userId);
        });
      });
    } catch (err) {
      list.innerHTML = `<div class="empty-state" style="padding:40px 20px"><p style="color:var(--danger);font-size:0.85rem">${err.message}</p></div>`;
    }
  }

  async function openChat(otherUserId) {
    activeDialog = { otherUserId, projectId: '' };
    localStorage.setItem(`chat_read_${otherUserId}`, Date.now());
    bindNavbar(); // update the badge
    
    const chatArea = document.getElementById('chat-area');
    const name = await getUserName(otherUserId);

    chatArea.innerHTML = `
      <div class="chat-header">
        <a href="#/profile/${otherUserId}" class="dialog-avatar" style="width:36px;height:36px;font-size:0.85rem;text-decoration:none;color:inherit;cursor:pointer;">${getInitials(name)}</a>
        <a href="#/profile/${otherUserId}" class="chat-header-name" style="text-decoration:none;color:inherit;cursor:pointer;">${escapeHtml(name)}</a>
      </div>
      <div class="chat-messages" id="chat-messages">
        <div class="loading-center"><div class="spinner"></div></div>
      </div>
      <div class="chat-input-area">
        <input class="form-input" id="msg-input" placeholder="Type a message..." />
        <button class="btn btn-primary" id="send-btn">Send</button>
      </div>
    `;

    await loadMessages(otherUserId, '');

    const sendMsg = async () => {
      const input = document.getElementById('msg-input');
      const text = input.value.trim();
      if (!text) return;
      input.value = '';
      try {
        await api.sendMessage(otherUserId, text, '');
        await loadMessages(otherUserId, '');
      } catch (err) {
        showToast(err.message, 'error');
      }
    };

    document.getElementById('send-btn').addEventListener('click', sendMsg);
    document.getElementById('msg-input').addEventListener('keydown', (e) => {
      if (e.key === 'Enter') sendMsg();
    });
  }

  async function loadMessages(otherUserId, projectId) {
    const container = document.getElementById('chat-messages');
    try {
      const messages = await api.getMessages(otherUserId, projectId || '');
      if (!messages || messages.length === 0) {
        container.innerHTML = '<div class="chat-empty"><p>No messages yet. Say hello! 👋</p></div>';
        return;
      }
      const myId = api.userId;
      const reversed = [...messages].reverse();
      container.innerHTML = reversed.map(m => {
        const senderId = m.sender_id || m.senderId;
        const isMine = senderId === myId;
        const time = m.timestamp ? new Date(m.timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }) : '';
        return `
          <div class="message-bubble ${isMine ? 'message-sent' : 'message-received'}">
            <div>${escapeHtml(m.content)}</div>
            <div class="message-time">${time}</div>
          </div>
        `;
      }).join('');
      container.scrollTop = container.scrollHeight;
    } catch (err) {
      container.innerHTML = `<div class="chat-empty"><p style="color:var(--danger)">${err.message}</p></div>`;
    }
  }

  loadDialogs();

  const pollInterval = setInterval(() => {
    const dialogsListEl = document.getElementById('dialogs-list');
    if (!dialogsListEl) {
      clearInterval(pollInterval);
      return;
    }
    loadDialogs();
    if (activeDialog) {
      loadMessages(activeDialog.otherUserId, activeDialog.projectId);
      localStorage.setItem(`chat_read_${activeDialog.otherUserId}`, Date.now());
      bindNavbar();
    }
  }, 3000);
}

function escapeHtml(str) {
  const div = document.createElement('div');
  div.textContent = str || '';
  return div.innerHTML;
}
