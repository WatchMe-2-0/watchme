// ═══════════════════════════════════════════════════════════════════
// WATCHME — Settings Page
// ═══════════════════════════════════════════════════════════════════

import { api } from '../api/client.js';
import { store } from '../state/store.js';
import { navigate } from '../router/router.js';
import { toastSuccess, toastError } from '../components/Toast.js';

export async function renderSettings() {
  const container = document.createElement('div');
  container.id = 'settings-page';

  let settings = {};
  try {
    const result = await api.getSettings();
    settings = result.data || {};
  } catch (e) {}

  const currentTheme = store.get('theme');

  container.innerHTML = `
    <style>
      #settings-page {
        margin-left: var(--sidebar-width);
        padding: var(--space-xl) var(--space-2xl);
        min-height: 100vh;
        max-width: 700px;
      }

      .settings-title {
        font-family: var(--font-mono);
        font-size: 1.4rem;
        margin-bottom: var(--space-2xl);
      }

      .settings-section {
        margin-bottom: var(--space-2xl);
        padding: var(--space-xl);
        border-radius: var(--radius-lg);
      }

      .settings-section-title {
        font-family: var(--font-mono);
        font-size: 0.9rem;
        color: var(--color-primary);
        margin-bottom: var(--space-lg);
        text-transform: uppercase;
        letter-spacing: 0.08em;
      }

      .settings-field {
        margin-bottom: var(--space-lg);
      }

      .settings-field label {
        display: block;
        font-family: var(--font-mono);
        font-size: 0.75rem;
        color: var(--color-text-dim);
        margin-bottom: var(--space-xs);
        text-transform: uppercase;
        letter-spacing: 0.05em;
      }

      .theme-selector {
        display: flex;
        gap: var(--space-md);
      }

      .theme-option {
        flex: 1;
        padding: var(--space-md);
        border-radius: var(--radius-md);
        border: 2px solid var(--glass-border);
        background: var(--glass-bg);
        cursor: pointer;
        text-align: center;
        font-family: var(--font-mono);
        font-size: 0.8rem;
        transition: all var(--transition-normal);
      }

      .theme-option:hover {
        border-color: var(--glass-border-hover);
      }

      .theme-option.active {
        border-color: var(--color-primary);
        box-shadow: 0 0 15px var(--color-primary-glow);
      }

      .theme-preview {
        width: 100%;
        height: 40px;
        border-radius: var(--radius-sm);
        margin-bottom: var(--space-sm);
      }

      .settings-actions {
        display: flex;
        gap: var(--space-md);
        margin-top: var(--space-lg);
      }

      .settings-back {
        margin-bottom: var(--space-xl);
      }
    </style>

    <button class="btn btn-ghost settings-back" id="settings-back">← Dashboard</button>

    <h1 class="settings-title glow-text">⚙️ Settings</h1>

    <!-- Theme -->
    <div class="settings-section glass">
      <h3 class="settings-section-title">Theme</h3>
      <div class="theme-selector">
        <div class="theme-option ${currentTheme === 'default' ? 'active' : ''}" data-theme="default">
          <div class="theme-preview" style="background:linear-gradient(135deg, #050505, #003050, #00bfff20)"></div>
          Default
        </div>
        <div class="theme-option ${currentTheme === 'night' ? 'active' : ''}" data-theme="night">
          <div class="theme-preview" style="background:linear-gradient(135deg, #0a0a0f, #1a1a3a, #4a9eff20)"></div>
          Night Mode
        </div>
        <div class="theme-option ${currentTheme === 'dracula' ? 'active' : ''}" data-theme="dracula">
          <div class="theme-preview" style="background:linear-gradient(135deg, #0d0b1a, #2a1040, #bd93f920)"></div>
          Dracula
        </div>
      </div>
    </div>

    <!-- Storage -->
    <div class="settings-section glass">
      <h3 class="settings-section-title">Storage</h3>
      <div class="settings-field">
        <label>Download Directory</label>
        <input type="text" class="input" id="setting-download-dir" value="${settings.download_dir || ''}" />
      </div>
      <div class="settings-field">
        <label>Poster Directory</label>
        <input type="text" class="input" id="setting-poster-dir" value="${settings.poster_dir || ''}" />
      </div>
    </div>

    <!-- API -->
    <div class="settings-section glass">
      <h3 class="settings-section-title">TMDB API</h3>
      <div class="settings-field">
        <label>API Key</label>
        <input type="text" class="input" id="setting-tmdb-key" value="${settings.tmdb_api_key || ''}" placeholder="Enter your TMDB API key" />
      </div>
    </div>

    <!-- Downloads -->
    <div class="settings-section glass">
      <h3 class="settings-section-title">Downloads</h3>
      <div class="settings-field">
        <label>Max Concurrent Downloads</label>
        <input type="number" class="input" id="setting-max-dl" value="${settings.max_concurrent_downloads || 3}" min="1" max="10" />
      </div>
    </div>

    <div class="settings-actions">
      <button class="btn btn-primary" id="save-settings">💾 Save Settings</button>
      <button class="btn btn-ghost" id="logout-btn">🚪 Switch Profile</button>
    </div>
  `;

  setTimeout(() => {
    // Theme switching
    container.querySelectorAll('[data-theme]').forEach(opt => {
      opt.addEventListener('click', () => {
        container.querySelectorAll('.theme-option').forEach(o => o.classList.remove('active'));
        opt.classList.add('active');
        store.set('theme', opt.dataset.theme);
        toastSuccess(`Theme changed to ${opt.dataset.theme}`);
      });
    });

    // Save settings
    document.getElementById('save-settings')?.addEventListener('click', async () => {
      try {
        await api.updateSettings({
          download_dir: document.getElementById('setting-download-dir')?.value,
          poster_dir: document.getElementById('setting-poster-dir')?.value,
          tmdb_api_key: document.getElementById('setting-tmdb-key')?.value,
          max_concurrent_downloads: parseInt(document.getElementById('setting-max-dl')?.value) || 3,
        });
        toastSuccess('Settings saved!');
      } catch (e) {
        toastError(e.message);
      }
    });

    // Logout
    document.getElementById('logout-btn')?.addEventListener('click', async () => {
      try {
        await api.logout();
      } catch (e) {}
      api.clearToken();
      store.logout();
      navigate('/profiles');
    });

    // Back
    document.getElementById('settings-back')?.addEventListener('click', () => navigate('/dashboard'));
  }, 0);

  return container;
}
