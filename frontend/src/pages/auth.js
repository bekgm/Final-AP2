import { api } from '../api.js';
import { router } from '../router.js';
import { renderNavbar, bindNavbar, showToast } from '../components.js';

export function renderLogin(app) {
  app.innerHTML = `
    ${renderNavbar()}
    <div class="auth-page">
      <div class="auth-card card-glass">
        <h1>Welcome Back</h1>
        <p class="subtitle">Sign in to your FreelanceHub account</p>
        <form id="login-form">
          <div class="form-group">
            <label class="form-label" for="login-email">Email</label>
            <input class="form-input" id="login-email" type="email" placeholder="you@example.com" required />
          </div>
          <div class="form-group">
            <label class="form-label" for="login-password">Password</label>
            <input class="form-input" id="login-password" type="password" placeholder="••••••••" required />
          </div>
          <button type="submit" class="btn btn-primary btn-lg" style="width:100%" id="login-submit">Sign In</button>
        </form>
        <p class="auth-footer">Don't have an account? <a href="#/register">Create one</a></p>
      </div>
    </div>
  `;

  bindNavbar();

  document.getElementById('login-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    const btn = document.getElementById('login-submit');
    btn.disabled = true;
    btn.textContent = 'Signing in...';
    try {
      await api.login(
        document.getElementById('login-email').value,
        document.getElementById('login-password').value
      );
      showToast('Welcome back!', 'success');
      router.navigate('/jobs');
    } catch (err) {
      showToast(err.message, 'error');
    } finally {
      btn.disabled = false;
      btn.textContent = 'Sign In';
    }
  });
}

export function renderRegister(app) {
  let selectedRole = 'client';

  app.innerHTML = `
    ${renderNavbar()}
    <div class="auth-page">
      <div class="auth-card card-glass">
        <h1>Join FreelanceHub</h1>
        <p class="subtitle">Create your account and start your journey</p>
        <form id="register-form">
          <div class="form-group">
            <label class="form-label">I want to</label>
            <div class="role-selector">
              <div class="role-option selected" data-role="client" id="role-client">
                <div class="role-name">Hire Talent</div>
              </div>
              <div class="role-option" data-role="freelancer" id="role-freelancer">
                <div class="role-name">Find Work</div>
              </div>
            </div>
          </div>
          <div class="form-group">
            <label class="form-label" for="reg-name">Full Name</label>
            <input class="form-input" id="reg-name" type="text" placeholder="John Doe" required />
          </div>
          <div class="form-group">
            <label class="form-label" for="reg-email">Email</label>
            <input class="form-input" id="reg-email" type="email" placeholder="you@example.com" required />
          </div>
          <div class="form-group">
            <label class="form-label" for="reg-password">Password</label>
            <input class="form-input" id="reg-password" type="password" placeholder="Min 6 characters" required minlength="6" />
          </div>
          <button type="submit" class="btn btn-primary btn-lg" style="width:100%" id="reg-submit">Create Account</button>
        </form>
        <p class="auth-footer">Already have an account? <a href="#/login">Sign in</a></p>
      </div>
    </div>
  `;

  bindNavbar();

  document.querySelectorAll('.role-option').forEach(opt => {
    opt.addEventListener('click', () => {
      document.querySelectorAll('.role-option').forEach(o => o.classList.remove('selected'));
      opt.classList.add('selected');
      selectedRole = opt.dataset.role;
    });
  });

  document.getElementById('register-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    const btn = document.getElementById('reg-submit');
    btn.disabled = true;
    btn.textContent = 'Creating account...';
    try {
      await api.register(
        document.getElementById('reg-email').value,
        document.getElementById('reg-password').value,
        document.getElementById('reg-name').value,
        selectedRole
      );
      showToast('Account created successfully!', 'success');
      router.navigate('/jobs');
    } catch (err) {
      showToast(err.message, 'error');
    } finally {
      btn.disabled = false;
      btn.textContent = 'Create Account';
    }
  });
}
