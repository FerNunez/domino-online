// Singleton sound manager, modeled loosely on lichess's sound.ts: preload,
// play(name), and a persisted mute flag. Kept as a plain module (not a
// React hook) so it can be called directly from non-component code, like
// the onmessage closure in useGameConnection — see useSound.ts for the
// thin hook wrapper components should prefer.
import { SOUNDS, SoundName } from "./sounds";

const MUTE_STORAGE_KEY = "domino:sound-muted";

// Safari still needs the webkit-prefixed constructor.
type AudioContextCtor = typeof AudioContext;
function getAudioContextCtor(): AudioContextCtor | undefined {
  if (typeof window === "undefined") return undefined;
  return (
    window.AudioContext ??
    (window as unknown as { webkitAudioContext?: AudioContextCtor })
      .webkitAudioContext
  );
}

let sharedContext: AudioContext | null = null;

// Created lazily on first play() rather than at module load — browsers block
// AudioContext construction/resume before a user gesture, and the earliest
// guaranteed gesture in this app is the lobby's "Start game" click.
function getAudioContext(): AudioContext | null {
  if (sharedContext) return sharedContext;
  const Ctor = getAudioContextCtor();
  if (!Ctor) return null;
  sharedContext = new Ctor();
  return sharedContext;
}

function readInitialMuted(): boolean {
  if (typeof window === "undefined") return false;
  try {
    return window.localStorage.getItem(MUTE_STORAGE_KEY) === "1";
  } catch {
    return false;
  }
}

let muted = readInitialMuted();
const muteListeners = new Set<(muted: boolean) => void>();

function setMuted(next: boolean) {
  muted = next;
  try {
    window.localStorage.setItem(MUTE_STORAGE_KEY, next ? "1" : "0");
  } catch {
    // localStorage unavailable (e.g. private mode) — mute state just won't
    // persist across reloads, which is fine.
  }
  muteListeners.forEach((listener) => listener(muted));
}

function isMuted(): boolean {
  return muted;
}

function toggleMuted(): boolean {
  setMuted(!muted);
  return muted;
}

function onMuteChange(listener: (muted: boolean) => void): () => void {
  muteListeners.add(listener);
  return () => muteListeners.delete(listener);
}

// Envelope helper: ramps a gain node from 0 -> peak -> 0 so oscillators
// don't click on start/stop.
function envelope(
  ctx: AudioContext,
  gain: GainNode,
  peak: number,
  attack: number,
  duration: number,
  startAt: number,
) {
  const g = gain.gain;
  g.setValueAtTime(0, startAt);
  g.linearRampToValueAtTime(peak, startAt + attack);
  g.exponentialRampToValueAtTime(0.001, startAt + duration);
}

function playTone(
  ctx: AudioContext,
  {
    type,
    freq,
    endFreq,
    startAt,
    duration,
    peak = 0.2,
  }: {
    type: OscillatorType;
    freq: number;
    endFreq?: number;
    startAt: number;
    duration: number;
    peak?: number;
  },
) {
  const osc = ctx.createOscillator();
  const gain = ctx.createGain();
  osc.type = type;
  osc.frequency.setValueAtTime(freq, startAt);
  if (endFreq !== undefined) {
    osc.frequency.linearRampToValueAtTime(endFreq, startAt + duration);
  }
  envelope(ctx, gain, peak, Math.min(0.02, duration / 4), duration, startAt);
  osc.connect(gain);
  gain.connect(ctx.destination);
  osc.start(startAt);
  osc.stop(startAt + duration + 0.05);
}

// Synthesized placeholders — used until real files land in public/sounds/.
// Each call schedules brand-new oscillator/gain nodes (never reused), so
// rapid repeated calls (e.g. two tiles played back to back) layer cleanly
// instead of one cutting the other off.
function synthesizeFallback(ctx: AudioContext, name: SoundName) {
  const now = ctx.currentTime;
  switch (name) {
    case SoundName.GameStart: {
      // Low-to-mid "clang" with a quick decay — bell-ish.
      playTone(ctx, {
        type: "triangle",
        freq: 440,
        startAt: now,
        duration: 0.5,
        peak: 0.25,
      });
      playTone(ctx, {
        type: "square",
        freq: 660,
        startAt: now,
        duration: 0.35,
        peak: 0.12,
      });
      break;
    }
    case SoundName.YourTurn: {
      // Sine sweeping up then down over ~0.3s — whistle-ish.
      const osc = ctx.createOscillator();
      const gain = ctx.createGain();
      osc.type = "sine";
      const up = 0.15;
      const down = 0.15;
      osc.frequency.setValueAtTime(500, now);
      osc.frequency.linearRampToValueAtTime(1100, now + up);
      osc.frequency.linearRampToValueAtTime(500, now + up + down);
      envelope(ctx, gain, 0.2, 0.02, up + down, now);
      osc.connect(gain);
      gain.connect(ctx.destination);
      osc.start(now);
      osc.stop(now + up + down + 0.05);
      break;
    }
    case SoundName.TilePlayed: {
      // Very short high square-wave click/tap.
      playTone(ctx, {
        type: "square",
        freq: 900,
        startAt: now,
        duration: 0.06,
        peak: 0.2,
      });
      break;
    }
    case SoundName.Pass: {
      // Two quick blips at different pitches — a cute "boop-boop".
      playTone(ctx, {
        type: "sine",
        freq: 520,
        startAt: now,
        duration: 0.1,
        peak: 0.2,
      });
      playTone(ctx, {
        type: "sine",
        freq: 700,
        startAt: now + 0.12,
        duration: 0.12,
        peak: 0.2,
      });
      break;
    }
  }
}

// Attempts to play the real asset via HTMLAudioElement; falls back to a
// synthesized tone on any load/decode/playback error (expected today, since
// no real files exist yet in public/sounds/). Each call creates its own
// Audio element / oscillator graph so overlapping calls never cancel each
// other.
function play(name: SoundName) {
  if (muted) return;
  if (typeof window === "undefined") return;

  const def = SOUNDS[name];
  let fellBack = false;
  const fallback = () => {
    if (fellBack) return;
    fellBack = true;
    const ctx = getAudioContext();
    if (ctx) synthesizeFallback(ctx, name);
  };

  try {
    const audio = new Audio(def.src);
    audio.volume = 0.6;
    audio.addEventListener("error", fallback, { once: true });
    const playPromise = audio.play();
    if (playPromise && typeof playPromise.catch === "function") {
      playPromise.catch(fallback);
    }
  } catch {
    fallback();
  }
}

export const soundManager = {
  play,
  isMuted,
  setMuted,
  toggleMuted,
  onMuteChange,
};
