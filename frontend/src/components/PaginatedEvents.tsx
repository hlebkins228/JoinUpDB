import { useEffect, useMemo, useState } from 'react';
import { isPastDate } from '../lib/format';
import type { EventResponse } from '../types';
import { EventCard } from './EventCard';
import { EventSectionDivider } from './EventSectionDivider';

interface PaginatedEventsProps {
  events: EventResponse[];
  onOpen(id: number): void;
  onOpenUser(id: number): void;
  showParticipantUsers?: boolean;
}

const PAGE_SIZE = 20;

export function PaginatedEvents({
  events,
  onOpen,
  onOpenUser,
  showParticipantUsers,
}: PaginatedEventsProps) {
  const [page, setPage] = useState(1);
  const orderedEvents = useMemo(() => [
    ...events.filter((event) => !isPastDate(event.event_time)),
    ...events.filter((event) => isPastDate(event.event_time)),
  ], [events]);
  const pageCount = Math.ceil(events.length / PAGE_SIZE);
  const currentPage = Math.min(page, pageCount);
  const pageEvents = orderedEvents.slice((currentPage - 1) * PAGE_SIZE, currentPage * PAGE_SIZE);

  useEffect(() => setPage(1), [events]);

  return (
    <>
      {pageEvents.map((event, index) => (
        <div key={event.id}>
          {isPastDate(event.event_time) && (index === 0 || !isPastDate(pageEvents[index - 1].event_time)) && (
            <EventSectionDivider label="прошедшие события" />
          )}
          <EventCard
            event={event}
            onOpen={onOpen}
            onOpenUser={onOpenUser}
            showParticipantUsers={showParticipantUsers}
          />
        </div>
      ))}
      {pageCount > 1 && (
        <nav className="pagination" aria-label="страницы объявлений">
          <button type="button" className="btn btn-sm btn-ghost" disabled={currentPage === 1} onClick={() => setPage(currentPage - 1)}>
            назад
          </button>
          <span>{currentPage} / {pageCount}</span>
          <button type="button" className="btn btn-sm btn-ghost" disabled={currentPage === pageCount} onClick={() => setPage(currentPage + 1)}>
            далее
          </button>
        </nav>
      )}
    </>
  );
}
