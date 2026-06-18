// ═══════════════════════════════════════════════════════════════════
// WATCHME — Aurora Borealis GLSL Shader Background
// GPU-accelerated fullscreen aurora effect via Three.js
// ═══════════════════════════════════════════════════════════════════

import * as THREE from 'three';

let renderer, scene, camera, material, mesh;
let animationId;
let mouseX = 0, mouseY = 0;
let startTime = Date.now();

const vertexShader = `
  varying vec2 vUv;
  void main() {
    vUv = uv;
    gl_Position = vec4(position, 1.0);
  }
`;

const fragmentShader = `
  precision highp float;

  uniform float uTime;
  uniform vec2 uResolution;
  uniform vec2 uMouse;
  uniform vec3 uColor1;
  uniform vec3 uColor2;
  uniform vec3 uColor3;
  uniform float uIntensity;

  varying vec2 vUv;

  // Simplex-like noise
  float hash(vec2 p) {
    return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5453);
  }

  float noise(vec2 p) {
    vec2 i = floor(p);
    vec2 f = fract(p);
    f = f * f * (3.0 - 2.0 * f);

    float a = hash(i);
    float b = hash(i + vec2(1.0, 0.0));
    float c = hash(i + vec2(0.0, 1.0));
    float d = hash(i + vec2(1.0, 1.0));

    return mix(mix(a, b, f.x), mix(c, d, f.x), f.y);
  }

  float fbm(vec2 p) {
    float value = 0.0;
    float amplitude = 0.5;
    float frequency = 1.0;

    for (int i = 0; i < 5; i++) {
      value += amplitude * noise(p * frequency);
      frequency *= 2.0;
      amplitude *= 0.5;
    }

    return value;
  }

  void main() {
    vec2 uv = vUv;
    float time = uTime * 0.15;

    // Subtle mouse parallax
    vec2 mouseInfluence = (uMouse - 0.5) * 0.05;
    uv += mouseInfluence;

    // Create aurora wave layers
    float n1 = fbm(vec2(uv.x * 3.0 + time * 0.5, uv.y * 1.5 + time * 0.3));
    float n2 = fbm(vec2(uv.x * 2.0 - time * 0.4, uv.y * 2.0 + time * 0.2));
    float n3 = fbm(vec2(uv.x * 4.0 + time * 0.3, uv.y * 1.0 - time * 0.5));

    // Aurora curtain effect — concentrated in upper portion
    float curtain = smoothstep(0.3, 0.8, uv.y);
    curtain *= smoothstep(1.0, 0.6, uv.y);

    // Wavering vertical bands
    float wave1 = sin(uv.x * 6.0 + time * 2.0 + n1 * 4.0) * 0.5 + 0.5;
    float wave2 = sin(uv.x * 4.0 - time * 1.5 + n2 * 3.0) * 0.5 + 0.5;
    float wave3 = sin(uv.x * 8.0 + time * 1.0 + n3 * 5.0) * 0.5 + 0.5;

    // Combine waves with noise
    float aurora = (wave1 * n1 + wave2 * n2 * 0.8 + wave3 * n3 * 0.6) * curtain;
    aurora = pow(aurora, 1.5) * uIntensity;

    // Color mixing — three-color gradient
    vec3 color = mix(uColor1, uColor2, wave1);
    color = mix(color, uColor3, wave2 * 0.5);

    // Star field in background
    float stars = step(0.998, hash(floor(uv * 400.0)));
    float starTwinkle = sin(time * 5.0 + hash(floor(uv * 400.0)) * 100.0) * 0.5 + 0.5;
    vec3 starColor = vec3(stars * starTwinkle * 0.6);

    // Final composition
    vec3 bg = vec3(0.02, 0.02, 0.03);
    vec3 finalColor = bg + starColor + color * aurora;

    // Subtle vignette
    float vignette = 1.0 - length((uv - 0.5) * 1.2);
    vignette = smoothstep(0.0, 0.7, vignette);
    finalColor *= vignette;

    gl_FragColor = vec4(finalColor, 1.0);
  }
`;

// Theme color presets for the aurora
const themeColors = {
  default: {
    color1: new THREE.Color(0.0, 0.75, 1.0),   // Electric blue
    color2: new THREE.Color(0.0, 1.0, 0.5),     // Cyan-green
    color3: new THREE.Color(0.3, 0.0, 0.8),     // Deep purple
    intensity: 0.8,
  },
  night: {
    color1: new THREE.Color(0.2, 0.4, 0.9),     // Soft blue
    color2: new THREE.Color(0.1, 0.6, 0.7),     // Muted teal
    color3: new THREE.Color(0.15, 0.1, 0.5),    // Dark indigo
    intensity: 0.5,
  },
  dracula: {
    color1: new THREE.Color(0.7, 0.4, 1.0),     // Purple
    color2: new THREE.Color(1.0, 0.2, 0.3),     // Red
    color3: new THREE.Color(0.3, 0.1, 0.6),     // Deep purple
    intensity: 0.7,
  },
};

export function initAurora(canvasId = 'aurora-canvas') {
  const canvas = document.getElementById(canvasId);
  if (!canvas) return;

  // Check for WebGL support
  try {
    renderer = new THREE.WebGLRenderer({
      canvas,
      alpha: true,
      antialias: false,
      powerPreference: 'low-power',
    });
  } catch (e) {
    console.warn('WebGL not available, falling back to CSS gradient');
    canvas.style.background = 'linear-gradient(180deg, #050510, #0a1020, #050505)';
    return;
  }

  renderer.setSize(window.innerWidth, window.innerHeight);
  renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2)); // Cap at 2x for perf

  scene = new THREE.Scene();
  camera = new THREE.OrthographicCamera(-1, 1, 1, -1, 0, 1);

  material = new THREE.ShaderMaterial({
    vertexShader,
    fragmentShader,
    uniforms: {
      uTime: { value: 0 },
      uResolution: { value: new THREE.Vector2(window.innerWidth, window.innerHeight) },
      uMouse: { value: new THREE.Vector2(0.5, 0.5) },
      uColor1: { value: themeColors.default.color1 },
      uColor2: { value: themeColors.default.color2 },
      uColor3: { value: themeColors.default.color3 },
      uIntensity: { value: themeColors.default.intensity },
    },
  });

  const geometry = new THREE.PlaneGeometry(2, 2);
  mesh = new THREE.Mesh(geometry, material);
  scene.add(mesh);

  // Mouse tracking
  window.addEventListener('mousemove', (e) => {
    mouseX = e.clientX / window.innerWidth;
    mouseY = 1.0 - e.clientY / window.innerHeight;
  });

  // Resize handler
  window.addEventListener('resize', () => {
    renderer.setSize(window.innerWidth, window.innerHeight);
    material.uniforms.uResolution.value.set(window.innerWidth, window.innerHeight);
  });

  animate();
}

function animate() {
  animationId = requestAnimationFrame(animate);

  const elapsed = (Date.now() - startTime) / 1000;
  material.uniforms.uTime.value = elapsed;
  material.uniforms.uMouse.value.set(mouseX, mouseY);

  renderer.render(scene, camera);
}

export function setAuroraTheme(themeName) {
  if (!material || !themeColors[themeName]) return;

  const colors = themeColors[themeName];

  // Smooth transition using lerp
  const lerp = (a, b, t) => a + (b - a) * t;
  const currentUniforms = material.uniforms;

  let t = 0;
  const transitionSpeed = 0.02;

  function transitionStep() {
    t += transitionSpeed;
    if (t >= 1) t = 1;

    currentUniforms.uColor1.value.lerp(colors.color1, transitionSpeed * 2);
    currentUniforms.uColor2.value.lerp(colors.color2, transitionSpeed * 2);
    currentUniforms.uColor3.value.lerp(colors.color3, transitionSpeed * 2);
    currentUniforms.uIntensity.value = lerp(
      currentUniforms.uIntensity.value,
      colors.intensity,
      transitionSpeed * 2
    );

    if (t < 1) requestAnimationFrame(transitionStep);
  }

  transitionStep();
}

export function setAuroraIntensity(intensity) {
  if (material) {
    material.uniforms.uIntensity.value = intensity;
  }
}

export function destroyAurora() {
  if (animationId) cancelAnimationFrame(animationId);
  if (renderer) renderer.dispose();
  if (material) material.dispose();
}
