// Simple hash-based router
class Router {
  constructor() {
    this.routes = {};
    this.currentRoute = null;
    window.addEventListener('hashchange', () => this.resolve());
  }

  on(path, handler) {
    this.routes[path] = handler;
    return this;
  }

  navigate(path) {
    window.location.hash = path;
  }

  resolve() {
    const hash = window.location.hash.slice(1) || '/';
    const parts = hash.split('/').filter(Boolean);

    // Try exact match first
    if (this.routes[hash]) {
      this.currentRoute = hash;
      this.routes[hash]({});
      return;
    }

    // Try pattern matching
    for (const [pattern, handler] of Object.entries(this.routes)) {
      const patternParts = pattern.split('/').filter(Boolean);
      if (patternParts.length !== parts.length) continue;

      const params = {};
      let match = true;
      for (let i = 0; i < patternParts.length; i++) {
        if (patternParts[i].startsWith(':')) {
          params[patternParts[i].slice(1)] = parts[i];
        } else if (patternParts[i] !== parts[i]) {
          match = false;
          break;
        }
      }
      if (match) {
        this.currentRoute = pattern;
        handler(params);
        return;
      }
    }

    // Default: redirect to home
    this.navigate('/');
  }

  start() {
    this.resolve();
  }
}

export const router = new Router();
