import { api } from '../api.js';
import { router } from '../router.js';
import { renderNavbar, bindNavbar, showToast, showModal, closeModal, formatBudget, statusBadge, formatDate } from '../components.js';

export function renderJobs(app) {
  const user = api.getUser();
  const isClient = user?.role === 'ROLE_CLIENT' || user?.role === 'client';

  app.innerHTML = `
    ${renderNavbar()}
    <div class="container">
      <div class="page-header">
        <h1>Browse Jobs</h1>
        <p>Discover exciting opportunities and find the perfect match</p>
        <div class="page-header-actions">
          <div style="display:flex;gap:8px;flex:1;max-width:400px;">
            <input class="form-input" id="search-input" placeholder="Search jobs..." style="flex:1" />
          </div>
          ${api.isAuthenticated() ? '<button class="btn btn-primary" id="create-job-btn">+ Post a Job</button>' : ''}
        </div>
      </div>
      <div id="jobs-container" class="jobs-grid">
        <div class="loading-center"><div class="spinner"></div></div>
      </div>
      <div id="pagination" style="display:flex;justify-content:center;gap:8px;padding:40px 0;"></div>
    </div>
  `;

  bindNavbar();

  let currentPage = 1;
  const pageSize = 12;
  let allJobs = [];

  async function loadJobs(page) {
    const container = document.getElementById('jobs-container');
    container.innerHTML = '<div class="loading-center"><div class="spinner"></div></div>';
    try {
      const data = await api.listJobs(page, pageSize);
      allJobs = data.jobs || [];
      currentPage = page;
      renderJobCards(allJobs);
      renderPagination(data.total || 0);
    } catch (err) {
      container.innerHTML = `<div class="empty-state"><div class="empty-state-icon">⚠️</div><h3>Failed to load jobs</h3><p>${err.message}</p></div>`;
    }
  }

  function renderJobCards(jobs) {
    const container = document.getElementById('jobs-container');
    if (!jobs || jobs.length === 0) {
      container.innerHTML = `<div class="empty-state" style="grid-column:1/-1"><div class="empty-state-icon">📋</div><h3>No jobs found</h3><p>Be the first to post a job opportunity!</p>${api.isAuthenticated() ? '<button class="btn btn-primary" onclick="document.getElementById(\'create-job-btn\')?.click()">Post a Job</button>' : ''}</div>`;
      return;
    }
    container.innerHTML = jobs.map(job => `
      <div class="card job-card" data-job-id="${job.id}">
        <div class="job-card-header">
          <div class="job-card-title">${escapeHtml(job.title)}</div>
          <div class="job-card-budget">${formatBudget(job.budget)}</div>
        </div>
        <div class="job-card-desc">${escapeHtml(job.description)}</div>
        <div class="job-card-footer">
          ${statusBadge(job.status)}
          <span class="job-card-meta">${formatDate(job.created_at)}</span>
        </div>
      </div>
    `).join('');

    container.querySelectorAll('.job-card').forEach(card => {
      card.addEventListener('click', () => router.navigate(`/jobs/${card.dataset.jobId}`));
    });
  }

  function renderPagination(total) {
    const pages = Math.ceil(total / pageSize);
    const el = document.getElementById('pagination');
    if (pages <= 1) { el.innerHTML = ''; return; }
    let html = '';
    for (let i = 1; i <= pages; i++) {
      html += `<button class="btn ${i === currentPage ? 'btn-primary' : 'btn-secondary'} btn-sm" data-page="${i}">${i}</button>`;
    }
    el.innerHTML = html;
    el.querySelectorAll('button').forEach(b => b.addEventListener('click', () => loadJobs(parseInt(b.dataset.page))));
  }

  // Search filter
  document.getElementById('search-input')?.addEventListener('input', (e) => {
    const q = e.target.value.toLowerCase();
    const filtered = allJobs.filter(j => j.title.toLowerCase().includes(q) || j.description.toLowerCase().includes(q));
    renderJobCards(filtered);
  });

  // Create Job Modal
  document.getElementById('create-job-btn')?.addEventListener('click', () => {
    const overlay = showModal(`
      <h2>Post a New Job</h2>
      <form id="create-job-form">
        <div class="form-group">
          <label class="form-label" for="job-title">Job Title</label>
          <input class="form-input" id="job-title" placeholder="e.g. Go Developer Needed" required />
        </div>
        <div class="form-group">
          <label class="form-label" for="job-desc">Description</label>
          <textarea class="form-textarea" id="job-desc" placeholder="Describe the project, requirements, and expectations..." required></textarea>
        </div>
        <div class="form-group">
          <label class="form-label" for="job-budget">Budget ($)</label>
          <input class="form-input" id="job-budget" type="number" min="1" step="1" placeholder="1500" required />
        </div>
        <div class="modal-actions">
          <button type="button" class="btn btn-secondary" id="cancel-modal">Cancel</button>
          <button type="submit" class="btn btn-primary" id="submit-job">Post Job</button>
        </div>
      </form>
    `);

    document.getElementById('cancel-modal').addEventListener('click', closeModal);

    document.getElementById('create-job-form').addEventListener('submit', async (e) => {
      e.preventDefault();
      const btn = document.getElementById('submit-job');
      btn.disabled = true;
      btn.textContent = 'Posting...';
      try {
        await api.createJob(
          document.getElementById('job-title').value,
          document.getElementById('job-desc').value,
          document.getElementById('job-budget').value
        );
        closeModal();
        showToast('Job posted successfully!', 'success');
        loadJobs(1);
      } catch (err) {
        showToast(err.message, 'error');
        btn.disabled = false;
        btn.textContent = 'Post Job';
      }
    });
  });

  loadJobs(1);
}

function escapeHtml(str) {
  const div = document.createElement('div');
  div.textContent = str || '';
  return div.innerHTML;
}
