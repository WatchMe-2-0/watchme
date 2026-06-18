// ═══════════════════════════════════════════════════════════════════
// WATCHME — PIN Entry Page
// ═══════════════════════════════════════════════════════════════════

import { api } from '../api/client.js';
import { navigate } from '../router/router.js';
import { store } from '../state/store.js';
import { toastError } from '../components/Toast.js';

export function renderPinEntry(params) {
  const profileId = parseInt(params.id);
  const container = document.createElement('div');
  container.id = 'pin-page';

  container.innerHTML = `
    <style>
      #pin-page {
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        min-height: 100vh;
      }

      .pin-card {
        padding: var(--space-2xl) var(--space-3xl);
        border-radius: var(--radius-lg);
        text-align: center;
        animation: fadeInScale 0.5s ease both;
      }

      .pin-title {
        font-family: var(--font-mono);
        font-size: 1.2rem;
        margin-bottom: var(--space-2xl);
      }

      .pin-dots {
        display: flex;
        gap: var(--space-md);
        justify-content: center;
        margin-bottom: var(--space-2xl);
      }

      .pin-dot {
        width: 48px;
        height: 56px;
        border-radius: var(--radius-md);
        border: 2px solid var(--glass-border);
        background: var(--glass-bg);
        display: flex;
        align-items: center;
        justify-content: center;
        font-family: var(--font-mono);
        font-size: 1.5rem;
        color: var(--color-primary);
        transition: all var(--transition-fast);
      }

      .pin-dot.filled {
        border-color: var(--color-primary);
        box-shadow: 0 0 15px var(--color-primary-glow);
      }

      .pin-dot.error {
        border-color: var(--color-danger);
        box-shadow: 0 0 15px var(--color-danger-glow);
      }

      .pin-keypad {
        display: grid;
        grid-template-columns: repeat(3, 1fr);
        gap: var(--space-sm);
        max-width: 240px;
        margin: 0 auto;
      }

      .pin-key {
        width: 64px;
        height: 64px;
        border-radius: var(--radius-md);
        border: 1px solid var(--glass-border);
        background: var(--glass-bg);
        color: var(--color-text);
        font-family: var(--font-mono);
        font-size: 1.3rem;
        cursor: pointer;
        transition: all var(--transition-fast);
        backdrop-filter: blur(10px);
      }

      .pin-key:hover {
        background: var(--color-surface-hover);
        border-color: var(--glass-border-hover);
        transform: scale(1.05);
      }

      .pin-key:active {
        transform: scale(0.95);
      }

      .pin-back-link {
        margin-top: var(--space-xl);
        color: var(--color-text-dim);
        font-size: 0.85rem;
        cursor: pointer;
        transition: color var(--transition-fast);
        background: none;
        border: none;
        font-family: var(--font-mono);
      }

      .pin-back-link:hover {
        color: var(--color-primary);
      }
    </style>

    <div class="pin-card glass" id="pin-card">
      <h2 class="pin-title glow-text">Enter PIN</h2>

      <div class="pin-dots" id="pin-dots">
        <div class="pin-dot" id="dot-0"></div>
        <div class="pin-dot" id="dot-1"></div>
        <div class="pin-dot" id="dot-2"></div>
        <div class="pin-dot" id="dot-3"></div>
      </div>

      <div class="pin-keypad">
        ${[1,2,3,4,5,6,7,8,9,'',0,'←'].map(k =>
          k === '' ? '<div></div>' :
          `<button class="pin-key" data-key="${k}">${k}</button>`
        ).join('')}
      </div>

      <button class="pin-back-link" id="pin-back">← Back to profiles</button>
    </div>
  `;

  setTimeout(() => {
    let pin = '';
    const maxLen = 4;

    function updateDots() {
      for (let i = 0; i < maxLen; i++) {
        const dot = document.getElementById(`dot-${i}`);
        if (dot) {
          dot.textContent = i < pin.length ? '•' : '';
          dot.className = `pin-dot ${i < pin.length ? 'filled' : ''}`;
        }
      }
    }

    async function submitPin() {
      try {
        const result = await api.login(profileId, pin);
        const token = result.data?.token;
        const profile = result.data?.profile;

        if (token) {
          api.setToken(token);
          store.login(profile, token);
          navigate('/dashboard');
        }
      } catch (e) {
        // Shake animation on error
        const card = document.getElementById('pin-card');
        card?.classList.add('animate-shake');
        setTimeout(() => card?.classList.remove('animate-shake'), 500);

        // Red dots
        for (let i = 0; i < maxLen; i++) {
          document.getElementById(`dot-${i}`)?.classList.add('error');
        }

        pin = '';
        setTimeout(updateDots, 600);
        toastError('Invalid PIN');
      }
    }

    // Keypad clicks
    container.querySelectorAll('.pin-key').forEach(key => {
      key.addEventListener('click', () => {
        const val = key.dataset.key;
        if (val === '←') {
          pin = pin.slice(0, -1);
        } else if (pin.length < maxLen) {
          pin += val;
        }
        updateDots();

        if (pin.length >= 3) {
          // Allow 3 or 4 digit PINs — submit after a short delay
          setTimeout(() => {
            if (pin.length >= 3) submitPin();
          }, 300);
        }
      });
    });

    // Keyboard support
    document.addEventListener('keydown', function pinKeyHandler(e) {
      if (e.key >= '0' && e.key <= '9' && pin.length < maxLen) {
        pin += e.key;
        updateDots();
        if (pin.length >= 3) {
          setTimeout(() => { if (pin.length >= 3) submitPin(); }, 300);
        }
      } else if (e.key === 'Backspace') {
        pin = pin.slice(0, -1);
        updateDots();
      }
    });

    // Back button
    document.getElementById('pin-back')?.addEventListener('click', () => navigate('/profiles'));
  }, 0);

  return container;
}
