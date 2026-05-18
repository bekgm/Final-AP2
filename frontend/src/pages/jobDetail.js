import { api } from '../api.js';
import { router } from '../router.js';
import { renderNavbar, bindNavbar, showToast, showModal, closeModal, formatBudget, statusBadge, formatDate, getInitials } from '../components.js';

export async function renderJobDetail(app, jobId) {
  const user = api.getUser();
  const isAuth = api.isAuthenticated();
  const isClient = user?.role === 'ROLE_CLIENT' || user?.role === 'client';
  const isFreelancer = user?.role === 'ROLE_FREELANCER' || user?.role === 'freelancer';

  app.innerHTML = `${renderNavbar()}<div class="job-detail container"><div class="loading-center"><div class="spinner"></div></div></div>`;
  bindNavbar();

  let job;
  try {
    job = await api.getJob(jobId);
  } catch (err) {
    app.querySelector('.job-detail').innerHTML = `<div class="empty-state"><div class="empty-state-icon">😕</div><h3>Job not found</h3><p>${err.message}</p><a href="#/jobs" class="btn btn-primary">Back to Jobs</a></div>`;
    return;
  }

  let clientName = 'Unknown Client';
  try {
    const clientData = await api.getUserById(job.client_id);
    clientName = clientData.name || clientData.email || 'Unknown';
  } catch { /* ignore */ }

  const isOwner = user?.id === job.client_id;
  const jobOpen = job.status === 'JOB_STATUS_OPEN';

  const appliedJobs = JSON.parse(localStorage.getItem(`applied_${user?.id}`) || '[]');
  const hasApplied = appliedJobs.includes(jobId);

  app.querySelector('.job-detail').innerHTML = `
    <div style="margin-bottom:20px"><a href="#/jobs" class="btn btn-ghost btn-sm">← Back to Jobs</a></div>
    <div class="job-detail-grid">
      <div class="job-detail-main">
        <h1>${escapeHtml(job.title)}</h1>
        <div class="job-detail-info">
          ${statusBadge(job.status)}
          <div class="job-detail-info-item">💰 ${formatBudget(job.budget)}</div>
          <div class="job-detail-info-item">👤 ${escapeHtml(clientName)}</div>
          <div class="job-detail-info-item">📅 ${formatDate(job.created_at)}</div>
        </div>
        <h3 style="font-weight:700;margin-bottom:12px;">Description</h3>
        <div class="job-detail-desc">${escapeHtml(job.description)}</div>
      </div>
      <div class="job-detail-sidebar">
        <div class="card">
          <div style="text-align:center;margin-bottom:20px;">
            <div style="font-size:2rem;font-weight:800;color:var(--success);">${formatBudget(job.budget)}</div>
            <div style="color:var(--text-secondary);font-size:0.85rem;">Project Budget</div>
          </div>
          ${isAuth && !isOwner && jobOpen ? `
            <button class="btn btn-${hasApplied ? 'secondary' : 'primary'} btn-lg" style="width:100%; margin-bottom: 10px;" id="apply-btn" ${hasApplied ? 'disabled' : ''}>
              ${hasApplied ? 'Already Applied' : 'Apply Now'}
            </button>
            <button class="btn btn-secondary btn-lg" style="width:100%" id="message-btn">Message Client</button>
          ` : ''}
          ${!isAuth ? `
            <a href="#/register" class="btn btn-primary btn-lg" style="width:100%;text-align:center;display:block;">Sign Up to Apply</a>
          ` : ''}
          ${isOwner ? '<div id="applications-section"></div>' : ''}
        </div>
      </div>
    </div>
  `;

  // Apply modal for freelancers
  const applyBtn = document.getElementById('apply-btn');
  if (applyBtn && !hasApplied) {
    applyBtn.addEventListener('click', () => {
      showModal(`
        <h2>Apply to this Job</h2>
        <form id="apply-form">
          <div class="form-group">
            <label class="form-label" for="cover-letter">Cover Letter</label>
            <textarea class="form-textarea" id="cover-letter" placeholder="Explain why you're the perfect fit for this project..." required style="min-height:150px;"></textarea>
          </div>
          <div class="modal-actions">
            <button type="button" class="btn btn-secondary" id="cancel-apply">Cancel</button>
            <button type="submit" class="btn btn-primary" id="submit-apply">Submit Application</button>
          </div>
        </form>
      `);
      document.getElementById('cancel-apply').addEventListener('click', closeModal);
      document.getElementById('apply-form').addEventListener('submit', async (e) => {
        e.preventDefault();
        const btn = document.getElementById('submit-apply');
        btn.disabled = true;
        btn.textContent = 'Submitting...';
        try {
          await api.applyToJob(jobId, document.getElementById('cover-letter').value);
          
          const applied = JSON.parse(localStorage.getItem(`applied_${user.id}`) || '[]');
          if (!applied.includes(jobId)) {
            applied.push(jobId);
            localStorage.setItem(`applied_${user.id}`, JSON.stringify(applied));
          }

          closeModal();
          showToast('Application submitted!', 'success');
          
          // Update button state immediately
          applyBtn.className = 'btn btn-secondary btn-lg';
          applyBtn.textContent = 'Already Applied';
          applyBtn.disabled = true;
        } catch (err) {
          showToast(err.message, 'error');
          btn.disabled = false;
          btn.textContent = 'Submit Application';
        }
      });
    });
  }

  // Message modal for freelancers
  document.getElementById('message-btn')?.addEventListener('click', () => {
    showModal(`
      <h2>Message Client</h2>
      <form id="message-form">
        <div class="form-group">
          <label class="form-label" for="message-content">Message</label>
          <textarea class="form-textarea" id="message-content" placeholder="Hello, I have a question about this job..." required style="min-height:100px;"></textarea>
        </div>
        <div class="modal-actions">
          <button type="button" class="btn btn-secondary" id="cancel-msg">Cancel</button>
          <button type="submit" class="btn btn-primary" id="submit-msg">Send Message</button>
        </div>
      </form>
    `);
    document.getElementById('cancel-msg').addEventListener('click', closeModal);
    document.getElementById('message-form').addEventListener('submit', async (e) => {
      e.preventDefault();
      const btn = document.getElementById('submit-msg');
      btn.disabled = true;
      btn.textContent = 'Sending...';
      try {
        await api.sendMessage(job.client_id, document.getElementById('message-content').value, jobId);
        closeModal();
        showToast('Message sent! Redirecting to chat...', 'success');
        setTimeout(() => router.navigate('/messages'), 1000);
      } catch (err) {
        showToast(err.message, 'error');
        btn.disabled = false;
        btn.textContent = 'Send Message';
      }
    });
  });

  // Load applications for job owner (this is simplified — in real app, would need a list applications endpoint)
  if (isOwner) {
    const section = document.getElementById('applications-section');
    if (section) {
      section.innerHTML = `<div style="margin-top:20px;"><h3 style="font-weight:700;margin-bottom:12px;font-size:0.95rem;">Applications</h3><p style="color:var(--text-secondary);font-size:0.85rem;">Applications for this job will appear here. Accept a freelancer to start the project.</p></div>`;
    }
  }
}

function escapeHtml(str) {
  const div = document.createElement('div');
  div.textContent = str || '';
  return div.innerHTML;
}
