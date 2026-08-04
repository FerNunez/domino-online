// Client-side, unverified decode — only for reading the userID claim to
// drive UI state (whose turn, host badge). Mirrors scripts/lobby-viewer.html's
// jwtUserId helper. The server is the source of truth for authorization.
export function decodeJwtUserId(token: string): string | null {
  try {
    const payload = token.split(".")[1];
    const json = JSON.parse(atob(payload.replace(/-/g, "+").replace(/_/g, "/")));
    return typeof json.userID === "string" ? json.userID : null;
  } catch {
    return null;
  }
}
