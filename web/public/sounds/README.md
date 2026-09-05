# Sound assets

Drop the real audio files here with these exact filenames — `soundManager.ts`
fetches them by convention (`/sounds/<name>.mp3`):

- `game-start.mp3` — boxing-ring bell, played on `GameEvents.GameStarted`
- `your-turn.mp3` — whistle, played when it becomes your turn
- `tile-played.mp3` — knock/click, played on every `GameEvents.PlayerMoveMade`
- `pass.mp3` — cute blip, played on `GameEvents.PlayerPassed`

Until a file is present (or if it fails to load/decode), `soundManager.ts`
falls back to a synthesized placeholder tone generated live via the Web Audio
API, so the feature is audible with zero assets. Once a real file lands here
under the exact name above, it's used automatically — no code changes needed.
