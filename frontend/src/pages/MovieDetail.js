// ═══════════════════════════════════════════════════════════════════
// WATCHME — Movie Detail Page
// ═══════════════════════════════════════════════════════════════════

import { api } from '../api/client.js';
import { navigate } from '../router/router.js';
import { toastSuccess, toastError } from '../components/Toast.js';

export async function renderMovieDetail(params) {
  const movieId = params.id;
  const container = document.createElement('div');
  container.id = 'movie-detail-page';

  let movie = null, cert = '';
  try {
    const result = await api.tmdbMovie(movieId);
    movie = result.data?.movie;
    cert = result.data?.certification || '';
  } catch (e) {
    // Try local movie
    try {
      const localResult = await api.getMovie(movieId);
      movie = localResult.data?.movie;
    } catch (e2) {
      return `<div style="padding:4rem; text-align:center; color:var(--color-text-dim)">Movie not found</div>`;
    }
  }

  if (!movie) return `<div style="padding:4rem; text-align:center; color:var(--color-text-dim)">Movie not found</div>`;

  const genres = movie.genres ? movie.genres.map(g => g.name).join(', ') : (movie.genre || '');
  const year = movie.release_date ? movie.release_date.slice(0, 4) : (movie.year || '');

  container.innerHTML = `
    <style>
      #movie-detail-page {
        margin-left: var(--sidebar-width);
        min-height: 100vh;
      }

      .movie-hero {
        position: relative;
        height: 50vh;
        min-height: 350px;
        overflow: hidden;
      }

      .movie-hero-bg {
        position: absolute;
        inset: 0;
        background-size: cover;
        background-position: center;
        filter: brightness(0.4) blur(2px);
      }

      .movie-hero-gradient {
        position: absolute;
        inset: 0;
        background: linear-gradient(transparent 30%, var(--color-bg) 100%);
      }

      .movie-hero-content {
        position: relative;
        display: flex;
        gap: var(--space-2xl);
        align-items: flex-end;
        height: 100%;
        padding: var(--space-2xl);
        max-width: 1000px;
      }

      .movie-detail-poster {
        width: 180px;
        border-radius: var(--radius-md);
        box-shadow: 0 8px 30px hsla(0,0%,0%,0.6);
        animation: fadeInUp 0.5s ease both;
      }

      .movie-detail-info {
        animation: fadeInUp 0.6s ease both 0.1s;
      }

      .movie-detail-title {
        font-family: var(--font-mono);
        font-size: 2rem;
        font-weight: 700;
        margin-bottom: var(--space-sm);
      }

      .movie-detail-meta {
        display: flex;
        gap: var(--space-md);
        align-items: center;
        font-size: 0.85rem;
        color: var(--color-text-dim);
        margin-bottom: var(--space-md);
        flex-wrap: wrap;
      }

      .movie-detail-badge {
        padding: 2px 8px;
        border-radius: var(--radius-sm);
        font-family: var(--font-mono);
        font-size: 0.7rem;
        border: 1px solid var(--glass-border);
      }

      .movie-detail-body {
        padding: var(--space-2xl);
        max-width: 1000px;
      }

      .movie-overview {
        font-size: 0.95rem;
        line-height: 1.8;
        color: var(--color-text-dim);
        margin-bottom: var(--space-2xl);
        animation: fadeInUp 0.7s ease both 0.2s;
      }

      .movie-actions {
        display: flex;
        gap: var(--space-md);
        animation: fadeInUp 0.8s ease both 0.3s;
      }
    </style>

    <div class="movie-hero">
      <div class="movie-hero-bg" style="background-image: url('${movie.backdrop_path || movie.poster_path || ''}')"></div>
      <div class="movie-hero-gradient"></div>
      <div class="movie-hero-content">
        ${movie.poster_path ? `<img class="movie-detail-poster" src="${movie.poster_path}" alt="${movie.title}" />` : ''}
        <div class="movie-detail-info">
          <h1 class="movie-detail-title">${movie.title}</h1>
          <div class="movie-detail-meta">
            ${year ? `<span>${year}</span>` : ''}
            ${movie.runtime ? `<span>${movie.runtime} min</span>` : ''}
            ${movie.vote_average ? `<span>⭐ ${movie.vote_average.toFixed(1)}</span>` : ''}
            ${cert ? `<span class="movie-detail-badge">${cert}</span>` : ''}
            ${genres ? `<span>${genres}</span>` : ''}
          </div>
        </div>
      </div>
    </div>

    <div class="movie-detail-body">
      ${movie.overview ? `<p class="movie-overview">${movie.overview}</p>` : ''}
      ${movie.tagline ? `<p class="text-dim text-sm" style="margin-bottom:var(--space-xl); font-style:italic;">"${movie.tagline}"</p>` : ''}

      <div class="movie-actions">
        <button class="btn btn-ghost" id="back-btn">← Back</button>
        ${movie.file_path ? `<button class="btn btn-primary" id="play-btn">▶ Play</button>` : ''}
        ${movie.id && !movie.file_path ? `<button class="btn btn-primary" id="magnet-btn">🧲 Paste Magnet Link</button>` : ''}
      </div>
    </div>
  `;

  setTimeout(() => {
    document.getElementById('back-btn')?.addEventListener('click', () => history.back());
    document.getElementById('play-btn')?.addEventListener('click', () => {
      if (movie.id) {
        const streamUrl = api.getStreamUrl(movie.id);
        window.open(streamUrl, '_blank');
      }
    });
  }, 0);

  return container;
}
