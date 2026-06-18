// ═══════════════════════════════════════════════════════════════════
// WATCHME — Admin Setup Page
// First-time admin account creation with recovery key
// ═══════════════════════════════════════════════════════════════════

import { api } from '../api/client.js';
import { navigate } from '../router/router.js';
import { store } from '../state/store.js';
import { toastSuccess, toastError } from '../components/Toast.js';

export function renderSetup() {
  const container = document.createElement('div');
  container.id = 'setup-page';
  container.innerHTML = `
    <style>
      #setup-page {
        display: flex;
        align-items: center;
        justify-content: center;
        min-height: 100vh;
        padding: var(--space-xl);
      }

      .setup-card {
        width: 100%;
        max-width: 440px;
        padding: var(--space-2xl);
        border-radius: var(--radius-lg);
        animation: fadeInScale 0.6s ease both;
      }

      .setup-title {
        font-family: var(--font-mono);
        font-size: 1.6rem;
        margin-bottom: var(--space-xs);
      }

      .setup-subtitle {
        color: var(--color-text-dim);
        font-size: 0.85rem;
        margin-bottom: var(--space-2xl);
      }

      .setup-field {
        margin-bottom: var(--space-lg);
      }

      .setup-field label {
        display: block;
        font-family: var(--font-mono);
        font-size: 0.8rem;
        color: var(--color-text-dim);
        margin-bottom: var(--space-xs);
        text-transform: uppercase;
        letter-spacing: 0.05em;
      }

      .setup-btn {
        width: 100%;
        margin-top: var(--space-md);
        padding: var(--space-md);
        font-size: 0.95rem;
      }

      .recovery-card {
        display: none;
        padding: var(--space-2xl);
        border-radius: var(--radius-lg);
        width: 100%;
        max-width: 500px;
        animation: fadeInScale 0.6s ease both;
      }

      .recovery-key {
        background: hsla(0, 0%, 0%, 0.4);
        padding: var(--space-md);
        border-radius: var(--radius-md);
        font-family: var(--font-mono);
        font-size: 0.75rem;
        word-break: break-all;
        color: var(--color-primary-bright);
        border: 1px solid var(--glass-border);
        margin: var(--space-lg) 0;
        user-select: all;
        cursor: pointer;
        line-height: 1.8;
      }

      .recovery-warning {
        color: var(--color-warning);
        font-size: 0.8rem;
        margin-bottom: var(--space-lg);
        display: flex;
        align-items: flex-start;
        gap: var(--space-sm);
      }
    </style>

    <div class="setup-card glass" id="setup-form-card">
      <h1 class="setup-title glow-text">Welcome to WATCHME</h1>
      <p class="setup-subtitle">Create your admin account to get started.</p>

      <div class="setup-field">
        <label for="setup-username">Username</label>
        <input type="text" id="setup-username" class="input" placeholder="admin" autocomplete="off" />
      </div>

      <div class="setup-field">
        <label for="setup-password">Password (6+ characters)</label>
        <input type="password" id="setup-password" class="input" placeholder="••••••" />
      </div>

      <button id="setup-submit" class="btn btn-primary setup-btn">
        CREATE ADMIN ACCOUNT
      </button>
    </div>

    <div class="recovery-card glass" id="recovery-card">
      <h2 class="setup-title glow-text">🔑 Recovery Key</h2>
      <div class="recovery-warning">
        ⚠️ Save this key now. You will not see it again. It's the only way to recover your admin account.
      </div>
      <div class="recovery-key" id="recovery-key-display" title="Click to copy">
        <!-- key will be inserted here -->
      </div>
      <button id="copy-key-btn" class="btn btn-ghost" style="width:100%; margin-bottom: var(--space-md);">
        📋 COPY TO CLIPBOARD
      </button>
      <button id="continue-btn" class="btn btn-primary setup-btn">
        I'VE SAVED IT → CONTINUE
      </button>
    </div>
  `;

  setTimeout(() => {
    const submitBtn = document.getElementById('setup-submit');
    const formCard = document.getElementById('setup-form-card');
    const recoveryCard = document.getElementById('recovery-card');
    const keyDisplay = document.getElementById('recovery-key-display');
    const copyBtn = document.getElementById('copy-key-btn');
    const continueBtn = document.getElementById('continue-btn');

    submitBtn?.addEventListener('click', async () => {
      const username = document.getElementById('setup-username').value.trim();
      const password = document.getElementById('setup-password').value;

      if (!username) return toastError('Username is required');
      if (password.length < 6) return toastError('Password must be at least 6 characters');

      submitBtn.disabled = true;
      submitBtn.textContent = 'Creating...';

      try {
        const result = await api.setup(username, password);
        const recoveryKey = result.data?.recovery_key || result.data?.RecoveryKey;

        // Show recovery key
        formCard.style.display = 'none';
        recoveryCard.style.display = 'block';
        keyDisplay.textContent = recoveryKey;

        store.set('isSetupComplete', true);
        toastSuccess('Admin account created!');
      } catch (e) {
        toastError(e.message);
        submitBtn.disabled = false;
        submitBtn.textContent = 'CREATE ADMIN ACCOUNT';
      }
    });

    copyBtn?.addEventListener('click', () => {
      navigator.clipboard.writeText(keyDisplay.textContent);
      toastSuccess('Recovery key copied to clipboard!');
      copyBtn.textContent = '✅ COPIED';
    });

    continueBtn?.addEventListener('click', () => {
      navigate('/profiles');
    });
  }, 0);

  return container;
}
