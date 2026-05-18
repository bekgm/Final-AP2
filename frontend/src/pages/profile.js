import { api } from '../api.js';
import { router } from '../router.js';
import { renderNavbar, bindNavbar, showToast, showModal, closeModal, getInitials } from '../components.js';

export async function renderProfile(app, targetUserId = null) {
  if (!api.isAuthenticated()) {
    router.navigate('/login');
    return;
  }

  app.innerHTML = `${renderNavbar()}<div class="profile-page container"><div class="loading-center"><div class="spinner"></div></div></div>`;
  bindNavbar();

  const isOwnProfile = !targetUserId || targetUserId === api.userId;
  const uidToFetch = targetUserId || api.userId;

  let user;
  try {
    user = await api.getUserById(uidToFetch);
    if (isOwnProfile) api.setUser(user);
  } catch (err) {
    app.querySelector('.profile-page').innerHTML = `<div class="empty-state"><h3>Failed to load profile</h3><p>${err.message}</p></div>`;
    return;
  }

  let roleName = String(user.role || '').replace('ROLE_', '').toLowerCase();
  if (roleName === '1') roleName = 'client';
  if (roleName === '2') roleName = 'freelancer';
  const roleLabel = roleName === 'client' ? 'Client' : 'Freelancer';

  const bioParts = (user.bio || '').split('|PORTFOLIO|');
  const bioText = bioParts[0];
  const portfolioLink = bioParts[1] || '';

  const allReviews = JSON.parse(localStorage.getItem('freelancer_reviews') || '[]');
  const userReviews = allReviews.filter(r => r.freelancerId === uidToFetch);

  app.querySelector('.profile-page').innerHTML = `
    <div style="margin-bottom:20px"><a href="#" onclick="history.back(); return false;" class="btn btn-ghost btn-sm">← Back</a></div>
    <div class="profile-header-section">
      <div class="profile-avatar-large">${getInitials(user.name)}</div>
      <div>
        <div class="profile-name">${escapeHtml(user.name)}</div>
        <div class="profile-email">${escapeHtml(user.email)}</div>
        <span class="badge badge-${roleName}" style="margin-top:6px;">${roleLabel}</span>
        ${bioText ? `<div class="profile-bio">${escapeHtml(bioText)}</div>` : ''}
        ${roleName === 'freelancer' && portfolioLink ? `<div style="margin-top:10px;"><a href="${escapeHtml(portfolioLink)}" target="_blank" style="color:var(--primary);text-decoration:underline;">🔗 View Portfolio</a></div>` : ''}
        ${roleName === 'freelancer' && user.skills?.length ? `<div class="profile-skills">${user.skills.map(s => `<span class="skill-tag">${escapeHtml(s)}</span>`).join('')}</div>` : ''}
      </div>
    </div>

    ${isOwnProfile ? `<button class="btn btn-secondary" id="edit-profile-btn" style="margin-bottom:32px;">✏️ Edit Profile</button>` : ''}

    ${roleName === 'freelancer' ? `
      <div class="profile-reviews-section" style="margin-top: 32px; margin-bottom: 32px;">
        <h2 style="font-size:1.3rem;font-weight:700;margin-bottom:16px;">Client Reviews ⭐</h2>
        ${userReviews.length === 0 ? `
          <p style="color:var(--text-secondary);font-style:italic;">No reviews yet from clients.</p>
        ` : `
          <div style="display:flex; flex-direction:column; gap:16px;">
            ${userReviews.map(r => `
              <div class="card" style="padding:20px; border: 1px solid var(--border); background: var(--bg-lighter);">
                <div style="display:flex; justify-content:space-between; align-items:center; margin-bottom:8px;">
                  <span style="font-weight:600; color:var(--text);">${escapeHtml(r.clientName)}</span>
                  <span style="font-size:0.8rem; color:var(--text-secondary);">${formatDate(r.date)}</span>
                </div>
                <div style="font-size:0.85rem; font-weight:600; color:var(--primary); margin-bottom:8px;">Project: ${escapeHtml(r.jobTitle)}</div>
                <p style="font-size:0.9rem; color:var(--text); line-height:1.5; font-style:italic;">"${escapeHtml(r.feedback)}"</p>
                <div style="color:#fbbf24; margin-top:8px;">★★★★★</div>
              </div>
            `).join('')}
          </div>
        `}
      </div>
    ` : ''}

    <h2 style="font-size:1.3rem;font-weight:700;margin-bottom:16px;">${roleName === 'client' ? 'Posted Jobs' : 'Activity'}</h2>
    <div id="user-jobs">
      <div class="loading-center"><div class="spinner"></div></div>
    </div>
  `;

  // Load user's jobs
  if (roleName === 'client') {
    try {
      const data = await api.listJobs(1, 50, api.userId);
      const jobs = data.jobs || [];
      const container = document.getElementById('user-jobs');
      if (jobs.length === 0) {
        container.innerHTML = `<div class="empty-state"><div class="empty-state-icon">📋</div><h3>No jobs posted yet</h3><p>Post your first job to get started!</p><a href="#/jobs" class="btn btn-primary">Browse Jobs</a></div>`;
      } else {
        container.innerHTML = `<div class="jobs-grid">${jobs.map(j => `
          <div class="card job-card" data-job-id="${j.id}" style="cursor:pointer">
            <div class="job-card-header">
              <div class="job-card-title">${escapeHtml(j.title)}</div>
              <div class="job-card-budget">$${Number(j.budget).toLocaleString()}</div>
            </div>
            <div class="job-card-desc">${escapeHtml(j.description)}</div>
            <div class="job-card-footer">${statusBadge(j.status)}<span class="job-card-meta">${formatDate(j.created_at)}</span></div>
          </div>
        `).join('')}</div>`;
        container.querySelectorAll('.job-card').forEach(c => c.addEventListener('click', () => router.navigate(`/jobs/${c.dataset.jobId}`)));
      }
    } catch (err) {
      document.getElementById('user-jobs').innerHTML = `<p style="color:var(--danger)">${err.message}</p>`;
    }
  } else {
    document.getElementById('user-jobs').innerHTML = `<div class="empty-state"><div class="empty-state-icon">🚀</div><h3>Your applications</h3><p>Browse jobs and apply to start freelancing!</p><a href="#/jobs" class="btn btn-primary">Find Work</a></div>`;
  }

  // Edit profile modal
  const editBtn = document.getElementById('edit-profile-btn');
  if (editBtn) {
    editBtn.addEventListener('click', () => {
    showModal(`
      <h2>Edit Profile</h2>
      <form id="edit-profile-form">
        <div class="form-group">
          <label class="form-label" for="edit-name">Name</label>
          <input class="form-input" id="edit-name" value="${escapeHtml(user.name || '')}" required />
        </div>
        <div class="form-group">
          <label class="form-label" for="edit-bio">Bio</label>
          <textarea class="form-textarea" id="edit-bio" placeholder="Tell us about yourself...">${escapeHtml(bioText)}</textarea>
        </div>
        ${roleName === 'freelancer' ? `
        <div class="form-group">
          <label class="form-label" for="edit-portfolio">Portfolio Link</label>
          <input class="form-input" id="edit-portfolio" type="url" value="${escapeHtml(portfolioLink)}" placeholder="https://github.com/..." />
        </div>
        <div class="form-group">
          <label class="form-label" for="edit-skills">Skills (comma-separated)</label>
          <input class="form-input" id="edit-skills" value="${(user.skills || []).join(', ')}" placeholder="Go, JavaScript, Docker..." />
        </div>
        ` : ''}
        <div class="modal-actions">
          <button type="button" class="btn btn-secondary" id="cancel-edit">Cancel</button>
          <button type="submit" class="btn btn-primary" id="save-profile">Save Changes</button>
        </div>
      </form>
    `);
    document.getElementById('cancel-edit').addEventListener('click', closeModal);
    document.getElementById('edit-profile-form').addEventListener('submit', async (e) => {
      e.preventDefault();
      const btn = document.getElementById('save-profile');
      btn.disabled = true;
      btn.textContent = 'Saving...';
      try {
        const skillsInput = document.getElementById('edit-skills');
        const skills = skillsInput ? skillsInput.value.split(',').map(s => s.trim()).filter(Boolean) : [];
        const portfolioInput = document.getElementById('edit-portfolio');
        const finalBio = document.getElementById('edit-bio').value + (portfolioInput && portfolioInput.value ? '|PORTFOLIO|' + portfolioInput.value : '');

        await api.updateUser(api.userId, {
          name: document.getElementById('edit-name').value,
          bio: finalBio,
          skills,
        });
        closeModal();
        showToast('Profile updated!', 'success');
        renderProfile(app);
      } catch (err) {
        showToast(err.message, 'error');
        btn.disabled = false;
        btn.textContent = 'Save Changes';
      }
    });
  });
  }
}

function escapeHtml(str) {
  const div = document.createElement('div');
  div.textContent = str || '';
  return div.innerHTML;
}

function statusBadge(status) {
  const s = (status || '').replace('JOB_STATUS_', '').toLowerCase();
  const map = { open: 'open', in_progress: 'progress', closed: 'closed' };
  return `<span class="badge badge-${map[s] || 'closed'}">${s.replace('_',' ')}</span>`;
}

function formatDate(dateStr) {
  if (!dateStr) return '';
  return new Date(dateStr).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
}
