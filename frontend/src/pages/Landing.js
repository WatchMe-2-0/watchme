// ═══════════════════════════════════════════════════════════════════
// WATCHME — Landing Page
// Full-screen aurora with glowing WATCHME logo
// ═══════════════════════════════════════════════════════════════════

import { navigate } from '../router/router.js';
import { store } from '../state/store.js';

export function renderLanding() {
  const container = document.createElement('div');
  container.id = 'landing-page';
  container.innerHTML = `
    <style>
      #landing-page {
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        min-height: 100vh;
        padding: var(--space-xl);
        text-align: center;
      }

      .landing-logo {
        font-family: var(--font-mono);
        font-size: clamp(3rem, 10vw, 7rem);
        font-weight: 700;
        letter-spacing: 0.15em;
        animation: fadeInScale 1.5s cubic-bezier(0.34, 1.56, 0.64, 1) both,
                   glow-pulse 3s ease-in-out infinite 1.5s;
      }

      .landing-tagline {
        font-family: var(--font-sans);
        font-size: clamp(0.9rem, 2vw, 1.1rem);
        color: var(--color-text-dim);
        margin-top: var(--space-lg);
        animation: fadeInUp 1s ease both 0.8s;
        letter-spacing: 0.05em;
      }

      .landing-enter {
        margin-top: var(--space-3xl);
        animation: fadeInUp 1s ease both 1.2s;
      }

      .landing-enter .btn {
        font-size: 1rem;
        padding: var(--space-md) var(--space-2xl);
        border-radius: var(--radius-xl);
        letter-spacing: 0.08em;
      }

      .landing-version {
        position: absolute;
        bottom: var(--space-xl);
        font-family: var(--font-mono);
        font-size: 0.7rem;
        color: var(--color-text-muted);
        animation: fadeIn 1s ease both 2s;
      }
    </style>

    <h1 class="landing-logo glow-text-strong">WATCHME</h1>
    <p class="landing-tagline">Your personal cinema. Anywhere on your network.</p>

    <div class="landing-enter">
      <button id="landing-enter-btn" class="btn btn-primary">
        ENTER →
      </button>
    </div>

    <div class="landing-version">v2.0.0</div>
  `;

  // Wire up enter button
  setTimeout(() => {
    const btn = document.getElementById('landing-enter-btn');
    if (btn) {
      btn.addEventListener('click', () => {
        if (store.get('isSetupComplete')) {
          navigate('/profiles');
        } else {
          navigate('/setup');
        }
      });
    }
  }, 0);

  return container;
}
