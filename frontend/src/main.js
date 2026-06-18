// ═══════════════════════════════════════════════════════════════════
// WATCHME — Application Entry Point
// ═══════════════════════════════════════════════════════════════════

import './styles/index.css';
import './styles/themes.css';
import './styles/responsive.css';

import { initAurora, setAuroraTheme } from './aurora/aurora.js';
import { registerRoute, initRouter, navigate } from './router/router.js';
import { store } from './state/store.js';
import { api } from './api/client.js';

// Import pages
import { renderLanding } from './pages/Landing.js';
import { renderSetup } from './pages/Setup.js';
import { renderProfileSelect } from './pages/ProfileSelect.js';
import { renderPinEntry } from './pages/PinEntry.js';
import { renderDashboard } from './pages/Dashboard.js';
import { renderBrowse } from './pages/Browse.js';
import { renderMovieDetail } from './pages/MovieDetail.js';
import { renderDownloads } from './pages/Downloads.js';
import { renderSettings } from './pages/Settings.js';

async function init() {
  // Apply saved theme
  const theme = store.get('theme');
  document.documentElement.setAttribute('data-theme', theme);

  // Initialize aurora background
  initAurora();
  setAuroraTheme(theme);

  // Listen for theme changes
  store.on('theme', (newTheme) => {
    document.documentElement.setAttribute('data-theme', newTheme);
    setAuroraTheme(newTheme);
  });

  // Register routes
  registerRoute('/', renderLanding);
  registerRoute('/setup', renderSetup);
  registerRoute('/profiles', renderProfileSelect);
  registerRoute('/pin/:id', renderPinEntry);
  registerRoute('/dashboard', renderDashboard);
  registerRoute('/browse', renderBrowse);
  registerRoute('/movie/:id', renderMovieDetail);
  registerRoute('/downloads', renderDownloads);
  registerRoute('/settings', renderSettings);

  // Check auth status and redirect accordingly
  try {
    const status = await api.getAuthStatus();
    store.set('isSetupComplete', status.data?.setup_complete || false);

    if (!store.get('isSetupComplete')) {
      navigate('/setup');
    } else {
      // Check if user has a valid session
      const token = localStorage.getItem('watchme_token');
      if (token) {
        api.setToken(token);
        store.set('isAuthenticated', true);
        // If on landing or root, go to dashboard
        const currentHash = window.location.hash.slice(1);
        if (!currentHash || currentHash === '/') {
          navigate('/profiles');
        }
      }
    }
  } catch (e) {
    console.warn('Backend not available:', e.message);
  }

  // Initialize router
  initRouter();
}

// Boot
init();
