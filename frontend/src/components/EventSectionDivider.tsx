interface EventSectionDividerProps {
  label: string;
}

export function EventSectionDivider({ label }: EventSectionDividerProps) {
  return (
    <div className="event-section-divider">
      <span>{label}</span>
    </div>
  );
}
