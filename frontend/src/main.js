import './style.css';
import { router } from './router.js';
import { initToast, bindNavbar } from './components.js';
import { renderHome } from './pages/home.js';
import { renderLogin, renderRegister } from './pages/auth.js';
import { renderJobs } from './pages/jobs.js';
import { renderJobDetail } from './pages/jobDetail.js';
import { renderMessages } from './pages/messages.js';
import { renderProfile } from './pages/profile.js';

const app = document.getElementById('app');

initToast();

// Start global real-time navbar notification updates
setInterval(() => {
  bindNavbar();
}, 3000);

router
  .on('/', () => renderHome(app))
  .on('/login', () => renderLogin(app))
  .on('/register', () => renderRegister(app))
  .on('/jobs', () => renderJobs(app))
  .on('/jobs/:id', (params) => renderJobDetail(app, params.id))
  .on('/messages', () => renderMessages(app))
  .on('/profile', () => renderProfile(app))
  .on('/profile/:id', (params) => renderProfile(app, params.id))
  .start();
