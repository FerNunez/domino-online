import { SVGProps } from "react";

// International "knock to pass" gesture — a closed fist rapping the table.
// Solid silhouette (rather than fine linework) so it stays a clean, legible
// fist shape at the small size an icon button uses.
export function KnockIcon(props: SVGProps<SVGSVGElement>) {
  return (
    <svg viewBox="0 0 24 24" fill="none" {...props}>
      <g fill="currentColor">
        <rect x="9" y="19" width="8" height="3.5" rx="1.6" />
        <rect x="6" y="11" width="13" height="9" rx="4" />
        <circle cx="8.4" cy="10.6" r="2.6" />
        <circle cx="11.9" cy="9.8" r="2.6" />
        <circle cx="15.4" cy="9.8" r="2.6" />
        <circle cx="18.6" cy="10.6" r="2.6" />
        <ellipse cx="4.6" cy="15.4" rx="2.7" ry="3.6" transform="rotate(-12 4.6 15.4)" />
      </g>
      <g stroke="currentColor" strokeWidth="1.6" strokeLinecap="round">
        <path d="M17 3.5 15 6" />
        <path d="M20.5 5 18.2 7" />
        <path d="M21.5 8.5 19 9.8" />
      </g>
    </svg>
  );
}
