// ═══════════════════════════════════════════════════════════════════
// WATCHME — API Client
// Fetch wrapper with auth, error handling, and base URL config
// ═══════════════════════════════════════════════════════════════════

const API_BASE = 'http://localhost:8000/api';

class ApiClient {
  constructor() {
    this.token = localStorage.getItem('watchme_token') || '';
  }

  setToken(token) {
    this.token = token;
    localStorage.setItem('watchme_token', token);
  }

  clearToken() {
    this.token = '';
    localStorage.removeItem('watchme_token');
  }

  async request(method, path, body = null, options = {}) {
    const url = `${API_BASE}${path}`;

    const headers = {
      ...options.headers,
    };

    // Add auth header if we have a token
    if (this.token) {
      headers['Authorization'] = `Bearer ${this.token}`;
    }

    // Only set Content-Type for JSON bodies
    if (body && !(body instanceof FormData)) {
      headers['Content-Type'] = 'application/json';
    }

    const config = {
      method,
      headers,
      credentials: 'include', // Include cookies
    };

    if (body) {
      config.body = body instanceof FormData ? body : JSON.stringify(body);
    }

    try {
      const response = await fetch(url, config);
      const data = await response.json();

      if (!response.ok) {
        throw new ApiError(
          data.error || `Request failed (${response.status})`,
          response.status,
          data
        );
      }

      return data;
    } catch (error) {
      if (error instanceof ApiError) throw error;
      throw new ApiError(`Network error: ${error.message}`, 0);
    }
  }

  // Convenience methods
  get(path) { return this.request('GET', path); }
  post(path, body) { return this.request('POST', path, body); }
  put(path, body) { return this.request('PUT', path, body); }
  delete(path) { return this.request('DELETE', path); }

  // ── Auth ────────────────────────────────────────────────────────
  getAuthStatus() { return this.get('/auth/status'); }
  setup(username, password) { return this.post('/auth/setup', { username, password }); }
  login(profileId, pin) { return this.post('/auth/login', { profile_id: profileId, pin }); }
  logout() { return this.post('/auth/logout'); }
  recover(recoveryKey, newPassword) {
    return this.post('/auth/recovery', { recovery_key: recoveryKey, new_password: newPassword });
  }

  // ── Profiles ────────────────────────────────────────────────────
  getProfiles() { return this.get('/profiles/'); }
  getAvatars() { return this.get('/profiles/avatars'); }
  createProfile(data) { return this.post('/profiles/', data); }
  updateProfile(id, data) { return this.put(`/profiles/${id}`, data); }
  deleteProfile(id) { return this.delete(`/profiles/${id}`); }

  // ── Movies ──────────────────────────────────────────────────────
  getMovies(page = 1, search = '') {
    let path = `/movies?page=${page}`;
    if (search) path += `&search=${encodeURIComponent(search)}`;
    return this.get(path);
  }
  getMovie(id) { return this.get(`/movies/${id}`); }
  deleteMovie(id) { return this.delete(`/movies/${id}`); }
  getStreamUrl(id) { return `${API_BASE}/stream/${id}`; }

  uploadMovie(formData) {
    return this.request('POST', '/upload', formData);
  }

  // ── Downloads ───────────────────────────────────────────────────
  startDownload(magnetLink, title) {
    return this.post('/downloads', { magnet_link: magnetLink, title });
  }
  getDownloads() { return this.get('/downloads'); }
  cancelDownload(id) { return this.delete(`/downloads/${id}`); }

  // SSE for download progress
  subscribeProgress(onMessage) {
    const url = `${API_BASE}/downloads/progress?token=${encodeURIComponent(this.token)}`;
    const eventSource = new EventSource(url);

    eventSource.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        onMessage(data);
      } catch (e) {
        // Heartbeat or invalid JSON
      }
    };

    eventSource.onerror = () => {
      // Reconnect handled by EventSource automatically
    };

    return eventSource; // Return so caller can close it
  }

  // ── TMDB ────────────────────────────────────────────────────────
  tmdbSearch(query, page = 1) { return this.get(`/tmdb/search?q=${encodeURIComponent(query)}&page=${page}`); }
  tmdbTrending(window = 'week', page = 1) { return this.get(`/tmdb/trending?window=${window}&page=${page}`); }
  tmdbTopRated(page = 1) { return this.get(`/tmdb/top-rated?page=${page}`); }
  tmdbByGenre(genreId, page = 1) { return this.get(`/tmdb/genre/${genreId}?page=${page}`); }
  tmdbGenres() { return this.get('/tmdb/genres'); }
  tmdbMovie(id) { return this.get(`/tmdb/movie/${id}`); }

  // ── Settings ────────────────────────────────────────────────────
  getSettings() { return this.get('/settings/'); }
  updateSettings(data) { return this.put('/settings/', data); }
}

class ApiError extends Error {
  constructor(message, status, data = null) {
    super(message);
    this.status = status;
    this.data = data;
    this.name = 'ApiError';
  }
}

export const api = new ApiClient();
export { ApiError };
