import { api } from '../api.js';
import { router } from '../router.js';
import { renderNavbar, bindNavbar, showToast } from '../components.js';

export function renderHome(app) {
  const isAuth = api.isAuthenticated();
  const user = api.getUser();

  app.innerHTML = `
    ${renderNavbar()}
    <main>
      <section class="hero">
        <div class="container">
          <h1>Find The Perfect<br/><span>Freelance Talent</span></h1>
          <p>Connect with top-tier freelancers and exciting projects. Post jobs, discover opportunities, and build something amazing together.</p>
          <div class="hero-actions">
            ${isAuth
              ? (user?.role === 'ROLE_CLIENT' || user?.role === 'client'
                ? '<a href="#/jobs" class="btn btn-primary btn-lg">Post a Job</a>'
                : '<a href="#/jobs" class="btn btn-primary btn-lg">Find Work</a>')
              : `<a href="#/register" class="btn btn-primary btn-lg">Get Started Free</a>
                 <a href="#/jobs" class="btn btn-secondary btn-lg">Browse Jobs</a>`
            }
          </div>
        </div>
      </section>

      <section class="container">
        <div class="stats">
          <div class="card stat-card">
            <div class="stat-number">500+</div>
            <div class="stat-label">Active Projects</div>
          </div>
          <div class="card stat-card">
            <div class="stat-number">1,200+</div>
            <div class="stat-label">Freelancers</div>
          </div>
          <div class="card stat-card">
            <div class="stat-number">98%</div>
            <div class="stat-label">Satisfaction Rate</div>
          </div>
        </div>
      </section>

      <section class="container" style="padding-bottom: 80px;">
        <h2 style="font-size:1.6rem;font-weight:800;margin-bottom:24px;text-align:center;">How It Works</h2>
        <div class="stats">
          <div class="card" style="text-align:center;padding:32px;">
            
            <h3 style="font-weight:700;margin-bottom:8px;">Post a Job</h3>
            <p style="color:var(--text-secondary);font-size:0.9rem;">Describe your project, set a budget, and publish it for freelancers to see.</p>
          </div>
          <div class="card" style="text-align:center;padding:32px;">
            
            <h3 style="font-weight:700;margin-bottom:8px;">Get Proposals</h3>
            <p style="color:var(--text-secondary);font-size:0.9rem;">Receive applications from skilled freelancers with cover letters.</p>
          </div>
          <div class="card" style="text-align:center;padding:32px;">
            
            <h3 style="font-weight:700;margin-bottom:8px;">Collaborate</h3>
            <p style="color:var(--text-secondary);font-size:0.9rem;">Accept a freelancer and start working together via built-in messaging.</p>
          </div>
        </div>
      </section>
    </main>
  `;

  bindNavbar();
}
