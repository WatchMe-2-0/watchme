// ═══════════════════════════════════════════════════════════════════
// WATCHME — Client-Side SPA Router
// Hash-based routing for simplicity (no server config needed)
// ═══════════════════════════════════════════════════════════════════

const routes = {};
let currentPage = null;
let appContainer = null;

export function registerRoute(path, renderFn) {
  routes[path] = renderFn;
}

export function navigate(path) {
  window.location.hash = path;
}

export function getCurrentPath() {
  return window.location.hash.slice(1) || '/';
}

export function initRouter(containerId = 'app') {
  appContainer = document.getElementById(containerId);

  window.addEventListener('hashchange', () => handleRoute());
  handleRoute();
}

async function handleRoute() {
  const path = getCurrentPath();

  // Find matching route (supports params like /movie/:id)
  let matchedRoute = null;
  let params = {};

  for (const [pattern, renderFn] of Object.entries(routes)) {
    const match = matchPath(pattern, path);
    if (match) {
      matchedRoute = renderFn;
      params = match;
      break;
    }
  }

  if (!matchedRoute) {
    // 404 fallback
    matchedRoute = routes['/'] || (() => '<h1>Page not found</h1>');
    params = {};
  }

  // Page transition
  if (currentPage) {
    appContainer.classList.add('page-exit');
    await sleep(200);
    appContainer.classList.remove('page-exit');
  }

  // Render new page
  const content = await matchedRoute(params);

  if (typeof content === 'string') {
    appContainer.innerHTML = content;
  } else if (content instanceof HTMLElement) {
    appContainer.innerHTML = '';
    appContainer.appendChild(content);
  }

  appContainer.classList.add('page-enter');
  currentPage = path;

  // Remove animation class after it plays
  setTimeout(() => {
    appContainer.classList.remove('page-enter');
  }, 400);

  // Scroll to top
  window.scrollTo(0, 0);
}

function matchPath(pattern, path) {
  const patternParts = pattern.split('/');
  const pathParts = path.split('/');

  if (patternParts.length !== pathParts.length) return null;

  const params = {};

  for (let i = 0; i < patternParts.length; i++) {
    if (patternParts[i].startsWith(':')) {
      params[patternParts[i].slice(1)] = pathParts[i];
    } else if (patternParts[i] !== pathParts[i]) {
      return null;
    }
  }

  return params;
}

function sleep(ms) {
  return new Promise(resolve => setTimeout(resolve, ms));
}
