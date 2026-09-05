import { useCallback, useEffect, useState } from "react";
import { soundManager } from "@/lib/sound/soundManager";
import { SoundName } from "@/lib/sound/sounds";

// Thin React wrapper around the soundManager singleton — components should
// use this instead of importing soundManager directly, so mute state stays
// in sync with re-renders. Non-component code (e.g. the WS onmessage
// closure in useGameConnection) calls soundManager.play() directly since
// hooks can't be used there.
export function useSound() {
  const [muted, setMutedState] = useState(() => soundManager.isMuted());

  useEffect(() => soundManager.onMuteChange(setMutedState), []);

  const play = useCallback((name: SoundName) => soundManager.play(name), []);
  const toggleMute = useCallback(() => soundManager.toggleMuted(), []);

  return { play, muted, toggleMute };
}
