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
          ${isAuth && isFreelancer && jobOpen ? `
            <button class="btn btn-primary btn-lg" style="width:100%" id="apply-btn">Apply Now</button>
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
  document.getElementById('apply-btn')?.addEventListener('click', () => {
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
        closeModal();
        showToast('Application submitted!', 'success');
      } catch (err) {
        showToast(err.message, 'error');
        btn.disabled = false;
        btn.textContent = 'Submit Application';
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
