import { redirect } from "next/navigation";

// The lobby and game screens share one page (app/lobby/[id]/page.tsx) so the
// WebSocket connection opened while waiting survives into the game without a
// reconnect — see that file's comment. This route exists only so a direct
// link to /game/{id} still lands somewhere sensible; note that arriving here
// after the game has already started means missing the initial
// game.hand_dealt event (no backend endpoint to re-fetch game state yet).
export default async function GamePageRedirect({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  redirect(`/lobby/${id}`);
}
