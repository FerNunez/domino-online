"use client";

import { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { createLobby, ensureGuestToken, joinLobby } from "@/lib/api";

export default function Home() {
  const router = useRouter();
  const [joinID, setJoinID] = useState("");
  const [createDisplayName, setCreateDisplayName] = useState("");
  const [displayName, setDisplayName] = useState("");
  const [loading, setLoading] = useState<"create" | "join" | null>(null);
  const [error, setError] = useState<string | null>(null);

  const handleCreate = async () => {
    setLoading("create");
    setError(null);
    try {
      await ensureGuestToken();
      const { lobbyID } = await createLobby(4, createDisplayName.trim());
      router.push(`/lobby/${lobbyID}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "failed to create lobby");
      setLoading(null);
    }
  };

  const handleJoin = async () => {
    if (!joinID.trim()) return;
    setLoading("join");
    setError(null);
    try {
      await ensureGuestToken();
      const { lobbyID } = await joinLobby(joinID.trim(), displayName.trim());
      router.push(`/lobby/${lobbyID}`);
    } catch (err) {
      setError(err instanceof Error ? err.message : "failed to join lobby");
      setLoading(null);
    }
  };

  return (
    <main className="flex min-h-screen flex-col items-center justify-center gap-6 p-6">
      <h1 className="text-3xl font-bold">Domino</h1>

      <Card className="w-full max-w-sm">
        <CardHeader>
          <CardTitle>Create a lobby</CardTitle>
        </CardHeader>
        <CardContent className="flex flex-col gap-2">
          <input
            className="w-full rounded-md border bg-background px-3 py-2 text-sm"
            placeholder="Your name (optional)"
            value={createDisplayName}
            onChange={(e) => setCreateDisplayName(e.target.value)}
          />
          <Button className="w-full" onClick={handleCreate} disabled={loading !== null}>
            {loading === "create" ? "Creating…" : "New 4-player lobby"}
          </Button>
        </CardContent>
      </Card>

      <Card className="w-full max-w-sm">
        <CardHeader>
          <CardTitle>Join a lobby</CardTitle>
        </CardHeader>
        <CardContent className="flex flex-col gap-2">
          <input
            className="w-full rounded-md border bg-background px-3 py-2 text-sm"
            placeholder="Lobby ID"
            value={joinID}
            onChange={(e) => setJoinID(e.target.value)}
          />
          <input
            className="w-full rounded-md border bg-background px-3 py-2 text-sm"
            placeholder="Your name (optional)"
            value={displayName}
            onChange={(e) => setDisplayName(e.target.value)}
          />
          <Button className="w-full" variant="secondary" onClick={handleJoin} disabled={loading !== null}>
            {loading === "join" ? "Joining…" : "Join"}
          </Button>
        </CardContent>
      </Card>

      {error && <p className="text-sm text-destructive">{error}</p>}

      <Link href="/games" className="text-sm text-muted-foreground underline-offset-4 hover:underline">
        My Games
      </Link>
    </main>
  );
}
