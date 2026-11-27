/* lumen — a calm one-touch game.
 * You are a small light circling a glowing orb. Tap to let go; you fly off
 * along the tangent. Drift into the next orb's ring to be caught. Each catch
 * plays a note, and the sky slowly travels from dusk to dawn.
 *
 * Plain Canvas 2D + Web Audio, no dependencies, no build step.
 */
(() => {
  'use strict';

  // ---------------------------------------------------------------- utils
  const TAU = Math.PI * 2;
  const clamp = (v, a, b) => (v < a ? a : v > b ? b : v);
  const lerp = (a, b, t) => a + (b - a) * t;
  const rand = (a, b) => a + Math.random() * (b - a);
  const pick = (arr) => arr[(Math.random() * arr.length) | 0];
  const damp = (a, b, lambda, dt) => lerp(a, b, 1 - Math.exp(-lambda * dt));
  const easeOutCubic = (t) => 1 - Math.pow(1 - t, 3);
  const easeOutBack = (t) => { const c1 = 1.70158, c3 = c1 + 1; return 1 + c3 * Math.pow(t - 1, 3) + c1 * Math.pow(t - 1, 2); };

  function hexToRgb(hex) {
    const n = parseInt(hex.slice(1), 16);
    return [(n >> 16) & 255, (n >> 8) & 255, n & 255];
  }
  function mixRgb(a, b, t) {
    return [lerp(a[0], b[0], t), lerp(a[1], b[1], t), lerp(a[2], b[2], t)];
  }
  const rgbStr = (c, a = 1) => `rgba(${c[0] | 0},${c[1] | 0},${c[2] | 0},${a})`;

  const reduceMotion = window.matchMedia && window.matchMedia('(prefers-reduced-motion: reduce)').matches;

  // ---------------------------------------------------------------- storage
  const STORE_KEY = 'lumen.v1';
  const Store = {
    data: { best: 0, bestZen: 0, dust: 0, light: 0, muted: false, mode: 'flow', plays: 0, perfects: 0 },
    load() {
      try {
        const raw = localStorage.getItem(STORE_KEY);
        if (raw) Object.assign(this.data, JSON.parse(raw));
      } catch (_) { /* private mode or corrupted: keep defaults */ }
    },
    save() {
      try { localStorage.setItem(STORE_KEY, JSON.stringify(this.data)); } catch (_) { /* ignore */ }
    },
  };
  Store.load();

  // ---------------------------------------------------------------- content
  // The sky travels through these as you climb. Every SKY_SPAN lights is one sky.
  const SKY_SPAN = 18;
  const SKIES = [
    { name: 'dusk',     top: '#23305e', mid: '#6b4f8f', bot: '#e9967a', stars: 0.35, aurora: 0, orbs: ['#ffd6a5', '#ffb8a8', '#f8c8dc'] },
    { name: 'twilight', top: '#141c45', mid: '#3b3f8c', bot: '#a274b4', stars: 0.6,  aurora: 0, orbs: ['#cdb9ff', '#a9c8ff', '#ffc9e3'] },
    { name: 'night',    top: '#050a22', mid: '#0f2257', bot: '#21507e', stars: 1.0,  aurora: 0, orbs: ['#a3dcff', '#b8f2e6', '#d4dcff'] },
    { name: 'aurora',   top: '#02101c', mid: '#06384a', bot: '#0e6466', stars: 0.85, aurora: 1, orbs: ['#82ffd2', '#a2f0ff', '#c6ffad'] },
    { name: 'deep',     top: '#0c041f', mid: '#2a0f4b', bot: '#5a2680', stars: 0.9,  aurora: 0.35, orbs: ['#ffa3f2', '#d8a6ff', '#a3b6ff'] },
    { name: 'dawn',     top: '#2f3c70', mid: '#c27a9c', bot: '#ffcf9f', stars: 0.2,  aurora: 0, orbs: ['#fff2c4', '#ffd3a3', '#ffbab4'] },
  ].map((s) => ({ ...s, topC: hexToRgb(s.top), midC: hexToRgb(s.mid), botC: hexToRgb(s.bot) }));

  // Lights unlock as you gather stardust over all your runs.
  const LIGHTS = [
    { name: 'Ember', color: '#ffe2a8', cost: 0 },
    { name: 'Moon',  color: '#e6efff', cost: 60 },
    { name: 'Jade',  color: '#a6ffd8', cost: 180 },
    { name: 'Rose',  color: '#ffb6c9', cost: 400 },
    { name: 'Iris',  color: '#c6b4ff', cost: 800 },
    { name: 'Sun',   color: '#ffcf5e', cost: 1500 },
  ];
  const unlockedCount = (dust) => LIGHTS.filter((l) => dust >= l.cost).length;

  const TIPS = [
    'Aim for the heart of a light for a perfect catch.',
    'Perfect catches are worth two.',
    'The sky changes every few lights. Dawn is waiting.',
    'Stardust gathers over every journey and wakes new lights.',
    'Moving lights reward patience. Wait for them.',
    'Zen mode has no endings, if you need a quiet one.',
    'Every catch plays the next note. Listen.',
  ];

  // ---------------------------------------------------------------- tuning
  const STEP = 1 / 120;          // fixed simulation step
  const MAX_FRAME = 0.25;        // clamp after a stall
  const FLY_SPEED = 470;         // world units / s
  const LOST_TIME = 2.4;         // seconds of flight before the light is lost
  const COL = 165;               // half-width of the column orbs live in
  const MOTE_PICKUP = 24;

  // ---------------------------------------------------------------- canvas
  const canvas = document.getElementById('game');
  const ctx = canvas.getContext('2d', { alpha: false });
  let W = 0, H = 0, DPR = 1, scale = 1, viewW = 0, viewH = 0;
  let vignette = null;

  function resize() {
    W = window.innerWidth;
    H = window.innerHeight;
    DPR = Math.min(window.devicePixelRatio || 1, 2);
    canvas.width = Math.round(W * DPR);
    canvas.height = Math.round(H * DPR);
    scale = Math.min(W / 420, H / 680);
    viewW = W / scale;
    viewH = H / scale;
    buildVignette();
    buildStars();
  }

  function buildVignette() {
    vignette = document.createElement('canvas');
    vignette.width = Math.max(1, Math.round(W / 2));
    vignette.height = Math.max(1, Math.round(H / 2));
    const v = vignette.getContext('2d');
    const g = v.createRadialGradient(vignette.width / 2, vignette.height * 0.45, Math.min(vignette.width, vignette.height) * 0.25,
      vignette.width / 2, vignette.height / 2, Math.hypot(vignette.width, vignette.height) * 0.62);
    g.addColorStop(0, 'rgba(0,0,0,0)');
    g.addColorStop(1, 'rgba(4,4,20,0.55)');
    v.fillStyle = g;
    v.fillRect(0, 0, vignette.width, vignette.height);
  }

  // Pre-rendered glow sprites: much cheaper than shadowBlur on phones.
  const glowCache = new Map();
  function glow(color) {
    let c = glowCache.get(color);
    if (c) return c;
    const S = 128;
    c = document.createElement('canvas');
    c.width = c.height = S;
    const g = c.getContext('2d');
    const rgb = hexToRgb(color);
    const grad = g.createRadialGradient(S / 2, S / 2, 0, S / 2, S / 2, S / 2);
    grad.addColorStop(0, 'rgba(255,255,255,1)');
    grad.addColorStop(0.12, rgbStr(rgb, 0.95));
    grad.addColorStop(0.32, rgbStr(rgb, 0.38));
    grad.addColorStop(0.6, rgbStr(rgb, 0.1));
    grad.addColorStop(1, rgbStr(rgb, 0));
    g.fillStyle = grad;
    g.fillRect(0, 0, S, S);
    glowCache.set(color, c);
    return c;
  }

  // ---------------------------------------------------------------- sound
  const Sound = (() => {
    let ac = null, master = null, bus = null, padTimer = 0, chord = 0, noiseBuf = null;
    let muted = !!Store.data.muted;
    const BASE = 146.83; // D3
    const PENTA = [0, 2, 4, 7, 9]; // D major pentatonic: nothing can clash
    const freq = (semi) => BASE * Math.pow(2, semi / 12);
    const scaleNote = (i) => {
      const o = Math.floor(i / 5);
      const d = PENTA[((i % 5) + 5) % 5];
      return freq(12 * o + d);
    };
    const CHORDS = [[0, 7, 16, 21], [-3, 4, 12, 19], [-10, -3, 7, 14], [-5, 4, 9, 14]];

    function impulse(sec, decay) {
      const rate = ac.sampleRate, len = (rate * sec) | 0;
      const buf = ac.createBuffer(2, len, rate);
      for (let ch = 0; ch < 2; ch++) {
        const d = buf.getChannelData(ch);
        for (let i = 0; i < len; i++) d[i] = (Math.random() * 2 - 1) * Math.pow(1 - i / len, decay);
      }
      return buf;
    }

    function init() {
      if (ac) return;
      const AC = window.AudioContext || window.webkitAudioContext;
      if (!AC) return;
      try { ac = new AC(); } catch (_) { ac = null; return; }
      master = ac.createGain();
      master.gain.value = muted ? 0 : 0.85;
      const comp = ac.createDynamicsCompressor();
      comp.threshold.value = -16;
      comp.ratio.value = 3;
      master.connect(comp);
      comp.connect(ac.destination);
      bus = ac.createGain();
      bus.connect(master);
      const verb = ac.createConvolver();
      verb.buffer = impulse(3.4, 2.4);
      const wet = ac.createGain();
      wet.gain.value = 0.5;
      bus.connect(verb);
      verb.connect(wet);
      wet.connect(master);
      noiseBuf = ac.createBuffer(1, ac.sampleRate, ac.sampleRate);
      const nd = noiseBuf.getChannelData(0);
      for (let i = 0; i < nd.length; i++) nd[i] = Math.random() * 2 - 1;
      playChord();
      padTimer = setInterval(playChord, 9000);
    }

    function resume() { if (ac && ac.state !== 'running') ac.resume().catch(() => {}); }
    function suspend() { if (ac && ac.state === 'running') ac.suspend().catch(() => {}); }

    function playChord() {
      if (!ac || ac.state !== 'running') return;
      const t = ac.currentTime;
      const lp = ac.createBiquadFilter();
      lp.type = 'lowpass';
      lp.frequency.value = 900;
      lp.Q.value = 0.4;
      const g = ac.createGain();
      g.gain.setValueAtTime(0.0001, t);
      g.gain.linearRampToValueAtTime(1, t + 3.5);
      g.gain.setValueAtTime(1, t + 8.5);
      g.gain.linearRampToValueAtTime(0.0001, t + 14);
      lp.connect(g);
      g.connect(bus);
      for (const semi of CHORDS[chord % CHORDS.length]) {
        for (const [type, det, vol] of [['sine', 0, 0.03], ['triangle', 7, 0.012]]) {
          const o = ac.createOscillator();
          o.type = type;
          o.frequency.value = freq(semi);
          o.detune.value = det + rand(-3, 3);
          const og = ac.createGain();
          og.gain.value = vol;
          o.connect(og);
          og.connect(lp);
          o.start(t);
          o.stop(t + 14.2);
        }
      }
      chord++;
    }

    function bell(f, vol, t, len = 2.6) {
      const o = ac.createOscillator();
      o.type = 'sine';
      o.frequency.value = f;
      const m = ac.createOscillator();
      m.type = 'sine';
      m.frequency.value = f * 3.5;
      const mg = ac.createGain();
      mg.gain.setValueAtTime(f * 0.7, t);
      mg.gain.exponentialRampToValueAtTime(1, t + 0.8);
      m.connect(mg);
      mg.connect(o.frequency);
      const g = ac.createGain();
      g.gain.setValueAtTime(0.0001, t);
      g.gain.exponentialRampToValueAtTime(vol, t + 0.01);
      g.gain.exponentialRampToValueAtTime(0.0001, t + len);
      o.connect(g);
      g.connect(bus);
      o.start(t); m.start(t);
      o.stop(t + len + 0.05); m.stop(t + len + 0.05);
    }

    const ready = () => ac && !muted && ac.state === 'running';

    return {
      init, resume, suspend,
      get muted() { return muted; },
      setMuted(m) {
        muted = m;
        if (master) master.gain.setTargetAtTime(m ? 0 : 0.85, ac.currentTime, 0.08);
      },
      catchNote(i, perfect) {
        if (!ready()) return;
        const t = ac.currentTime;
        bell(scaleNote(i), 0.2, t);
        if (perfect) {
          bell(scaleNote(i + 2), 0.11, t + 0.07);
          bell(scaleNote(i + 5), 0.08, t + 0.15, 3.2);
        }
      },
      mote(i) {
        if (!ready()) return;
        bell(scaleNote(14 + (i % 5)), 0.045, ac.currentTime, 0.9);
      },
      launch() {
        if (!ready()) return;
        const t = ac.currentTime;
        const src = ac.createBufferSource();
        src.buffer = noiseBuf;
        const bp = ac.createBiquadFilter();
        bp.type = 'bandpass';
        bp.Q.value = 1.2;
        bp.frequency.setValueAtTime(500, t);
        bp.frequency.exponentialRampToValueAtTime(1600, t + 0.28);
        const g = ac.createGain();
        g.gain.setValueAtTime(0.0001, t);
        g.gain.exponentialRampToValueAtTime(0.06, t + 0.04);
        g.gain.exponentialRampToValueAtTime(0.0001, t + 0.34);
        src.connect(bp); bp.connect(g); g.connect(bus);
        src.start(t, Math.random() * 0.5);
        src.stop(t + 0.4);
      },
      lost() {
        if (!ready()) return;
        const t = ac.currentTime;
        bell(scaleNote(4), 0.1, t, 3);
        bell(scaleNote(2), 0.09, t + 0.35, 3);
        bell(scaleNote(-1), 0.1, t + 0.7, 4);
      },
      tick() {
        if (!ready()) return;
        bell(scaleNote(10), 0.04, ac.currentTime, 0.6);
      },
    };
  })();

  // ---------------------------------------------------------------- state
  const S = {
    scene: 'menu',      // menu | play | over | paused
    mode: Store.data.mode === 'zen' ? 'zen' : 'flow',
    t: 0,               // simulation time
    timeScale: 1,
    slowUntil: 0,       // real-time ms when the slow breath ends
    score: 0,
    maxIdx: 0,
    noteIdx: 5,
    perfectStreak: 0,
    runDust: 0,
    runPerfects: 0,
    skyT: 0,            // displayed sky position (smoothed)
    trauma: 0,
    shakeT: 0,
    tutorial: false,
    hintShown: '',
    camStartY: 0,
  };

  const cam = { x: 0, y: 0, px: 0, py: 0 };
  const player = { state: 'orbit', orb: null, from: null, angle: 0, angVel: 0, rr: 0, x: 0, y: 0, px: 0, py: 0, vx: 0, vy: 0, flyT: 0, lostT: 0, alpha: 1, trail: [], pop: 0 };
  let orbs = [];
  let motes = [];
  const particles = [];
  const ripples = [];
  const popups = [];
  let orbCount = 0;

  // ---------------------------------------------------------------- world gen
  function difficulty(idx) { return clamp(idx / 70, 0, 1); }

  function makeOrb(x, y, idx, opts = {}) {
    const d = difficulty(idx);
    const sky = SKIES[Math.floor(idx / SKY_SPAN) % SKIES.length];
    const r = opts.r ?? lerp(60, 38, d) * rand(0.9, 1.1);
    const o = {
      idx, x, y, baseX: x,
      r, baseR: r,
      core: r * 0.3,
      w: opts.w ?? lerp(1.75, 2.85, d) * rand(0.9, 1.1),
      dir: Math.random() < 0.5 ? 1 : -1,
      color: pick(sky.orbs),
      phase: rand(0, TAU),
      moveAmp: 0, moveFreq: 0,
      breathe: 0,
      visited: false,
      pulse: 0,
      spin: rand(0, TAU),
    };
    if (!opts.still && idx > 9 && Math.random() < lerp(0.15, 0.55, clamp((idx - 10) / 50, 0, 1))) {
      o.moveAmp = Math.min(rand(28, 75), COL + 20 - Math.abs(x));
      o.moveFreq = rand(0.35, 0.7);
      if (o.moveAmp < 15) o.moveAmp = 0;
    }
    if (!opts.still && idx > 28 && Math.random() < 0.22) o.breathe = rand(0.1, 0.18);
    return o;
  }

  function spawnNext() {
    const prev = orbs[orbs.length - 1];
    const idx = orbCount++;
    const d = difficulty(idx);
    let x = 0, y = 0;
    for (let tries = 0; tries < 12; tries++) {
      x = rand(-COL, COL);
      const dx = Math.abs(x - prev.baseX);
      if (dx > 28 && dx < 240) break;
    }
    y = prev.y - (lerp(180, 245, d) + rand(-22, 22));
    const o = makeOrb(x, y, idx);
    // Keep ring edges comfortably apart.
    const gap = Math.hypot(o.x - prev.baseX, o.y - prev.y) - o.r - prev.r;
    if (gap < 70) o.y -= 70 - gap;
    orbs.push(o);
    // Stardust along the likely path.
    const n = Math.random() < 0.75 ? ((Math.random() * 3) | 0) + 1 : 0;
    for (let i = 0; i < n; i++) {
      const t = rand(0.32, 0.68);
      motes.push({
        x: lerp(prev.baseX, o.baseX, t) + rand(-36, 36),
        y: lerp(prev.y, o.y, t) + rand(-20, 20),
        ph: rand(0, TAU),
        got: false,
        a: 0,
      });
    }
  }

  function ensureWorld() {
    const top = cam.y - viewH * 1.4;
    while (orbs[orbs.length - 1].y > top) spawnNext();
    const bottom = cam.y + viewH * 1.2;
    // Keep the orb we're on and the one we came from.
    orbs = orbs.filter((o) => o.y < bottom || o === player.orb || o === player.from);
    motes = motes.filter((m) => m.y < bottom && !(m.got && m.a <= 0));
  }

  // ---------------------------------------------------------------- run lifecycle
  function newRun() {
    orbs = [];
    motes = [];
    particles.length = 0;
    ripples.length = 0;
    popups.length = 0;
    orbCount = 0;
    const first = makeOrb(0, 0, orbCount++, { r: 62, w: 1.7, still: true });
    first.visited = true;
    orbs.push(first);
    S.score = 0;
    S.maxIdx = 0;
    S.noteIdx = 5;
    S.perfectStreak = 0;
    S.runDust = 0;
    S.runPerfects = 0;
    S.t = 0;
    S.timeScale = 1;
    S.slowUntil = 0;
    S.trauma = 0;
    S.skyT = 0;
    S.tutorial = Store.data.plays < 2;
    Object.assign(player, { state: 'orbit', orb: first, from: null, angle: Math.PI * 0.75, angVel: first.w * first.dir, rr: first.r, flyT: 0, lostT: 0, alpha: 1, pop: 0 });
    placeOnOrbit();
    player.px = player.x; player.py = player.y;
    player.trail.length = 0;
    cam.x = 0;
    cam.y = first.y - viewH * 0.12;
    cam.px = cam.x; cam.py = cam.y;
    S.camStartY = cam.y;
    ensureWorld();
    updateHud(true);
  }

  function placeOnOrbit() {
    const o = player.orb;
    player.x = o.x + Math.cos(player.angle) * player.rr;
    player.y = o.y + Math.sin(player.angle) * player.rr;
  }

  // ---------------------------------------------------------------- actions
  let inputQueued = false;

  function release() {
    if (player.state !== 'orbit') return;
    const s = Math.sign(player.angVel) || 1;
    const tx = -Math.sin(player.angle) * s;
    const ty = Math.cos(player.angle) * s;
    player.vx = tx * FLY_SPEED;
    player.vy = ty * FLY_SPEED;
    player.from = player.orb;
    player.orb = null;
    player.state = 'fly';
    player.flyT = 0;
    player.pop = 1;
    if (S.scene === 'play') Sound.launch();
    burst(player.x, player.y, 5, '#ffffff', 60, 0.5);
    if (S.tutorial) setHint('');
  }

  function tryCatch(o) {
    const dx = player.x - o.x, dy = player.y - o.y;
    const dist = Math.hypot(dx, dy);
    if (dist > o.r + 4) return false;

    // How straight at the heart were we flying?
    const sp = Math.hypot(player.vx, player.vy) || 1;
    const ux = player.vx / sp, uy = player.vy / sp;
    const perp = Math.abs(ux * (o.y - player.y) - uy * (o.x - player.x));
    const perfect = perp < o.r * 0.3;

    const cross = dx * player.vy - dy * player.vx;
    const dir = Math.abs(cross) < 1e-3 ? o.dir : Math.sign(cross);

    player.state = 'orbit';
    player.orb = o;
    player.angle = Math.atan2(dy, dx);
    player.angVel = dir * o.w;
    player.rr = dist;
    player.flyT = 0;
    o.pulse = 1;

    const fresh = o.idx > S.maxIdx;
    if (fresh) {
      S.maxIdx = o.idx;
      o.visited = true;
      const pts = perfect ? 2 : 1;
      S.score += pts;
      if (perfect) { S.perfectStreak++; S.runPerfects++; } else S.perfectStreak = 0;
      S.noteIdx += perfect ? 2 : 1;
      if (S.noteIdx > 15) S.noteIdx -= 10;
      Sound.catchNote(S.noteIdx, perfect);
      if (S.scene === 'play') popups.push({ x: o.x, y: o.y - o.r - 18, text: perfect ? (S.perfectStreak >= 3 ? `flawless ×${S.perfectStreak}` : 'perfect') : '+1', life: 1.3, max: 1.3, big: perfect });
      if (perfect && S.scene === 'play') popups.push({ x: o.x, y: o.y - o.r - 40, text: '+2', life: 1.3, max: 1.3, big: false });
      bumpScore();
      if (S.tutorial && S.score >= 2 && S.score < 4) setHint('Aim for the heart for a perfect');
      else if (S.tutorial && S.score >= 5) { setHint(''); S.tutorial = false; }
    } else {
      Sound.tick();
    }

    // Feedback bundle (calm tiers: no harsh shake, a slow breath on perfects).
    ripples.push({ x: o.x, y: o.y, r: o.r, max: o.r * (perfect ? 2.6 : 1.8), life: perfect ? 1.1 : 0.8, maxLife: perfect ? 1.1 : 0.8, color: o.color });
    burst(player.x, player.y, perfect ? 22 : 9, o.color, perfect ? 150 : 90, perfect ? 1.1 : 0.8);
    if (perfect) {
      ripples.push({ x: o.x, y: o.y, r: o.core, max: o.r * 1.4, life: 0.7, maxLife: 0.7, color: '#ffffff' });
      if (!reduceMotion) {
        addTrauma(0.22);
        S.timeScale = 0.3;
        S.slowUntil = performance.now() + 140;
      }
      if (navigator.vibrate && (!navigator.userActivation || navigator.userActivation.hasBeenActive)) { try { navigator.vibrate(12); } catch (_) { /* ignore */ } }
    } else if (!reduceMotion) {
      addTrauma(0.08);
    }
    return true;
  }

  function loseLight() {
    player.state = 'lost';
    player.lostT = 0;
    S.perfectStreak = 0;
    Sound.lost();
    burst(player.x, player.y, 14, LIGHTS[Store.data.light].color, 50, 1.4);
  }

  function respawnZen() {
    const o = player.from && orbs.includes(player.from) ? player.from : orbs.find((q) => q.idx === S.maxIdx) || orbs[0];
    Object.assign(player, { state: 'orbit', orb: o, from: null, angle: -Math.PI / 2, angVel: o.w * o.dir, rr: o.r, alpha: 0, flyT: 0 });
    placeOnOrbit();
    player.px = player.x; player.py = player.y;
    player.trail.length = 0;
    ripples.push({ x: o.x, y: o.y, r: o.core, max: o.r * 1.5, life: 0.9, maxLife: 0.9, color: o.color });
  }

  function addTrauma(a) { S.trauma = clamp(S.trauma + a, 0, 1); }

  function burst(x, y, n, color, speed, life) {
    for (let i = 0; i < n; i++) {
      const a = rand(0, TAU), v = rand(0.25, 1) * speed;
      particles.push({ x, y, vx: Math.cos(a) * v, vy: Math.sin(a) * v, life: life * rand(0.6, 1), max: life, size: rand(5, 11), color });
    }
    if (particles.length > 400) particles.splice(0, particles.length - 400);
  }

  // ---------------------------------------------------------------- update
  function update(dt) {
    S.t += dt;

    for (const o of orbs) {
      if (o.moveAmp) o.x = o.baseX + Math.sin(S.t * o.moveFreq * TAU * 0.25 + o.phase) * o.moveAmp;
      if (o.breathe) o.r = o.baseR * (1 + Math.sin(S.t * 1.3 + o.phase) * o.breathe);
      o.pulse = Math.max(0, o.pulse - dt * 1.8);
      o.spin += dt * o.w * 0.15;
    }

    player.px = player.x; player.py = player.y;
    player.pop = Math.max(0, player.pop - dt * 3);

    if (player.state === 'orbit') {
      const o = player.orb;
      player.angle += player.angVel * dt;
      player.rr = damp(player.rr, o.r, 10, dt);
      placeOnOrbit();
      if (player.alpha < 1) player.alpha = Math.min(1, player.alpha + dt * 2.5);
    } else if (player.state === 'fly') {
      player.x += player.vx * dt;
      player.y += player.vy * dt;
      player.flyT += dt;
      for (const o of orbs) {
        if (o === player.from) continue;
        if (tryCatch(o)) break;
      }
      if (player.state === 'fly') {
        const offX = Math.abs(player.x - cam.x) > viewW / 2 + 40;
        const offY = player.y > cam.y + viewH / 2 + 40 || player.y < cam.y - viewH / 2 - 160;
        if (player.flyT > LOST_TIME || offX || offY) loseLight();
      }
    } else if (player.state === 'lost') {
      player.lostT += dt;
      player.vx *= Math.pow(0.12, dt);
      player.vy *= Math.pow(0.12, dt);
      player.x += player.vx * dt;
      player.y += player.vy * dt;
      player.alpha = Math.max(0, 1 - player.lostT / 0.9);
      if (player.lostT > 1.1) {
        if (S.scene === 'menu') newRun();
        else if (S.scene === 'play') {
          if (S.mode === 'zen') respawnZen();
          else endRun();
        }
      }
    }

    // trail
    if (player.state !== 'lost' || player.alpha > 0.05) {
      player.trail.push(player.x, player.y);
      if (player.trail.length > 64) player.trail.splice(0, 2);
    } else if (player.trail.length) player.trail.splice(0, 2);

    // stardust
    for (let i = 0; i < motes.length; i++) {
      const m = motes[i];
      if (m.got) { m.a -= dt * 3; continue; }
      m.a = Math.min(1, m.a + dt * 1.5);
      const my = m.y + Math.sin(S.t * 1.2 + m.ph) * 5;
      if (player.state !== 'lost' && Math.hypot(player.x - m.x, player.y - my) < MOTE_PICKUP) {
        m.got = true;
        S.runDust++;
        Sound.mote(S.runDust);
        burst(m.x, my, 6, '#fff4d6', 70, 0.6);
        updateHud();
      }
    }

    for (let i = particles.length - 1; i >= 0; i--) {
      const p = particles[i];
      p.life -= dt;
      if (p.life <= 0) { particles.splice(i, 1); continue; }
      p.x += p.vx * dt; p.y += p.vy * dt;
      p.vx *= Math.pow(0.15, dt); p.vy *= Math.pow(0.15, dt);
    }
    for (let i = ripples.length - 1; i >= 0; i--) {
      const r = ripples[i];
      r.life -= dt;
      if (r.life <= 0) ripples.splice(i, 1);
    }
    for (let i = popups.length - 1; i >= 0; i--) {
      const p = popups[i];
      p.life -= dt;
      p.y -= dt * 22;
      if (p.life <= 0) popups.splice(i, 1);
    }

    // camera
    cam.px = cam.x; cam.py = cam.y;
    let tx, ty;
    if (player.state === 'orbit') { tx = player.orb.x * 0.35; ty = player.orb.y - viewH * 0.12; }
    else { tx = player.x * 0.35; ty = Math.min(player.y, (player.from ? player.from.y : player.y)) - viewH * 0.12; }
    cam.x = damp(cam.x, tx, 2.2, dt);
    cam.y = damp(cam.y, ty, player.state === 'fly' ? 2.6 : 2.0, dt);

    ensureWorld();
  }

  // ---------------------------------------------------------------- render
  let stars = [];
  function buildStars() {
    const n = Math.round(clamp((W * H) / 7000, 60, 220));
    stars = [];
    for (let i = 0; i < n; i++) {
      stars.push({ x: Math.random() * W, y: Math.random() * H, s: Math.random() < 0.12 ? rand(1.4, 2.2) : rand(0.6, 1.3), tw: rand(0.5, 2), ph: rand(0, TAU), depth: rand(0.02, 0.08) });
    }
  }
  const bokeh = Array.from({ length: 14 }, () => ({ x: Math.random(), y: Math.random(), r: rand(30, 90), depth: rand(0.15, 0.35), ph: rand(0, TAU) }));

  function skyAt(t) {
    const i = Math.floor(t), f = t - i;
    const a = SKIES[((i % SKIES.length) + SKIES.length) % SKIES.length];
    const b = SKIES[(((i + 1) % SKIES.length) + SKIES.length) % SKIES.length];
    // ease the blend so each sky lingers
    const e = f < 0.7 ? 0 : (f - 0.7) / 0.3;
    const k = e * e * (3 - 2 * e);
    return {
      top: mixRgb(a.topC, b.topC, k),
      mid: mixRgb(a.midC, b.midC, k),
      bot: mixRgb(a.botC, b.botC, k),
      stars: lerp(a.stars, b.stars, k),
      aurora: lerp(a.aurora, b.aurora, k),
    };
  }

  function render(alpha, realT) {
    const camX = lerp(cam.px, cam.x, alpha);
    const camY = lerp(cam.py, cam.y, alpha);
    const sky = skyAt(S.skyT);

    ctx.setTransform(DPR, 0, 0, DPR, 0, 0);
    ctx.globalCompositeOperation = 'source-over';
    ctx.globalAlpha = 1;

    // sky
    const g = ctx.createLinearGradient(0, 0, 0, H);
    g.addColorStop(0, rgbStr(sky.top));
    g.addColorStop(0.58, rgbStr(sky.mid));
    g.addColorStop(1, rgbStr(sky.bot));
    ctx.fillStyle = g;
    ctx.fillRect(0, 0, W, H);

    const rise = (S.camStartY - camY) * scale; // pixels climbed

    // stars
    ctx.fillStyle = '#fff';
    for (const st of stars) {
      const y = ((st.y + rise * st.depth) % H + H) % H;
      const a = sky.stars * (0.45 + 0.55 * Math.sin(realT * st.tw + st.ph));
      if (a <= 0.02) continue;
      ctx.globalAlpha = a;
      ctx.fillRect(st.x, y, st.s, st.s);
    }
    ctx.globalAlpha = 1;

    // aurora ribbons
    if (sky.aurora > 0.01) drawAurora(sky.aurora, realT, rise);

    // bokeh
    ctx.globalCompositeOperation = 'lighter';
    for (const b of bokeh) {
      const y = (((b.y * H + rise * b.depth) % (H + 200)) + H + 200) % (H + 200) - 100;
      const x = b.x * W + Math.sin(realT * 0.1 + b.ph) * 20;
      ctx.globalAlpha = 0.05 + 0.03 * Math.sin(realT * 0.3 + b.ph);
      const s = b.r * 2;
      ctx.drawImage(glow('#ffffff'), x - s, y - s, s * 2, s * 2);
    }
    ctx.globalAlpha = 1;
    ctx.globalCompositeOperation = 'source-over';

    // hills at the start of the journey
    drawHills(sky, rise);

    // world transform
    let shx = 0, shy = 0, rot = 0;
    if (S.trauma > 0) {
      const sh = S.trauma * S.trauma;
      shx = 10 * sh * Math.sin(S.shakeT * 1.7);
      shy = 8 * sh * Math.sin(S.shakeT * 2.3);
      rot = 0.012 * sh * Math.sin(S.shakeT * 1.1);
    }
    ctx.setTransform(DPR, 0, 0, DPR, 0, 0);
    ctx.translate(W / 2 + shx, H / 2 + shy);
    ctx.rotate(rot);
    ctx.scale(scale, scale);
    ctx.translate(-camX, -camY);
    const worldA = S.scene === 'menu' ? 0.45 : S.scene === 'over' ? 0.5 : 1;
    S.worldA = damp(S.worldA ?? worldA, worldA, 4, 1 / 60);

    const target = orbs.find((o) => o.idx === S.maxIdx + 1);

    // aim guide
    const showGuide = S.scene === 'play' && player.state === 'orbit' && (S.mode === 'zen' || S.tutorial);
    if (showGuide) {
      const s = Math.sign(player.angVel) || 1;
      const tx = -Math.sin(player.angle) * s, ty = Math.cos(player.angle) * s;
      ctx.fillStyle = '#fff';
      for (let i = 1; i <= 14; i++) {
        const d = i * 14;
        ctx.globalAlpha = 0.32 * (1 - i / 15);
        ctx.beginPath();
        ctx.arc(player.x + tx * d, player.y + ty * d, 1.6, 0, TAU);
        ctx.fill();
      }
      ctx.globalAlpha = 1;
    }

    // stardust
    ctx.globalCompositeOperation = 'lighter';
    for (const m of motes) {
      if (m.a <= 0) continue;
      const y = m.y + Math.sin(realT * 1.2 + m.ph) * 5;
      const s = (m.got ? 22 + (1 - m.a) * 20 : 16 + Math.sin(realT * 3 + m.ph) * 2);
      ctx.globalAlpha = m.a * 0.95 * S.worldA;
      ctx.drawImage(glow('#fff1cf'), m.x - s, y - s, s * 2, s * 2);
    }

    // orbs
    for (const o of orbs) {
      const isTarget = o === target;
      const isHome = o === player.orb;
      const dim = o.visited && !isHome ? 0.45 : 1;
      const pulse = easeOutCubic(o.pulse);
      const rgb = hexToRgb(o.color);

      // ring
      ctx.globalCompositeOperation = 'source-over';
      const ringA = (isTarget ? 0.55 + 0.15 * Math.sin(realT * 2.4) : isHome ? 0.5 : 0.28) * dim;
      ctx.strokeStyle = rgbStr(rgb, ringA * S.worldA);
      ctx.lineWidth = 1.4 + pulse * 2;
      ctx.beginPath();
      ctx.arc(o.x, o.y, o.r * (1 + pulse * 0.08), 0, TAU);
      ctx.stroke();

      // little satellites drifting along the ring
      ctx.fillStyle = rgbStr(rgb, 0.7 * dim * S.worldA);
      for (let k = 0; k < 3; k++) {
        const a = o.spin * o.dir + (k * TAU) / 3;
        ctx.beginPath();
        ctx.arc(o.x + Math.cos(a) * o.r, o.y + Math.sin(a) * o.r, 1.6, 0, TAU);
        ctx.fill();
      }

      // core
      ctx.globalCompositeOperation = 'lighter';
      const breathe = 1 + Math.sin(realT * 1.6 + o.phase) * 0.06;
      const squash = 1 + pulse * 0.35 * (1 - pulse);
      const cs = o.core * 3.4 * breathe * (1 + pulse * 0.25) * squash;
      ctx.globalAlpha = 0.9 * dim * S.worldA;
      ctx.drawImage(glow(o.color), o.x - cs, o.y - cs, cs * 2, cs * 2);
      ctx.globalAlpha = 1;
    }

    // ripples
    ctx.globalCompositeOperation = 'lighter';
    for (const r of ripples) {
      const k = 1 - r.life / r.maxLife;
      const e = easeOutCubic(k);
      ctx.strokeStyle = rgbStr(hexToRgb(r.color), (1 - k) * 0.55);
      ctx.lineWidth = 2 * (1 - k) + 0.5;
      ctx.beginPath();
      ctx.arc(r.x, r.y, lerp(r.r, r.max, e), 0, TAU);
      ctx.stroke();
    }

    // particles
    for (const p of particles) {
      const k = p.life / p.max;
      const s = p.size * (0.4 + 0.6 * k);
      ctx.globalAlpha = k * 0.9;
      ctx.drawImage(glow(p.color), p.x - s, p.y - s, s * 2, s * 2);
    }
    ctx.globalAlpha = 1;

    // player trail + light
    const lightColor = LIGHTS[Store.data.light].color;
    const lrgb = hexToRgb(lightColor);
    const tr = player.trail;
    if (tr.length >= 4) {
      ctx.lineCap = 'round';
      const n = tr.length / 2;
      for (let i = 1; i < n; i++) {
        const k = i / n;
        ctx.strokeStyle = rgbStr(lrgb, k * 0.45 * player.alpha);
        ctx.lineWidth = 1 + k * 5;
        ctx.beginPath();
        ctx.moveTo(tr[(i - 1) * 2], tr[(i - 1) * 2 + 1]);
        ctx.lineTo(tr[i * 2], tr[i * 2 + 1]);
        ctx.stroke();
      }
    }
    if (player.alpha > 0.01) {
      const px = lerp(player.px, player.x, alpha);
      const py = lerp(player.py, player.y, alpha);
      const ps = 30 * (1 + player.pop * 0.35) * (1 + Math.sin(realT * 4) * 0.04);
      ctx.globalAlpha = player.alpha;
      ctx.drawImage(glow(lightColor), px - ps, py - ps, ps * 2, ps * 2);
      ctx.globalCompositeOperation = 'source-over';
      ctx.fillStyle = '#fffdf6';
      ctx.beginPath();
      ctx.arc(px, py, 4.2, 0, TAU);
      ctx.fill();
      ctx.globalAlpha = 1;
    }

    // popups
    ctx.globalCompositeOperation = 'source-over';
    ctx.textAlign = 'center';
    ctx.textBaseline = 'middle';
    for (const p of popups) {
      const k = p.life / p.max;
      const appear = easeOutBack(clamp((1 - k) * 5, 0, 1));
      ctx.globalAlpha = Math.min(1, k * 2.2);
      ctx.fillStyle = '#fffaf0';
      ctx.font = `italic ${p.big ? 22 : 17}px "Iowan Old Style","Palatino Linotype",Palatino,"Book Antiqua",Georgia,serif`;
      ctx.save();
      ctx.translate(p.x, p.y);
      ctx.scale(appear, appear);
      ctx.fillText(p.text, 0, 0);
      ctx.restore();
    }
    ctx.globalAlpha = 1;

    // vignette
    ctx.setTransform(DPR, 0, 0, DPR, 0, 0);
    ctx.drawImage(vignette, 0, 0, W, H);
  }

  function drawAurora(strength, t, rise) {
    ctx.globalCompositeOperation = 'lighter';
    const bands = [
      { y: 0.22, amp: 46, len: 0.0042, sp: 0.18, h: 150, c: [110, 255, 200] },
      { y: 0.34, amp: 36, len: 0.0058, sp: -0.13, h: 110, c: [120, 210, 255] },
      { y: 0.16, amp: 30, len: 0.0031, sp: 0.09, h: 90, c: [190, 150, 255] },
    ];
    for (const b of bands) {
      const baseY = b.y * H + (rise * 0.04) % 40;
      const grad = ctx.createLinearGradient(0, baseY - b.h, 0, baseY + 30);
      grad.addColorStop(0, rgbStr(b.c, 0));
      grad.addColorStop(0.7, rgbStr(b.c, 0.16 * strength));
      grad.addColorStop(1, rgbStr(b.c, 0));
      ctx.fillStyle = grad;
      ctx.beginPath();
      const step = 18;
      ctx.moveTo(0, H);
      for (let x = 0; x <= W + step; x += step) {
        const y = baseY + Math.sin(x * b.len + t * b.sp) * b.amp + Math.sin(x * b.len * 2.3 - t * b.sp * 1.7) * b.amp * 0.35;
        if (x === 0) ctx.moveTo(x, y + 30);
        ctx.lineTo(x, y - b.h);
      }
      for (let x = W + step; x >= 0; x -= step) {
        const y = baseY + Math.sin(x * b.len + t * b.sp) * b.amp + Math.sin(x * b.len * 2.3 - t * b.sp * 1.7) * b.amp * 0.35;
        ctx.lineTo(x, y + 30);
      }
      ctx.closePath();
      ctx.fill();
    }
    ctx.globalCompositeOperation = 'source-over';
  }

  function drawHills(sky, rise) {
    const layers = [
      { p: 0.35, base: 0.8, amp: 0.07, seed: 1.3, shade: 0.55 },
      { p: 0.6, base: 0.88, amp: 0.05, seed: 4.1, shade: 0.75 },
    ];
    const dark = [12, 10, 32];
    for (const L of layers) {
      const y0 = H * L.base + rise * L.p;
      if (y0 - H * L.amp * 2 > H) continue;
      const col = mixRgb(sky.mid, dark, L.shade);
      ctx.fillStyle = rgbStr(col);
      ctx.beginPath();
      ctx.moveTo(0, H);
      const step = 12;
      for (let x = 0; x <= W + step; x += step) {
        const u = x / Math.max(W, 1);
        const h = Math.sin(u * 5.1 + L.seed) * 0.6 + Math.sin(u * 11.7 + L.seed * 2) * 0.25 + Math.sin(u * 2.3 + L.seed * 3) * 0.8;
        ctx.lineTo(x, y0 - h * H * L.amp);
      }
      ctx.lineTo(W, H);
      ctx.closePath();
      ctx.fill();
    }
  }

  // ---------------------------------------------------------------- HUD / DOM
  const $ = (id) => document.getElementById(id);
  const el = {
    hud: $('hud'), score: $('score'), dust: $('dustCount'), hint: $('hint'),
    menu: $('menu'), over: $('over'), pause: $('pause'),
    play: $('playBtn'), sound: $('soundBtn'), stats: $('stats'),
    lightRow: $('lightRow'), lightName: $('lightName'),
    overKicker: $('overKicker'), overScore: $('overScore'), overBest: $('overBest'), overDust: $('overDust'), overUnlock: $('overUnlock'), overTip: $('overTip'),
    again: $('againBtn'), home: $('homeBtn'), resume: $('resumeBtn'), quit: $('quitBtn'), pauseBtn: $('pauseBtn'),
  };
  let lastHud = { score: -1, dust: -1 };

  function updateHud(force) {
    if (force || lastHud.score !== S.score) { el.score.textContent = S.score; lastHud.score = S.score; }
    if (force || lastHud.dust !== S.runDust) { el.dust.textContent = S.runDust; lastHud.dust = S.runDust; }
  }
  function bumpScore() {
    updateHud();
    el.score.classList.add('bump');
    setTimeout(() => el.score.classList.remove('bump'), 70);
  }
  function setHint(text) {
    if (S.hintShown === text) return;
    S.hintShown = text;
    if (text) { el.hint.textContent = text; el.hint.classList.add('show'); }
    else el.hint.classList.remove('show');
  }

  const SOUND_ON = '<svg viewBox="0 0 24 24" width="22" height="22" aria-hidden="true"><path d="M4 9.5h3.2L12 5.5v13l-4.8-4H4z"/><path d="M15.5 8.8a4.4 4.4 0 0 1 0 6.4M18 6.4a7.8 7.8 0 0 1 0 11.2" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round"/></svg>';
  const SOUND_OFF = '<svg viewBox="0 0 24 24" width="22" height="22" aria-hidden="true"><path d="M4 9.5h3.2L12 5.5v13l-4.8-4H4z"/><path d="M16 9.5l5 5M21 9.5l-5 5" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round"/></svg>';
  function renderSoundBtn() {
    el.sound.innerHTML = Sound.muted ? SOUND_OFF : SOUND_ON;
    el.sound.setAttribute('aria-label', Sound.muted ? 'Sound off' : 'Sound on');
  }

  function applyLightColor() {
    document.documentElement.style.setProperty('--light', LIGHTS[Store.data.light].color);
  }

  function renderMenu() {
    for (const b of document.querySelectorAll('.mode')) b.setAttribute('aria-checked', String(b.dataset.mode === S.mode));
    const dust = Store.data.dust;
    el.lightRow.innerHTML = '';
    LIGHTS.forEach((l, i) => {
      const b = document.createElement('button');
      const locked = dust < l.cost;
      b.className = 'light-dot' + (locked ? ' locked' : '');
      b.style.setProperty('--c', l.color);
      b.setAttribute('aria-pressed', String(i === Store.data.light));
      b.setAttribute('aria-label', locked ? `${l.name}, wakes at ${l.cost} stardust` : l.name);
      b.innerHTML = '<i></i>';
      b.addEventListener('click', () => {
        Sound.init(); Sound.resume();
        if (locked) { el.lightName.textContent = `${l.name} wakes at ${l.cost} stardust`; return; }
        Store.data.light = i; Store.save();
        applyLightColor();
        Sound.tick();
        renderMenu();
      });
      b.addEventListener('pointerenter', () => { el.lightName.textContent = locked ? `${l.name} wakes at ${l.cost} stardust` : l.name; });
      el.lightRow.appendChild(b);
    });
    el.lightName.textContent = LIGHTS[Store.data.light].name;
    const best = S.mode === 'zen' ? Store.data.bestZen : Store.data.best;
    el.stats.innerHTML = `<span>Best ${best}</span><span>Stardust ${dust}</span>`;
    renderSoundBtn();
    applyLightColor();
  }

  function show(screen) {
    for (const s of [el.menu, el.over, el.pause]) s.classList.add('hidden');
    if (screen) screen.classList.remove('hidden');
    el.hud.classList.toggle('hidden', !(S.scene === 'play' || S.scene === 'paused'));
  }

  // ---------------------------------------------------------------- scenes
  let overShownAt = 0;

  function startGame() {
    Sound.init();
    Sound.resume();
    S.scene = 'play';
    newRun();
    show(null);
    if (S.tutorial) setHint('Tap anywhere to let go'); else setHint('');
  }

  function commitRun() {
    const beforeDust = Store.data.dust;
    const beforeLights = unlockedCount(beforeDust);
    Store.data.dust += S.runDust;
    Store.data.plays++;
    Store.data.perfects += S.runPerfects;
    let newBest = false;
    if (S.mode === 'flow' && S.score > Store.data.best) { Store.data.best = S.score; newBest = S.score > 0; }
    if (S.mode === 'zen' && S.score > Store.data.bestZen) { Store.data.bestZen = S.score; }
    const afterLights = unlockedCount(Store.data.dust);
    Store.save();
    S.runDust = 0;
    S.runPerfects = 0;
    return { newBest, unlocked: afterLights > beforeLights ? LIGHTS.slice(beforeLights, afterLights) : [] };
  }

  function endRun() {
    const score = S.score, dust = S.runDust;
    const { newBest, unlocked } = commitRun();
    S.scene = 'over';
    setHint('');
    el.overKicker.textContent = score >= 2 * SKY_SPAN ? 'What a journey' : score >= SKY_SPAN ? 'You saw a new sky' : 'The light rests';
    el.overScore.textContent = score;
    el.overBest.textContent = newBest ? 'A new best' : `Best ${Store.data.best}`;
    el.overBest.classList.toggle('new', newBest);
    el.overDust.textContent = dust > 0 ? `+${dust} stardust, ${Store.data.dust} in all` : `${Store.data.dust} stardust in all`;
    if (unlocked.length) {
      el.overUnlock.textContent = `A new light wakes: ${unlocked.map((l) => l.name).join(', ')}`;
      el.overUnlock.classList.remove('hidden');
    } else el.overUnlock.classList.add('hidden');
    el.overTip.textContent = nextTip();
    overShownAt = performance.now();
    show(el.over);
    setTimeout(() => { if (S.scene === 'over') el.again.focus({ preventScroll: true }); }, 50);
  }

  let tipI = (Math.random() * TIPS.length) | 0;
  function nextTip() {
    const lightsLeft = LIGHTS.find((l) => l.cost > Store.data.dust);
    if (lightsLeft && Math.random() < 0.35) return `${lightsLeft.cost - Store.data.dust} more stardust wakes ${lightsLeft.name}.`;
    tipI = (tipI + 1) % TIPS.length;
    return TIPS[tipI];
  }

  function goHome() {
    if (S.scene === 'play' || S.scene === 'paused') commitRun();
    S.scene = 'menu';
    setHint('');
    newRun(); // fresh scenery behind the menu
    renderMenu();
    show(el.menu);
  }

  function pauseGame() {
    if (S.scene !== 'play') return;
    S.scene = 'paused';
    show(el.pause);
    el.resume.focus({ preventScroll: true });
  }
  function resumeGame() {
    if (S.scene !== 'paused') return;
    S.scene = 'play';
    Sound.resume();
    show(null);
    last = performance.now() / 1000;
  }

  // ---------------------------------------------------------------- input
  function onPress(e) {
    if (S.scene === 'play') {
      if (e) e.preventDefault();
      Sound.resume();
      inputQueued = true;
    } else if (S.scene === 'over' && performance.now() - overShownAt > 650) {
      startGame();
    }
  }

  canvas.addEventListener('pointerdown', (e) => {
    if (e.pointerType === 'mouse' && e.button !== 0) return;
    onPress(e);
  });
  el.over.addEventListener('pointerdown', (e) => {
    if (e.target.closest('button')) return;
    onPress(e);
  });
  window.addEventListener('keydown', (e) => {
    if (e.repeat) return;
    const k = e.key;
    if (k === ' ' || k === 'Enter' || k === 'ArrowUp' || k === 'w' || k === 'W') {
      if (S.scene === 'play') { e.preventDefault(); onPress(); }
      else if (S.scene === 'menu' && k === ' ' && document.activeElement === document.body) { e.preventDefault(); startGame(); }
    } else if (k === 'Escape' || k === 'p' || k === 'P') {
      if (S.scene === 'play') pauseGame();
      else if (S.scene === 'paused') resumeGame();
    } else if (k === 'm' || k === 'M') {
      toggleSound();
    }
  });

  function toggleSound() {
    Sound.init();
    Sound.setMuted(!Sound.muted);
    Store.data.muted = Sound.muted;
    Store.save();
    renderSoundBtn();
    Sound.resume();
  }

  el.play.addEventListener('click', startGame);
  el.again.addEventListener('click', startGame);
  el.home.addEventListener('click', goHome);
  el.quit.addEventListener('click', goHome);
  el.resume.addEventListener('click', resumeGame);
  el.pauseBtn.addEventListener('pointerdown', (e) => e.stopPropagation());
  el.pauseBtn.addEventListener('click', pauseGame);
  el.sound.addEventListener('click', toggleSound);
  for (const b of document.querySelectorAll('.mode')) {
    b.addEventListener('click', () => {
      S.mode = b.dataset.mode;
      Store.data.mode = S.mode;
      Store.save();
      Sound.init(); Sound.resume(); Sound.tick();
      renderMenu();
    });
  }

  document.addEventListener('visibilitychange', () => {
    if (document.hidden) { pauseGame(); Sound.suspend(); }
    else if (S.scene !== 'paused') Sound.resume();
  });
  window.addEventListener('resize', resize);
  window.addEventListener('orientationchange', () => setTimeout(resize, 200));
  // No pinch-zoom or double-tap zoom on iOS.
  document.addEventListener('gesturestart', (e) => e.preventDefault());
  document.addEventListener('dblclick', (e) => e.preventDefault());

  // ---------------------------------------------------------------- loop
  let acc = 0;
  let last = performance.now() / 1000;

  function frame(nowMs) {
    const now = nowMs / 1000;
    const realDt = Math.min(now - last, MAX_FRAME);
    last = now;

    if (S.scene !== 'paused') {
      // The slow breath after a perfect runs in real time so it always ends.
      if (S.timeScale < 1 && nowMs >= S.slowUntil) S.timeScale = Math.min(1, S.timeScale + realDt * 6);
      if (S.timeScale > 1) S.timeScale = 1;

      acc += realDt;
      while (acc >= STEP) {
        if (inputQueued && S.scene === 'play') { inputQueued = false; release(); }
        if (S.scene === 'menu') demoPilot();
        update(STEP * S.timeScale);
        acc -= STEP;
      }
      // feedback decays in real time
      S.trauma = Math.max(0, S.trauma - realDt * 1.6);
      S.shakeT += realDt * 30;
      S.skyT = damp(S.skyT, S.score / SKY_SPAN, 0.8, realDt);
    }

    render(acc / STEP, now);
    requestAnimationFrame(frame);
  }

  // Behind the home screen the light drifts on its own, as a living preview.
  function demoPilot() {
    if (player.state !== 'orbit') return;
    const next = orbs.find((o) => o.idx === S.maxIdx + 1);
    if (!next) return;
    const s = Math.sign(player.angVel) || 1;
    const tx = -Math.sin(player.angle) * s, ty = Math.cos(player.angle) * s;
    const dx = next.x - player.x, dy = next.y - player.y;
    const d = Math.hypot(dx, dy);
    const dot = (tx * dx + ty * dy) / d;
    const perp = Math.abs(tx * dy - ty * dx);
    if (dot > 0.9 && perp < next.r * 0.25 && player.rr > player.orb.r - 1) {
      release(); // same rules as the player; score is hidden behind the menu
    }
  }

  // ---------------------------------------------------------------- boot
  resize();
  newRun();
  renderMenu();
  show(el.menu);
  requestAnimationFrame(frame);

  // Debug hooks for automated tests (harmless in play).
  window.__lumen = { S, player, get orbs() { return orbs; }, release: () => { inputQueued = true; } };
})();
