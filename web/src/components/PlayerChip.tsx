import { cn } from "@/lib/utils";

interface PlayerChipProps {
  name: string;
  size?: "sm" | "md";
  active?: boolean;
  className?: string;
}

const SIZE_CLASSES: Record<NonNullable<PlayerChipProps["size"]>, string> = {
  sm: "h-5 w-5 text-[10px]",
  md: "h-10 w-10 text-sm",
};

// The two-letter badge used to identify a player without printing their raw
// guest id — every guest's display name *is* its id (no friendly names are
// assigned), so this is the only compact, consistent way to tell seats apart
// at a glance. Shared by PlayerSeat (at the table) and RoundActionsList (in
// the move history) so a player reads the same everywhere.
export function PlayerChip({ name, size = "md", active, className }: PlayerChipProps) {
  return (
    <div
      className={cn(
        "flex shrink-0 items-center justify-center rounded-full font-semibold",
        active ? "bg-primary text-primary-foreground" : "bg-muted text-muted-foreground",
        SIZE_CLASSES[size],
        className
      )}
    >
      {name.slice(0, 2).toUpperCase()}
    </div>
  );
}
