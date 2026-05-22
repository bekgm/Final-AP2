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
    app.querySelector('.job-detail').innerHTML = `<div class="empty-state"><h3>Job not found</h3><p>${err.message}</p><a href="#/jobs" class="btn btn-primary">Back to Jobs</a></div>`;
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
          <div class="job-detail-info-item">Budget: ${formatBudget(job.budget)}</div>
          <div class="job-detail-info-item">Client: ${escapeHtml(clientName)}</div>
          <div class="job-detail-info-item">Posted: ${formatDate(job.created_at)}</div>
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
          ${isOwner && (job.status === 'JOB_STATUS_IN_PROGRESS' || job.status === 'in_progress') ? `
            <div style="margin-top: 15px; padding: 15px; border-radius: 8px; border: 1px solid var(--border); background: var(--bg-lighter);">
              <h4 style="margin-bottom:10px; font-weight:700; color:var(--text);">Manage Project</h4>
              <p style="font-size:0.85rem; color:var(--text-secondary); margin-bottom:12px;">This project is currently in progress. Once the freelancer delivers the work, you can mark it as completed.</p>
              <button class="btn btn-success" id="complete-btn" style="width:100%">Complete Project</button>
            </div>
          ` : ''}
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
          const coverLetter = document.getElementById('cover-letter').value;
          const resp = await api.applyToJob(jobId, coverLetter);
          
          const applied = JSON.parse(localStorage.getItem(`applied_${user.id}`) || '[]');
          if (!applied.includes(jobId)) {
            applied.push(jobId);
            localStorage.setItem(`applied_${user.id}`, JSON.stringify(applied));
          }

          const globalApps = JSON.parse(localStorage.getItem('global_applications') || '[]');
          globalApps.push({
            jobId,
            applicationId: resp.application?.id || resp.id || '',
            freelancerId: user.id,
            freelancerName: user.name,
            coverLetter: coverLetter
          });
          localStorage.setItem('global_applications', JSON.stringify(globalApps));

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

  // Complete project modal
  const completeBtn = document.getElementById('complete-btn');
  if (completeBtn) {
    completeBtn.addEventListener('click', () => {
      showModal(`
        <h2>Complete Project</h2>
        <p style="margin-bottom: 20px;">Are you sure you want to mark this project as <strong>completed / done</strong>? This will finalize the contract and close the project.</p>
        <form id="complete-form">
          <div class="form-group">
            <label class="form-label" for="feedback">Feedback for Freelancer (Optional)</label>
            <textarea class="form-textarea" id="feedback" placeholder="Leave some feedback about the freelancer's work..." style="min-height:100px;"></textarea>
          </div>
          <div class="modal-actions" style="margin-top:20px;">
            <button type="button" class="btn btn-secondary" id="cancel-complete">Cancel</button>
            <button type="submit" class="btn btn-success" id="confirm-complete">Mark as Completed</button>
          </div>
        </form>
      `);
      document.getElementById('cancel-complete').addEventListener('click', closeModal);
      document.getElementById('complete-form').addEventListener('submit', async (e) => {
        e.preventDefault();
        const btn = document.getElementById('confirm-complete');
        btn.disabled = true;
        btn.textContent = 'Completing...';
        try {
          await api.completeJob(jobId);
          
          // Save review to localStorage
          const feedback = document.getElementById('feedback').value.trim() || 'No feedback left.';
          const acceptedFreelancerId = localStorage.getItem(`accepted_freelancer_${jobId}`);
          if (acceptedFreelancerId) {
            const reviews = JSON.parse(localStorage.getItem('freelancer_reviews') || '[]');
            reviews.push({
              freelancerId: acceptedFreelancerId,
              clientName: user.name || user.email || 'Client',
              jobTitle: job.title,
              feedback: feedback,
              date: new Date().toISOString()
            });
            localStorage.setItem('freelancer_reviews', JSON.stringify(reviews));
          }

          closeModal();
          showToast('Project completed successfully!', 'success');
          // Reload page
          renderJobDetail(app, jobId);
        } catch (err) {
          showToast(err.message, 'error');
          btn.disabled = false;
          btn.textContent = 'Mark as Completed';
        }
      });
    });
  }

  // Load applications for job owner
  if (isOwner) {
    let jobApps = [];
    try {
      const applicationsData = await api.listApplications(jobId);
      jobApps = await Promise.all((applicationsData.applications || []).map(async a => {
        let freelancerName = a.freelancer_id;
        try {
          const freelancer = await api.getUserById(a.freelancer_id);
          freelancerName = freelancer.name || freelancer.email || a.freelancer_id;
        } catch {
        }
        return {
          applicationId: a.id,
          freelancerId: a.freelancer_id,
          freelancerName,
          coverLetter: a.cover_letter,
          status: a.status,
        };
      }));
    } catch (err) {
      showToast(`Failed to load applications: ${err.message}`, 'error');
    }
    
    window.acceptFreelancerMock = (appId, fname, freelancerId) => {
      showModal(`
        <h2>Accept Request</h2>
        <p>Are you sure you want to accept <strong>${escapeHtml(fname)}</strong> for this project?</p>
        <div class="modal-actions" style="margin-top: 20px;">
          <button class="btn btn-secondary" id="cancel-accept-btn">Cancel</button>
          <button class="btn btn-primary" id="confirm-accept-btn">Yes, Accept</button>
        </div>
      `);
      document.getElementById('cancel-accept-btn').addEventListener('click', closeModal);
      document.getElementById('confirm-accept-btn').addEventListener('click', async () => {
        const btn = document.getElementById('confirm-accept-btn');
        btn.disabled = true;
        btn.textContent = 'Accepting...';
        try {
          await api.acceptFreelancer(jobId, appId);
          localStorage.setItem(`accepted_freelancer_${jobId}`, freelancerId);
          try {
            await api.sendMessage(freelancerId, `You are accepted! Let's discuss the project details for "${escapeHtml(job.title)}".`, jobId);
          } catch(e) {
            console.error("Failed to auto-send message:", e);
          }
          closeModal();
          showToast(`Accepted ${fname}! Redirecting to messages...`, 'success');
          setTimeout(() => router.navigate('/messages'), 1500);
        } catch (err) {
          showToast(err.message, 'error');
          btn.disabled = false;
          btn.textContent = 'Yes, Accept';
        }
      });
    };

    const section = document.getElementById('applications-section');
    if (section) {
      section.innerHTML = `
        <div style="margin-top:20px;">
          <h3 style="font-weight:700;margin-bottom:12px;font-size:0.95rem;">Applications (${jobApps.length})</h3>
          ${jobApps.length === 0 ? `<p style="color:var(--text-secondary);font-size:0.85rem;">Applications for this job will appear here.</p>` : ''}
          <div style="display:flex; flex-direction:column; gap:10px;">
            ${jobApps.map(a => `
              <div class="card" style="padding:15px; background:var(--bg-lighter); border: 1px solid var(--border);">
                <div style="font-weight:600;">${escapeHtml(a.freelancerName)}</div>
                <div style="margin-top:8px; font-size:0.9rem; color:var(--text-secondary);">${escapeHtml(a.coverLetter)}</div>
                ${jobOpen ? `<button class="btn btn-sm btn-primary" style="margin-top:12px" onclick="acceptFreelancerMock('${a.applicationId}', '${escapeHtml(a.freelancerName)}', '${a.freelancerId}')">Accept Freelancer</button>` : ''}
              </div>
            `).join('')}
          </div>
        </div>
      `;
    }
  }
}

function escapeHtml(str) {
  const div = document.createElement('div');
  div.textContent = str || '';
  return div.innerHTML;
}
