// ═══════════════════════════════════════════════════════════════════
// WATCHME — Profile Selection Page (Netflix-style)
// ═══════════════════════════════════════════════════════════════════

import { api } from '../api/client.js';
import { navigate } from '../router/router.js';
import { toastError } from '../components/Toast.js';

const avatarEmojis = {
  aurora: '🌌', nebula: '🌀', comet: '☄️', galaxy: '🌟',
  nova: '💫', pulsar: '⚡', quasar: '🔮', stellar: '⭐',
  orbit: '🪐', prism: '🔷', flux: '🌊', echo: '🎭',
};

export async function renderProfileSelect() {
  const container = document.createElement('div');
  container.id = 'profile-select-page';

  let profiles = [];
  try {
    const result = await api.getProfiles();
    profiles = result.data || [];
  } catch (e) {
    toastError('Failed to load profiles');
  }

  const profileCards = profiles.map((p, i) => `
    <div class="profile-card glass animate-fade-in-up delay-${i + 1}" data-profile-id="${p.id}">
      <div class="profile-avatar">${avatarEmojis[p.avatar] || '🌌'}</div>
      <div class="profile-name">${p.name}</div>
      ${p.is_kids ? '<span class="profile-badge">KIDS</span>' : ''}
    </div>
  `).join('');

  container.innerHTML = `
    <style>
      #profile-select-page {
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        min-height: 100vh;
        padding: var(--space-2xl);
      }

      .profile-title {
        font-family: var(--font-mono);
        font-size: 1.5rem;
        margin-bottom: var(--space-3xl);
        animation: fadeInUp 0.5s ease both;
      }

      .profiles-grid {
        display: flex;
        flex-wrap: wrap;
        gap: var(--space-xl);
        justify-content: center;
        max-width: 800px;
      }

      .profile-card {
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: var(--space-md);
        padding: var(--space-xl) var(--space-2xl);
        border-radius: var(--radius-lg);
        cursor: pointer;
        transition: all var(--transition-normal);
        min-width: 140px;
      }

      .profile-card:hover {
        transform: translateY(-8px) scale(1.05);
        border-color: var(--color-primary);
        box-shadow: 0 12px 40px var(--color-primary-glow);
      }

      .profile-avatar {
        font-size: 3rem;
        transition: transform var(--transition-spring);
      }

      .profile-card:hover .profile-avatar {
        transform: scale(1.15);
      }

      .profile-name {
        font-family: var(--font-mono);
        font-size: 0.9rem;
        font-weight: 500;
      }

      .profile-badge {
        font-family: var(--font-mono);
        font-size: 0.65rem;
        padding: 2px 8px;
        border-radius: var(--radius-full);
        background: var(--color-primary-dim);
        color: var(--color-primary-bright);
        letter-spacing: 0.1em;
      }

      .add-profile-card {
        border: 2px dashed var(--glass-border);
        background: transparent;
        opacity: 0.6;
      }

      .add-profile-card:hover {
        opacity: 1;
        border-color: var(--color-primary);
      }

      .add-icon {
        font-size: 2.5rem;
        color: var(--color-text-dim);
      }
    </style>

    <h1 class="profile-title glow-text">Who's watching?</h1>

    <div class="profiles-grid">
      ${profileCards}
      <div class="profile-card add-profile-card animate-fade-in-up delay-${profiles.length + 1}" id="add-profile-btn">
        <div class="add-icon">+</div>
        <div class="profile-name text-dim">Add Profile</div>
      </div>
    </div>
  `;

  setTimeout(() => {
    // Click on profile → go to PIN entry
    container.querySelectorAll('[data-profile-id]').forEach(card => {
      card.addEventListener('click', () => {
        const id = card.dataset.profileId;
        navigate(`/pin/${id}`);
      });
    });

    // Add profile button
    document.getElementById('add-profile-btn')?.addEventListener('click', () => {
      navigate('/settings');
    });
  }, 0);

  return container;
}
