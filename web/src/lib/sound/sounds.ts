// Sound catalog: maps each named effect to its asset path under
// public/sounds/. Real files are dropped in later (see the README in that
// folder) — soundManager falls back to a synthesized tone per SoundName
// until then, so this list is also what synthesizeFallback switches on.

export enum SoundName {
  GameStart = "game-start",
  YourTurn = "your-turn",
  TilePlayed = "tile-played",
  Pass = "pass",
}

interface SoundDef {
  src: string;
}

export const SOUNDS: Record<SoundName, SoundDef> = {
  [SoundName.GameStart]: { src: "/sounds/game-start.mp3" },
  [SoundName.YourTurn]: { src: "/sounds/your-turn.mp3" },
  [SoundName.TilePlayed]: { src: "/sounds/tile-played.mp3" },
  [SoundName.Pass]: { src: "/sounds/pass.mp3" },
};
