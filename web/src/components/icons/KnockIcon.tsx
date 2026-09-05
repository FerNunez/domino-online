import { SVGProps } from "react";

// International "knock to pass" gesture — a fist rapping the table.
// Drawn as outline linework (matching the app's lucide-react icon set)
// rather than a solid silhouette: at icon-button size, separate strokes
// for the fingers/thumb stay legible where filled shapes would just
// fuse into a blob.
export function KnockIcon(props: SVGProps<SVGSVGElement>) {
  return (
    <svg
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth={1.8}
      strokeLinecap="round"
      strokeLinejoin="round"
      {...props}
    >
      <path d="M8 10.5c0-1.4 1.6-2.3 3-2.3h1.6c1.6 0 2.9 1.1 3.3 2.6l.4 1.6c.5 1.9-.9 3.8-2.9 3.8H10.5A2.5 2.5 0 0 1 8 13.7z" />
      <path d="M9.8 8.4V6.6" />
      <path d="M12.4 8V5.9" />
      <path d="M15 8.6v-1.9" />
      <path d="M8 12c-1.8-.2-3 .7-3 2.2s1.3 2.4 3.1 2" />
      <path d="M3.5 19.8h17" />
      <path d="M18 4.5 16.4 6.8" />
      <path d="M21 6.5 19 8.3" />
    </svg>
  );
}
