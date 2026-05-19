const API_BASE = import.meta.env.VITE_API_URL || '';
const MSG_BASE = import.meta.env.VITE_MSG_URL || '';

class ApiClient {
  constructor() {
    this.token = localStorage.getItem('token') || '';
    this.userId = localStorage.getItem('userId') || '';
  }

  setAuth(token, userId) {
    this.token = token;
    this.userId = userId;
    localStorage.setItem('token', token);
    localStorage.setItem('userId', userId);
  }

  clearAuth() {
    this.token = '';
    this.userId = '';
    localStorage.removeItem('token');
    localStorage.removeItem('userId');
    localStorage.removeItem('user');
  }

  isAuthenticated() {
    return !!this.token && !!this.userId;
  }

  getUser() {
    const raw = localStorage.getItem('user');
    return raw ? JSON.parse(raw) : null;
  }

  setUser(user) {
    localStorage.setItem('user', JSON.stringify(user));
  }

  async request(url, options = {}) {
    const headers = { 'Content-Type': 'application/json', ...options.headers };
    if (this.token) headers['Authorization'] = `Bearer ${this.token}`;
    if (this.userId) headers['X-User-ID'] = this.userId;

    const resp = await fetch(url, { ...options, headers });
    const data = await resp.json().catch(() => null);

    if (!resp.ok) {
      const msg = data?.error || `Request failed (${resp.status})`;
      throw new Error(msg);
    }
    return data;
  }

  // Auth
  async register(email, password, name, role) {
    const data = await this.request(`${API_BASE}/users/register`, {
      method: 'POST',
      body: JSON.stringify({ email, password, name, role }),
    });
    const user = data.user || {};
    this.setAuth(data.token, user.id);
    this.setUser(user);
    return data;
  }

  async login(email, password) {
    const data = await this.request(`${API_BASE}/users/login`, {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    });
    const user = data.user || {};
    this.setAuth(data.token, user.id);
    this.setUser(user);
    return data;
  }

  // Users
  async getUserById(userId) {
    return this.request(`${API_BASE}/users/${userId}`);
  }

  async updateUser(userId, fields) {
    return this.request(`${API_BASE}/users/${userId}`, {
      method: 'PATCH',
      body: JSON.stringify(fields),
    });
  }

  // Jobs
  async createJob(title, description, budget) {
    return this.request(`${API_BASE}/jobs`, {
      method: 'POST',
      body: JSON.stringify({ client_id: this.userId, title, description, budget: parseFloat(budget) }),
    });
  }

  async getJob(jobId) {
    return this.request(`${API_BASE}/jobs/${jobId}`);
  }

  async listJobs(page = 1, pageSize = 20, clientId = '') {
    let url = `${API_BASE}/jobs?page=${page}&page_size=${pageSize}`;
    if (clientId) url += `&client_id=${clientId}`;
    return this.request(url);
  }

  async applyToJob(jobId, coverLetter) {
    return this.request(`${API_BASE}/jobs/${jobId}/apply`, {
      method: 'POST',
      body: JSON.stringify({ freelancer_id: this.userId, cover_letter: coverLetter }),
    });
  }

  async acceptFreelancer(jobId, applicationId) {
    return this.request(`${API_BASE}/jobs/${jobId}/accept`, {
      method: 'POST',
      body: JSON.stringify({ application_id: applicationId }),
    });
  }

  async completeJob(jobId) {
    return this.request(`${API_BASE}/jobs/${jobId}/complete`, {
      method: 'POST',
    });
  }

  // Messaging
  async sendMessage(receiverId, content, projectId = '') {
    return this.request(`${MSG_BASE}/api/messages`, {
      method: 'POST',
      body: JSON.stringify({ receiver_id: receiverId, content, project_id: projectId }),
    });
  }

  async getMessages(otherUserId, projectId = '', limit = 50, offset = 0) {
    let url = `${MSG_BASE}/api/messages?user_id=${otherUserId}&limit=${limit}&offset=${offset}`;
    if (projectId) url += `&project_id=${projectId}`;
    return this.request(url);
  }

  async getDialogs() {
    return this.request(`${MSG_BASE}/api/dialogs`);
  }
}

export const api = new ApiClient();
