// ═══════════════════════════════════════════════════════════════════
// WATCHME — Dashboard Page
// Sidebar + Movie Grid + Active Downloads
// ═══════════════════════════════════════════════════════════════════

import { api } from '../api/client.js';
import { navigate } from '../router/router.js';
import { store } from '../state/store.js';
import { toastSuccess, toastError } from '../components/Toast.js';

export async function renderDashboard() {
  const container = document.createElement('div');
  container.id = 'dashboard-page';

  // Fetch movies
  let movies = [];
  let downloads = [];
  try {
    const [moviesRes, downloadsRes] = await Promise.all([
      api.getMovies().catch(() => ({ data: { movies: [] } })),
      api.getDownloads().catch(() => ({ data: [] })),
    ]);
    movies = moviesRes.data?.movies || [];
    downloads = (downloadsRes.data || []).filter(d => d.status === 'downloading' || d.status === 'queued');
  } catch (e) {}

  const movieCards = movies.length > 0
    ? movies.map((m, i) => `
      <div class="movie-card glass animate-fade-in-up delay-${(i % 5) + 1}" data-movie-id="${m.id}">
        <div class="movie-poster">
          ${m.poster_url
            ? `<img src="${m.poster_url}" alt="${m.title}" loading="lazy" />`
            : `<div class="movie-poster-placeholder">🎬</div>`
          }
        </div>
        <div class="movie-info">
          <div class="movie-title">${m.title}</div>
          ${m.rating ? `<div class="movie-rating">⭐ ${m.rating}</div>` : ''}
        </div>
      </div>
    `).join('')
    : '<div class="empty-state animate-fade-in">No movies yet. Use the magnet link to add content.</div>';

  const downloadBars = downloads.map(d => `
    <div class="download-item glass">
      <div class="download-info">
        <span class="download-title">${d.title || 'Downloading...'}</span>
        <span class="download-status">${d.live_progress ? d.live_progress.toFixed(1) + '%' : d.status}</span>
      </div>
      <div class="download-bar">
        <div class="download-progress" style="width: ${d.live_progress || d.progress || 0}%"></div>
      </div>
    </div>
  `).join('');

  container.innerHTML = `
    <style>
      #dashboard-page {
        display: flex;
        min-height: 100vh;
      }

      /* ── Sidebar ── */
      .sidebar {
        width: var(--sidebar-width);
        min-height: 100vh;
        padding: var(--space-lg) var(--space-sm);
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: var(--space-md);
        position: fixed;
        left: 0;
        top: 0;
        z-index: 100;
        border-right: 1px solid var(--glass-border);
        animation: slideInLeft 0.4s ease both;
      }

      .sidebar-icon {
        width: 48px;
        height: 48px;
        display: flex;
        align-items: center;
        justify-content: center;
        border-radius: var(--radius-md);
        font-size: 1.3rem;
        cursor: pointer;
        transition: all var(--transition-normal);
        border: 1px solid transparent;
        position: relative;
      }

      .sidebar-icon:hover {
        background: var(--color-surface-hover);
        border-color: var(--glass-border-hover);
        transform: scale(1.1);
      }

      .sidebar-icon.active {
        background: var(--color-surface-active);
        border-color: var(--color-primary);
        box-shadow: 0 0 15px var(--color-primary-glow);
      }

      .sidebar-icon .tooltip {
        position: absolute;
        left: 60px;
        background: var(--color-surface);
        padding: 4px 10px;
        border-radius: var(--radius-sm);
        font-family: var(--font-mono);
        font-size: 0.7rem;
        white-space: nowrap;
        opacity: 0;
        pointer-events: none;
        transition: opacity var(--transition-fast);
        border: 1px solid var(--glass-border);
      }

      .sidebar-icon:hover .tooltip {
        opacity: 1;
      }

      .sidebar-spacer { flex: 1; }

      .sidebar-wordmark {
        font-family: var(--font-mono);
        font-size: 0.55rem;
        font-weight: 600;
        letter-spacing: 0.15em;
        writing-mode: vertical-rl;
        text-orientation: mixed;
        padding: var(--space-md) 0;
      }

      /* ── Main Content ── */
      .main-content {
        flex: 1;
        margin-left: var(--sidebar-width);
        padding: var(--space-xl) var(--space-2xl);
      }

      .section-title {
        font-family: var(--font-mono);
        font-size: 1.1rem;
        margin-bottom: var(--space-lg);
        color: var(--color-text-dim);
        letter-spacing: 0.05em;
      }

      /* ── Downloads ── */
      .downloads-section {
        margin-bottom: var(--space-2xl);
      }

      .download-item {
        padding: var(--space-sm) var(--space-md);
        border-radius: var(--radius-md);
        margin-bottom: var(--space-sm);
      }

      .download-info {
        display: flex;
        justify-content: space-between;
        font-family: var(--font-mono);
        font-size: 0.8rem;
        margin-bottom: var(--space-xs);
      }

      .download-title { color: var(--color-text); }
      .download-status { color: var(--color-success); }

      .download-bar {
        height: 4px;
        background: hsla(0,0%,100%,0.1);
        border-radius: var(--radius-full);
        overflow: hidden;
      }

      .download-progress {
        height: 100%;
        background: linear-gradient(90deg, var(--color-success), hsl(145, 80%, 55%));
        border-radius: var(--radius-full);
        transition: width 0.5s ease;
        animation: progress-glow 2s ease-in-out infinite;
      }

      /* ── Movie Grid ── */
      .movie-grid {
        display: grid;
        grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
        gap: var(--space-lg);
      }

      .movie-card {
        border-radius: var(--radius-md);
        overflow: hidden;
        cursor: pointer;
        transition: all var(--transition-normal);
      }

      .movie-card:hover {
        transform: translateY(-6px) scale(1.03);
        box-shadow: 0 12px 35px hsla(0,0%,0%,0.5);
        border-color: var(--color-primary);
      }

      .movie-poster {
        aspect-ratio: 2/3;
        overflow: hidden;
        background: var(--color-surface);
      }

      .movie-poster img {
        width: 100%;
        height: 100%;
        object-fit: cover;
        transition: transform var(--transition-slow);
      }

      .movie-card:hover .movie-poster img {
        transform: scale(1.08);
      }

      .movie-poster-placeholder {
        width: 100%;
        height: 100%;
        display: flex;
        align-items: center;
        justify-content: center;
        font-size: 3rem;
        background: linear-gradient(135deg, var(--color-surface), var(--color-surface-hover));
      }

      .movie-info {
        padding: var(--space-sm) var(--space-md);
      }

      .movie-title {
        font-family: var(--font-mono);
        font-size: 0.8rem;
        font-weight: 500;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
      }

      .movie-rating {
        font-size: 0.7rem;
        color: var(--color-text-dim);
        margin-top: 2px;
      }

      .empty-state {
        grid-column: 1 / -1;
        text-align: center;
        padding: var(--space-3xl);
        color: var(--color-text-dim);
        font-family: var(--font-mono);
        font-size: 0.9rem;
      }

      /* ── Magnet Modal ── */
      .modal-overlay {
        display: none;
        position: fixed;
        inset: 0;
        background: hsla(0,0%,0%,0.7);
        z-index: 1000;
        align-items: center;
        justify-content: center;
        backdrop-filter: blur(5px);
      }

      .modal-overlay.active {
        display: flex;
      }

      .modal-card {
        width: 90%;
        max-width: 500px;
        padding: var(--space-2xl);
        border-radius: var(--radius-lg);
        animation: fadeInScale 0.3s ease both;
      }

      .modal-title {
        font-family: var(--font-mono);
        font-size: 1.1rem;
        margin-bottom: var(--space-lg);
      }

      .modal-field {
        margin-bottom: var(--space-md);
      }

      .modal-field label {
        display: block;
        font-family: var(--font-mono);
        font-size: 0.75rem;
        color: var(--color-text-dim);
        margin-bottom: var(--space-xs);
        text-transform: uppercase;
      }

      .modal-actions {
        display: flex;
        gap: var(--space-md);
        justify-content: flex-end;
        margin-top: var(--space-lg);
      }
    </style>

    <!-- Sidebar -->
    <nav class="sidebar glass">
      <div class="sidebar-icon active" data-nav="dashboard">
        🏠
        <span class="tooltip">Home</span>
      </div>
      <div class="sidebar-icon" data-nav="magnet" id="magnet-btn">
        🔗
        <span class="tooltip">Magnet Link</span>
      </div>
      <div class="sidebar-icon" data-nav="browse">
        🎬
        <span class="tooltip">Browse Movies</span>
      </div>
      <div class="sidebar-icon" data-nav="downloads">
        📥
        <span class="tooltip">Downloads</span>
      </div>
      <div class="sidebar-icon" data-nav="settings">
        ⚙️
        <span class="tooltip">Settings</span>
      </div>

      <div class="sidebar-spacer"></div>

      <div class="sidebar-wordmark glow-text animate-glow">WATCHME</div>
    </nav>

    <!-- Main Content -->
    <main class="main-content">
      ${downloads.length > 0 ? `
        <section class="downloads-section">
          <h2 class="section-title">📥 Active Downloads</h2>
          ${downloadBars}
        </section>
      ` : ''}

      <section>
        <h2 class="section-title">🎬 Your Movies</h2>
        <div class="movie-grid">
          ${movieCards}
        </div>
      </section>
    </main>

    <!-- Magnet Link Modal -->
    <div class="modal-overlay" id="magnet-modal">
      <div class="modal-card glass">
        <h3 class="modal-title glow-text">🧲 Paste Magnet Link</h3>
        <div class="modal-field">
          <label for="magnet-title">Movie Title</label>
          <input type="text" id="magnet-title" class="input" placeholder="e.g. Interstellar" />
        </div>
        <div class="modal-field">
          <label for="magnet-link">Magnet Link</label>
          <input type="text" id="magnet-link" class="input" placeholder="magnet:?xt=urn:btih:..." />
        </div>
        <div class="modal-actions">
          <button class="btn btn-ghost" id="magnet-cancel">Cancel</button>
          <button class="btn btn-primary" id="magnet-submit">🚀 Start Download</button>
        </div>
      </div>
    </div>
  `;

  setTimeout(() => {
    // Sidebar navigation
    container.querySelectorAll('[data-nav]').forEach(icon => {
      icon.addEventListener('click', () => {
        const nav = icon.dataset.nav;
        if (nav === 'dashboard') navigate('/dashboard');
        else if (nav === 'browse') navigate('/browse');
        else if (nav === 'downloads') navigate('/downloads');
        else if (nav === 'settings') navigate('/settings');
      });
    });

    // Magnet modal
    const modal = document.getElementById('magnet-modal');
    document.getElementById('magnet-btn')?.addEventListener('click', () => modal?.classList.add('active'));
    document.getElementById('magnet-cancel')?.addEventListener('click', () => modal?.classList.remove('active'));
    modal?.addEventListener('click', (e) => { if (e.target === modal) modal.classList.remove('active'); });

    document.getElementById('magnet-submit')?.addEventListener('click', async () => {
      const title = document.getElementById('magnet-title')?.value.trim();
      const link = document.getElementById('magnet-link')?.value.trim();
      if (!link) return toastError('Magnet link is required');

      try {
        await api.startDownload(link, title);
        toastSuccess('Download started!');
        modal?.classList.remove('active');
        navigate('/downloads');
      } catch (e) {
        toastError(e.message);
      }
    });

    // Movie card clicks
    container.querySelectorAll('[data-movie-id]').forEach(card => {
      card.addEventListener('click', () => navigate(`/movie/${card.dataset.movieId}`));
    });
  }, 0);

  return container;
}
