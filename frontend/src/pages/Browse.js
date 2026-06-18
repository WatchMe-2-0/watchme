// ═══════════════════════════════════════════════════════════════════
// WATCHME — Browse Page (TMDB)
// Trending, Top Rated, By Genre
// ═══════════════════════════════════════════════════════════════════

import { api } from '../api/client.js';
import { navigate } from '../router/router.js';
import { toastError } from '../components/Toast.js';

export async function renderBrowse() {
  const container = document.createElement('div');
  container.id = 'browse-page';

  let trending = [], topRated = [], genres = [];
  try {
    const [trendRes, topRes, genreRes] = await Promise.all([
      api.tmdbTrending().catch(() => ({ data: { results: [] } })),
      api.tmdbTopRated().catch(() => ({ data: { results: [] } })),
      api.tmdbGenres().catch(() => ({ data: [] })),
    ]);
    trending = trendRes.data?.results || [];
    topRated = topRes.data?.results || [];
    genres = genreRes.data || [];
  } catch (e) {
    toastError('Failed to load TMDB data. Check your API key in Settings.');
  }

  function movieRow(title, movies) {
    if (!movies.length) return `<p class="text-dim text-sm" style="padding:var(--space-md)">Configure your TMDB API key in Settings to browse movies.</p>`;
    return `
      <section class="browse-section">
        <h3 class="browse-section-title">${title}</h3>
        <div class="browse-row">
          ${movies.slice(0, 20).map(m => `
            <div class="browse-card" data-tmdb-id="${m.id}">
              <img src="${m.poster_path || ''}" alt="${m.title}" loading="lazy" onerror="this.style.display='none'" />
              <div class="browse-card-overlay">
                <div class="browse-card-title">${m.title}</div>
                <div class="browse-card-meta">⭐ ${m.vote_average?.toFixed(1) || '—'} · ${(m.release_date || '').slice(0, 4)}</div>
              </div>
            </div>
          `).join('')}
        </div>
      </section>
    `;
  }

  const genreTabs = genres.slice(0, 12).map(g =>
    `<button class="genre-tab btn btn-ghost" data-genre-id="${g.id}">${g.name}</button>`
  ).join('');

  container.innerHTML = `
    <style>
      #browse-page {
        margin-left: var(--sidebar-width);
        padding: var(--space-xl) var(--space-2xl);
        min-height: 100vh;
      }

      .browse-header {
        display: flex;
        align-items: center;
        gap: var(--space-lg);
        margin-bottom: var(--space-2xl);
      }

      .browse-title {
        font-family: var(--font-mono);
        font-size: 1.4rem;
      }

      .browse-search {
        max-width: 300px;
      }

      .browse-section {
        margin-bottom: var(--space-2xl);
      }

      .browse-section-title {
        font-family: var(--font-mono);
        font-size: 1rem;
        color: var(--color-text-dim);
        margin-bottom: var(--space-md);
        letter-spacing: 0.03em;
      }

      .browse-row {
        display: flex;
        gap: var(--space-md);
        overflow-x: auto;
        padding-bottom: var(--space-md);
        scroll-snap-type: x mandatory;
      }

      .browse-row::-webkit-scrollbar { height: 4px; }

      .browse-card {
        flex: 0 0 150px;
        aspect-ratio: 2/3;
        border-radius: var(--radius-md);
        overflow: hidden;
        cursor: pointer;
        position: relative;
        scroll-snap-align: start;
        transition: transform var(--transition-normal);
        border: 1px solid transparent;
      }

      .browse-card:hover {
        transform: scale(1.08);
        border-color: var(--color-primary);
        box-shadow: 0 8px 30px var(--color-primary-glow);
        z-index: 10;
      }

      .browse-card img {
        width: 100%;
        height: 100%;
        object-fit: cover;
      }

      .browse-card-overlay {
        position: absolute;
        bottom: 0;
        left: 0;
        right: 0;
        padding: var(--space-sm);
        background: linear-gradient(transparent, hsla(0,0%,0%,0.9));
        opacity: 0;
        transition: opacity var(--transition-fast);
      }

      .browse-card:hover .browse-card-overlay {
        opacity: 1;
      }

      .browse-card-title {
        font-family: var(--font-mono);
        font-size: 0.7rem;
        font-weight: 500;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
      }

      .browse-card-meta {
        font-size: 0.6rem;
        color: var(--color-text-dim);
      }

      .genre-tabs {
        display: flex;
        flex-wrap: wrap;
        gap: var(--space-sm);
        margin-bottom: var(--space-lg);
      }

      .genre-tab {
        font-size: 0.75rem;
        padding: var(--space-xs) var(--space-md);
      }

      .genre-results {
        display: grid;
        grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
        gap: var(--space-md);
      }
    </style>

    <div class="browse-header">
      <h1 class="browse-title glow-text">Browse Movies</h1>
      <input type="text" class="input browse-search" id="browse-search" placeholder="🔍 Search movies..." />
    </div>

    <div id="browse-content">
      ${movieRow('🔥 Trending This Week', trending)}
      ${movieRow('⭐ Top Rated', topRated)}

      <section class="browse-section">
        <h3 class="browse-section-title">🎭 By Genre</h3>
        <div class="genre-tabs">${genreTabs}</div>
        <div id="genre-results" class="genre-results"></div>
      </section>
    </div>

    <div id="search-results" style="display:none">
      <h3 class="browse-section-title">🔍 Search Results</h3>
      <div id="search-grid" class="genre-results"></div>
    </div>
  `;

  setTimeout(() => {
    // Movie click → detail page
    container.addEventListener('click', (e) => {
      const card = e.target.closest('[data-tmdb-id]');
      if (card) navigate(`/movie/${card.dataset.tmdbId}`);
    });

    // Search
    let searchTimeout;
    document.getElementById('browse-search')?.addEventListener('input', (e) => {
      clearTimeout(searchTimeout);
      const query = e.target.value.trim();

      if (!query) {
        document.getElementById('browse-content').style.display = '';
        document.getElementById('search-results').style.display = 'none';
        return;
      }

      searchTimeout = setTimeout(async () => {
        try {
          const result = await api.tmdbSearch(query);
          const movies = result.data?.results || [];

          document.getElementById('browse-content').style.display = 'none';
          document.getElementById('search-results').style.display = '';
          document.getElementById('search-grid').innerHTML = movies.map(m => `
            <div class="browse-card" data-tmdb-id="${m.id}">
              <img src="${m.poster_path || ''}" alt="${m.title}" loading="lazy" onerror="this.style.display='none'" />
              <div class="browse-card-overlay" style="opacity:1">
                <div class="browse-card-title">${m.title}</div>
                <div class="browse-card-meta">⭐ ${m.vote_average?.toFixed(1)} · ${(m.release_date||'').slice(0,4)}</div>
              </div>
            </div>
          `).join('');
        } catch (e) {
          toastError('Search failed');
        }
      }, 400);
    });

    // Genre tabs
    container.querySelectorAll('[data-genre-id]').forEach(tab => {
      tab.addEventListener('click', async () => {
        const genreId = tab.dataset.genreId;
        container.querySelectorAll('.genre-tab').forEach(t => t.classList.remove('btn-primary'));
        tab.classList.add('btn-primary');
        tab.classList.remove('btn-ghost');

        try {
          const result = await api.tmdbByGenre(genreId);
          const movies = result.data?.results || [];
          document.getElementById('genre-results').innerHTML = movies.map(m => `
            <div class="browse-card" data-tmdb-id="${m.id}">
              <img src="${m.poster_path || ''}" alt="${m.title}" loading="lazy" onerror="this.style.display='none'" />
              <div class="browse-card-overlay" style="opacity:1">
                <div class="browse-card-title">${m.title}</div>
              </div>
            </div>
          `).join('');
        } catch (e) {
          toastError('Failed to load genre');
        }
      });
    });
  }, 0);

  return container;
}
