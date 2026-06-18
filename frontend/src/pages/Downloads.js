// ═══════════════════════════════════════════════════════════════════
// WATCHME — Downloads Page
// ═══════════════════════════════════════════════════════════════════

import { api } from '../api/client.js';
import { toastSuccess, toastError } from '../components/Toast.js';

export async function renderDownloads() {
  const container = document.createElement('div');
  container.id = 'downloads-page';

  let downloads = [];
  try {
    const result = await api.getDownloads();
    downloads = result.data || [];
  } catch (e) {}

  function formatBytes(bytes) {
    if (!bytes) return '—';
    const units = ['B', 'KB', 'MB', 'GB'];
    let i = 0;
    while (bytes >= 1024 && i < units.length - 1) { bytes /= 1024; i++; }
    return `${bytes.toFixed(1)} ${units[i]}`;
  }

  function formatSpeed(bps) {
    if (!bps) return '—';
    return formatBytes(bps) + '/s';
  }

  function formatETA(seconds) {
    if (!seconds) return '—';
    const m = Math.floor(seconds / 60);
    const h = Math.floor(m / 60);
    if (h > 0) return `${h}h ${m % 60}m`;
    return `${m}m ${seconds % 60}s`;
  }

  const statusColors = {
    queued: 'var(--color-warning)',
    downloading: 'var(--color-primary)',
    completed: 'var(--color-success)',
    failed: 'var(--color-danger)',
    cancelled: 'var(--color-text-muted)',
  };

  container.innerHTML = `
    <style>
      #downloads-page {
        margin-left: var(--sidebar-width);
        padding: var(--space-xl) var(--space-2xl);
        min-height: 100vh;
      }

      .dl-header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        margin-bottom: var(--space-2xl);
      }

      .dl-title {
        font-family: var(--font-mono);
        font-size: 1.4rem;
      }

      .dl-list {
        display: flex;
        flex-direction: column;
        gap: var(--space-md);
      }

      .dl-item {
        padding: var(--space-lg);
        border-radius: var(--radius-md);
        animation: fadeInUp 0.4s ease both;
      }

      .dl-item-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: var(--space-sm);
      }

      .dl-item-title {
        font-family: var(--font-mono);
        font-size: 0.9rem;
        font-weight: 500;
      }

      .dl-item-status {
        font-family: var(--font-mono);
        font-size: 0.75rem;
        padding: 2px 8px;
        border-radius: var(--radius-full);
        border: 1px solid;
      }

      .dl-progress-bar {
        height: 6px;
        background: hsla(0,0%,100%,0.1);
        border-radius: var(--radius-full);
        overflow: hidden;
        margin-bottom: var(--space-sm);
      }

      .dl-progress-fill {
        height: 100%;
        background: linear-gradient(90deg, var(--color-success), hsl(145,80%,55%));
        border-radius: var(--radius-full);
        transition: width 0.5s ease;
      }

      .dl-item-meta {
        display: flex;
        gap: var(--space-lg);
        font-size: 0.75rem;
        color: var(--color-text-dim);
        font-family: var(--font-mono);
      }

      .dl-empty {
        text-align: center;
        padding: var(--space-3xl);
        color: var(--color-text-dim);
        font-family: var(--font-mono);
      }
    </style>

    <div class="dl-header">
      <h1 class="dl-title glow-text">📥 Downloads</h1>
    </div>

    <div class="dl-list" id="dl-list">
      ${downloads.length === 0 ? '<div class="dl-empty">No downloads yet. Paste a magnet link from the dashboard.</div>' : ''}
      ${downloads.map((d, i) => `
        <div class="dl-item glass" style="animation-delay:${i * 100}ms" data-dl-id="${d.id}">
          <div class="dl-item-header">
            <span class="dl-item-title">${d.title || d.info_hash?.slice(0, 12) || 'Unknown'}</span>
            <span class="dl-item-status" style="color:${statusColors[d.status] || 'var(--color-text)'};border-color:${statusColors[d.status] || 'var(--color-text)'}">
              ${d.status}
            </span>
          </div>
          <div class="dl-progress-bar">
            <div class="dl-progress-fill" style="width:${d.live_progress || d.progress || 0}%"></div>
          </div>
          <div class="dl-item-meta">
            <span>📊 ${(d.live_progress || d.progress || 0).toFixed(1)}%</span>
            <span>⚡ ${formatSpeed(d.live_speed || d.speed)}</span>
            <span>👥 ${d.live_peers || d.peers || 0} peers</span>
            <span>⏱ ETA: ${formatETA(d.live_eta || d.eta)}</span>
            <span>📦 ${formatBytes(d.downloaded || 0)} / ${formatBytes(d.file_size || 0)}</span>
            ${d.status === 'downloading' || d.status === 'queued'
              ? `<button class="btn btn-ghost text-xs cancel-dl" data-cancel-id="${d.id}" style="padding:2px 8px;font-size:0.65rem;">✕ Cancel</button>`
              : ''
            }
          </div>
        </div>
      `).join('')}
    </div>
  `;

  setTimeout(() => {
    // Cancel buttons
    container.querySelectorAll('.cancel-dl').forEach(btn => {
      btn.addEventListener('click', async (e) => {
        e.stopPropagation();
        const id = btn.dataset.cancelId;
        try {
          await api.cancelDownload(id);
          toastSuccess('Download cancelled');
          btn.closest('.dl-item')?.remove();
        } catch (e) {
          toastError(e.message);
        }
      });
    });

    // Subscribe to SSE progress updates
    try {
      const eventSource = api.subscribeProgress((update) => {
        const item = container.querySelector(`[data-dl-id="${update.download_id}"]`);
        if (!item) return;

        const fill = item.querySelector('.dl-progress-fill');
        if (fill) fill.style.width = `${update.progress}%`;

        const meta = item.querySelector('.dl-item-meta');
        if (meta) {
          const spans = meta.querySelectorAll('span');
          if (spans[0]) spans[0].textContent = `📊 ${update.progress.toFixed(1)}%`;
          if (spans[1]) spans[1].textContent = `⚡ ${formatSpeed(update.speed)}`;
          if (spans[2]) spans[2].textContent = `👥 ${update.peers} peers`;
          if (spans[3]) spans[3].textContent = `⏱ ETA: ${formatETA(update.eta)}`;
        }

        if (update.status === 'completed') {
          const statusBadge = item.querySelector('.dl-item-status');
          if (statusBadge) {
            statusBadge.textContent = 'completed';
            statusBadge.style.color = 'var(--color-success)';
            statusBadge.style.borderColor = 'var(--color-success)';
          }
        }
      });
    } catch (e) {}
  }, 0);

  return container;
}
