// ═══════════════════════════════════════════════════════════════════
// WATCHME — Video Player Component
// Custom glassmorphism video player with controls
// ═══════════════════════════════════════════════════════════════════

export function createVideoPlayer(streamUrl, title = '') {
  const overlay = document.createElement('div');
  overlay.id = 'video-player-overlay';

  overlay.innerHTML = `
    <style>
      #video-player-overlay {
        position: fixed;
        inset: 0;
        z-index: 9999;
        background: #000;
        display: flex;
        flex-direction: column;
        animation: fadeIn 0.3s ease both;
      }

      .vp-topbar {
        position: absolute;
        top: 0;
        left: 0;
        right: 0;
        z-index: 10;
        display: flex;
        align-items: center;
        gap: var(--space-md);
        padding: var(--space-md) var(--space-xl);
        background: linear-gradient(to bottom, hsla(0,0%,0%,0.8), transparent);
        opacity: 0;
        transition: opacity var(--transition-normal);
      }

      .vp-topbar.visible { opacity: 1; }

      .vp-back {
        width: 40px;
        height: 40px;
        border-radius: var(--radius-full);
        border: 1px solid hsla(0,0%,100%,0.15);
        background: hsla(0,0%,0%,0.5);
        color: white;
        font-size: 1.1rem;
        cursor: pointer;
        display: flex;
        align-items: center;
        justify-content: center;
        transition: all var(--transition-fast);
        backdrop-filter: blur(10px);
      }

      .vp-back:hover {
        background: hsla(0,0%,100%,0.15);
        border-color: var(--color-primary);
        transform: scale(1.1);
      }

      .vp-title {
        font-family: var(--font-mono);
        font-size: 0.9rem;
        color: white;
        opacity: 0.9;
      }

      .vp-video-container {
        flex: 1;
        display: flex;
        align-items: center;
        justify-content: center;
        cursor: pointer;
        position: relative;
      }

      .vp-video {
        width: 100%;
        height: 100%;
        object-fit: contain;
        background: #000;
      }

      .vp-big-play {
        position: absolute;
        width: 80px;
        height: 80px;
        border-radius: var(--radius-full);
        background: hsla(195, 100%, 50%, 0.3);
        border: 2px solid var(--color-primary);
        color: white;
        font-size: 2rem;
        display: flex;
        align-items: center;
        justify-content: center;
        cursor: pointer;
        transition: all var(--transition-normal);
        backdrop-filter: blur(20px);
        opacity: 1;
        pointer-events: auto;
      }

      .vp-big-play:hover {
        transform: scale(1.15);
        background: hsla(195, 100%, 50%, 0.5);
        box-shadow: 0 0 40px var(--color-primary-glow);
      }

      .vp-big-play.hidden {
        opacity: 0;
        pointer-events: none;
        transform: scale(0.8);
      }

      /* ── Controls Bar ── */
      .vp-controls {
        position: absolute;
        bottom: 0;
        left: 0;
        right: 0;
        z-index: 10;
        padding: var(--space-md) var(--space-xl);
        background: linear-gradient(to top, hsla(0,0%,0%,0.85), transparent);
        opacity: 0;
        transition: opacity var(--transition-normal);
      }

      .vp-controls.visible { opacity: 1; }

      /* ── Progress Bar ── */
      .vp-progress-container {
        position: relative;
        height: 20px;
        display: flex;
        align-items: center;
        cursor: pointer;
        margin-bottom: var(--space-sm);
      }

      .vp-progress-track {
        width: 100%;
        height: 4px;
        background: hsla(0,0%,100%,0.15);
        border-radius: var(--radius-full);
        position: relative;
        transition: height var(--transition-fast);
      }

      .vp-progress-container:hover .vp-progress-track {
        height: 8px;
      }

      .vp-progress-buffered {
        position: absolute;
        top: 0;
        left: 0;
        height: 100%;
        background: hsla(0,0%,100%,0.15);
        border-radius: var(--radius-full);
      }

      .vp-progress-fill {
        position: absolute;
        top: 0;
        left: 0;
        height: 100%;
        background: var(--color-primary);
        border-radius: var(--radius-full);
        transition: width 0.1s linear;
        box-shadow: 0 0 8px var(--color-primary-glow);
      }

      .vp-progress-thumb {
        position: absolute;
        top: 50%;
        width: 14px;
        height: 14px;
        border-radius: var(--radius-full);
        background: var(--color-primary);
        border: 2px solid white;
        transform: translate(-50%, -50%);
        opacity: 0;
        transition: opacity var(--transition-fast);
        box-shadow: 0 0 10px var(--color-primary-glow);
      }

      .vp-progress-container:hover .vp-progress-thumb {
        opacity: 1;
      }

      .vp-progress-tooltip {
        position: absolute;
        bottom: 100%;
        padding: 2px 8px;
        border-radius: var(--radius-sm);
        background: hsla(0,0%,0%,0.9);
        color: white;
        font-family: var(--font-mono);
        font-size: 0.7rem;
        pointer-events: none;
        opacity: 0;
        transform: translateX(-50%);
        transition: opacity var(--transition-fast);
        white-space: nowrap;
        margin-bottom: 8px;
      }

      .vp-progress-container:hover .vp-progress-tooltip {
        opacity: 1;
      }

      /* ── Control Buttons ── */
      .vp-control-row {
        display: flex;
        align-items: center;
        gap: var(--space-md);
      }

      .vp-ctrl-btn {
        background: none;
        border: none;
        color: white;
        font-size: 1.2rem;
        cursor: pointer;
        padding: 4px;
        transition: all var(--transition-fast);
        opacity: 0.85;
      }

      .vp-ctrl-btn:hover {
        opacity: 1;
        transform: scale(1.15);
      }

      .vp-time {
        font-family: var(--font-mono);
        font-size: 0.75rem;
        color: hsla(0,0%,100%,0.7);
        min-width: 100px;
      }

      .vp-spacer { flex: 1; }

      /* ── Volume ── */
      .vp-volume-group {
        display: flex;
        align-items: center;
        gap: var(--space-sm);
      }

      .vp-volume-slider {
        width: 80px;
        height: 4px;
        -webkit-appearance: none;
        appearance: none;
        background: hsla(0,0%,100%,0.2);
        border-radius: var(--radius-full);
        outline: none;
        cursor: pointer;
      }

      .vp-volume-slider::-webkit-slider-thumb {
        -webkit-appearance: none;
        width: 12px;
        height: 12px;
        border-radius: 50%;
        background: white;
        cursor: pointer;
      }

      /* ── Loading ── */
      .vp-loading {
        position: absolute;
        width: 48px;
        height: 48px;
        border: 3px solid hsla(0,0%,100%,0.1);
        border-top-color: var(--color-primary);
        border-radius: 50%;
        animation: spin 0.8s linear infinite;
        display: none;
      }

      .vp-loading.active { display: block; }

      /* ── Keyboard Hint ── */
      .vp-hint {
        position: absolute;
        top: 50%;
        left: 50%;
        transform: translate(-50%, -50%);
        font-family: var(--font-mono);
        font-size: 2rem;
        color: white;
        opacity: 0;
        pointer-events: none;
        transition: opacity 0.15s ease;
      }

      .vp-hint.flash {
        opacity: 0.8;
        animation: fadeIn 0.15s ease reverse both 0.5s;
      }
    </style>

    <div class="vp-topbar" id="vp-topbar">
      <button class="vp-back" id="vp-close" title="Close player">✕</button>
      <span class="vp-title">${title}</span>
    </div>

    <div class="vp-video-container" id="vp-container">
      <video class="vp-video" id="vp-video" preload="metadata">
        <source src="${streamUrl}" type="video/mp4" />
      </video>

      <button class="vp-big-play" id="vp-big-play">▶</button>
      <div class="vp-loading" id="vp-loading"></div>
      <div class="vp-hint" id="vp-hint"></div>
    </div>

    <div class="vp-controls" id="vp-controls">
      <div class="vp-progress-container" id="vp-progress">
        <div class="vp-progress-track">
          <div class="vp-progress-buffered" id="vp-buffered"></div>
          <div class="vp-progress-fill" id="vp-fill"></div>
        </div>
        <div class="vp-progress-thumb" id="vp-thumb"></div>
        <div class="vp-progress-tooltip" id="vp-tooltip">0:00</div>
      </div>

      <div class="vp-control-row">
        <button class="vp-ctrl-btn" id="vp-playpause" title="Play/Pause">▶</button>
        <button class="vp-ctrl-btn" id="vp-skip-back" title="Back 10s">⏪</button>
        <button class="vp-ctrl-btn" id="vp-skip-fwd" title="Forward 10s">⏩</button>
        <span class="vp-time" id="vp-time">0:00 / 0:00</span>

        <div class="vp-spacer"></div>

        <div class="vp-volume-group">
          <button class="vp-ctrl-btn" id="vp-mute" title="Mute">🔊</button>
          <input type="range" class="vp-volume-slider" id="vp-volume" min="0" max="1" step="0.05" value="1" />
        </div>

        <button class="vp-ctrl-btn" id="vp-pip" title="Picture-in-Picture">📌</button>
        <button class="vp-ctrl-btn" id="vp-fullscreen" title="Fullscreen">⛶</button>
      </div>
    </div>
  `;

  document.body.appendChild(overlay);

  // ── Wire up controls ────────────────────────────────────────────
  const video = document.getElementById('vp-video');
  const bigPlay = document.getElementById('vp-big-play');
  const loading = document.getElementById('vp-loading');
  const playPause = document.getElementById('vp-playpause');
  const fill = document.getElementById('vp-fill');
  const thumb = document.getElementById('vp-thumb');
  const buffered = document.getElementById('vp-buffered');
  const timeDisplay = document.getElementById('vp-time');
  const progressBar = document.getElementById('vp-progress');
  const tooltip = document.getElementById('vp-tooltip');
  const topbar = document.getElementById('vp-topbar');
  const controls = document.getElementById('vp-controls');
  const hint = document.getElementById('vp-hint');
  const container = document.getElementById('vp-container');

  let hideTimeout;

  function showControls() {
    topbar.classList.add('visible');
    controls.classList.add('visible');
    document.body.style.cursor = '';
    clearTimeout(hideTimeout);
    hideTimeout = setTimeout(hideControls, 3000);
  }

  function hideControls() {
    if (!video.paused) {
      topbar.classList.remove('visible');
      controls.classList.remove('visible');
      document.body.style.cursor = 'none';
    }
  }

  function formatTime(s) {
    if (isNaN(s)) return '0:00';
    const h = Math.floor(s / 3600);
    const m = Math.floor((s % 3600) / 60);
    const sec = Math.floor(s % 60);
    if (h > 0) return `${h}:${m.toString().padStart(2, '0')}:${sec.toString().padStart(2, '0')}`;
    return `${m}:${sec.toString().padStart(2, '0')}`;
  }

  function flashHint(text) {
    hint.textContent = text;
    hint.classList.remove('flash');
    void hint.offsetWidth; // Force reflow
    hint.classList.add('flash');
  }

  function togglePlay() {
    if (video.paused) {
      video.play();
    } else {
      video.pause();
    }
  }

  // Mouse activity
  overlay.addEventListener('mousemove', showControls);
  overlay.addEventListener('click', (e) => {
    if (e.target === container || e.target === video) {
      togglePlay();
    }
  });

  // Big play button
  bigPlay.addEventListener('click', (e) => {
    e.stopPropagation();
    video.play();
  });

  // Play/pause button
  playPause.addEventListener('click', togglePlay);

  // Video events
  video.addEventListener('play', () => {
    playPause.textContent = '⏸';
    bigPlay.classList.add('hidden');
    showControls();
  });

  video.addEventListener('pause', () => {
    playPause.textContent = '▶';
    bigPlay.classList.remove('hidden');
    showControls();
    clearTimeout(hideTimeout);
  });

  video.addEventListener('waiting', () => loading.classList.add('active'));
  video.addEventListener('canplay', () => loading.classList.remove('active'));
  video.addEventListener('playing', () => loading.classList.remove('active'));

  // Time update
  video.addEventListener('timeupdate', () => {
    if (!video.duration) return;
    const pct = (video.currentTime / video.duration) * 100;
    fill.style.width = `${pct}%`;
    thumb.style.left = `${pct}%`;
    timeDisplay.textContent = `${formatTime(video.currentTime)} / ${formatTime(video.duration)}`;
  });

  // Buffered
  video.addEventListener('progress', () => {
    if (video.buffered.length > 0) {
      const end = video.buffered.end(video.buffered.length - 1);
      buffered.style.width = `${(end / video.duration) * 100}%`;
    }
  });

  // Progress bar click/drag
  let isDragging = false;

  function seekTo(e) {
    const rect = progressBar.getBoundingClientRect();
    const pct = Math.max(0, Math.min(1, (e.clientX - rect.left) / rect.width));
    video.currentTime = pct * video.duration;
  }

  progressBar.addEventListener('mousedown', (e) => {
    isDragging = true;
    seekTo(e);
  });

  document.addEventListener('mousemove', (e) => {
    if (isDragging) seekTo(e);

    // Tooltip
    const rect = progressBar.getBoundingClientRect();
    if (e.clientX >= rect.left && e.clientX <= rect.right) {
      const pct = (e.clientX - rect.left) / rect.width;
      tooltip.textContent = formatTime(pct * (video.duration || 0));
      tooltip.style.left = `${pct * 100}%`;
    }
  });

  document.addEventListener('mouseup', () => { isDragging = false; });

  // Skip
  document.getElementById('vp-skip-back').addEventListener('click', () => {
    video.currentTime = Math.max(0, video.currentTime - 10);
    flashHint('−10s');
  });

  document.getElementById('vp-skip-fwd').addEventListener('click', () => {
    video.currentTime = Math.min(video.duration, video.currentTime + 10);
    flashHint('+10s');
  });

  // Volume
  const volumeSlider = document.getElementById('vp-volume');
  const muteBtn = document.getElementById('vp-mute');

  volumeSlider.addEventListener('input', () => {
    video.volume = parseFloat(volumeSlider.value);
    video.muted = false;
    muteBtn.textContent = video.volume === 0 ? '🔇' : video.volume < 0.5 ? '🔉' : '🔊';
  });

  muteBtn.addEventListener('click', () => {
    video.muted = !video.muted;
    muteBtn.textContent = video.muted ? '🔇' : '🔊';
    volumeSlider.value = video.muted ? 0 : video.volume;
  });

  // Picture-in-Picture
  document.getElementById('vp-pip')?.addEventListener('click', () => {
    if (document.pictureInPictureElement) {
      document.exitPictureInPicture();
    } else if (video.requestPictureInPicture) {
      video.requestPictureInPicture();
    }
  });

  // Fullscreen
  document.getElementById('vp-fullscreen')?.addEventListener('click', () => {
    if (document.fullscreenElement) {
      document.exitFullscreen();
    } else {
      overlay.requestFullscreen();
    }
  });

  // Close
  document.getElementById('vp-close').addEventListener('click', () => {
    video.pause();
    video.src = '';
    overlay.remove();
    document.body.style.cursor = '';
  });

  // Keyboard shortcuts
  function handleKey(e) {
    if (!document.getElementById('video-player-overlay')) {
      document.removeEventListener('keydown', handleKey);
      return;
    }

    switch (e.key) {
      case ' ':
      case 'k':
        e.preventDefault();
        togglePlay();
        break;
      case 'ArrowLeft':
        e.preventDefault();
        video.currentTime = Math.max(0, video.currentTime - 10);
        flashHint('−10s');
        break;
      case 'ArrowRight':
        e.preventDefault();
        video.currentTime = Math.min(video.duration, video.currentTime + 10);
        flashHint('+10s');
        break;
      case 'ArrowUp':
        e.preventDefault();
        video.volume = Math.min(1, video.volume + 0.1);
        volumeSlider.value = video.volume;
        flashHint(`🔊 ${Math.round(video.volume * 100)}%`);
        break;
      case 'ArrowDown':
        e.preventDefault();
        video.volume = Math.max(0, video.volume - 0.1);
        volumeSlider.value = video.volume;
        flashHint(`🔉 ${Math.round(video.volume * 100)}%`);
        break;
      case 'f':
        if (document.fullscreenElement) {
          document.exitFullscreen();
        } else {
          overlay.requestFullscreen();
        }
        break;
      case 'm':
        video.muted = !video.muted;
        muteBtn.textContent = video.muted ? '🔇' : '🔊';
        flashHint(video.muted ? '🔇 Muted' : '🔊 Unmuted');
        break;
      case 'Escape':
        video.pause();
        video.src = '';
        overlay.remove();
        document.body.style.cursor = '';
        break;
    }
  }

  document.addEventListener('keydown', handleKey);

  // Initial controls show
  showControls();

  return overlay;
}
