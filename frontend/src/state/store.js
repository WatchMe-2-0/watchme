// ═══════════════════════════════════════════════════════════════════
// WATCHME — Simple Reactive State Store
// Pub/sub state management with localStorage persistence
// ═══════════════════════════════════════════════════════════════════

class Store {
  constructor() {
    this.state = {
      // Auth
      isSetupComplete: false,
      currentProfile: null,
      isAuthenticated: false,

      // Theme
      theme: localStorage.getItem('watchme_theme') || 'default',

      // Movies
      movies: [],
      moviesTotal: 0,

      // Downloads
      downloads: [],
      activeDownloads: [],

      // UI
      sidebarExpanded: false,
      currentPage: '/',
    };

    this.listeners = {};
  }

  // Get a state value
  get(key) {
    return this.state[key];
  }

  // Set a state value and notify listeners
  set(key, value) {
    const oldValue = this.state[key];
    this.state[key] = value;

    // Persist theme
    if (key === 'theme') {
      localStorage.setItem('watchme_theme', value);
    }

    // Notify listeners
    if (this.listeners[key]) {
      this.listeners[key].forEach(fn => fn(value, oldValue));
    }

    // Also notify wildcard listeners
    if (this.listeners['*']) {
      this.listeners['*'].forEach(fn => fn(key, value, oldValue));
    }
  }

  // Subscribe to changes on a key
  on(key, callback) {
    if (!this.listeners[key]) {
      this.listeners[key] = [];
    }
    this.listeners[key].push(callback);

    // Return unsubscribe function
    return () => {
      this.listeners[key] = this.listeners[key].filter(fn => fn !== callback);
    };
  }

  // Update multiple values at once
  update(updates) {
    Object.entries(updates).forEach(([key, value]) => {
      this.set(key, value);
    });
  }

  // Login a profile
  login(profile, token) {
    this.update({
      currentProfile: profile,
      isAuthenticated: true,
    });
  }

  // Logout
  logout() {
    this.update({
      currentProfile: null,
      isAuthenticated: false,
    });
  }
}

export const store = new Store();
