import { useEffect, useMemo, useState } from 'react';
import { EventCard } from '../components/EventCard';
import { EventSectionDivider } from '../components/EventSectionDivider';
import { getCreatedEvents } from '../api/events';
import type { EventResponse } from '../types';
import { ApiError } from '../api/client';
import { useAuth } from '../auth/AuthContext';
import { isPastDate } from '../lib/format';

interface MyEventsScreenProps {
  onOpenEvent(id: number): void;
  onOpenUser(id: number): void;
  onCreateEvent(): void;
  refreshKey?: number;
}

export function MyEventsScreen({
  onOpenEvent,
  onOpenUser,
  onCreateEvent,
  refreshKey,
}: MyEventsScreenProps) {
  const { user } = useAuth();
  const [events, setEvents] = useState<EventResponse[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    async function load() {
      if (!user) return;
      setLoading(true);
      setError(null);
      try {
        const all = await getCreatedEvents();
        if (cancelled) return;
        setEvents(all.filter((ev) => !ev.deleted));
      } catch (err) {
        if (cancelled) return;
        setError(err instanceof ApiError ? err.message : String(err));
      } finally {
        if (!cancelled) setLoading(false);
      }
    }
    load();
    return () => {
      cancelled = true;
    };
  }, [user, refreshKey]);

  const activeEvents = useMemo(
    () => events.filter((event) => !isPastDate(event.event_time)),
    [events],
  );
  const pastEvents = useMemo(
    () => events.filter((event) => isPastDate(event.event_time)),
    [events],
  );

  return (
    <div className="container">
      <div className="section-head">
        <h2>мои события</h2>
        <button type="button" className="btn btn-sm btn-accent" onClick={onCreateEvent}>
          + создать
        </button>
      </div>

      {error && <div className="alert">{error}</div>}
      {loading ? (
        <div className="empty-state">загружаем…</div>
      ) : events.length === 0 ? (
        <div className="empty-state">вы ещё не создавали событий</div>
      ) : (
        <>
          {activeEvents.map((ev) => (
            <EventCard
              key={ev.id}
              event={ev}
              onOpen={onOpenEvent}
              onOpenUser={onOpenUser}
            />
          ))}
          {pastEvents.length > 0 && (
            <>
              <EventSectionDivider label="прошедшие события" />
              {pastEvents.map((ev) => (
                <EventCard
                  key={ev.id}
                  event={ev}
                  onOpen={onOpenEvent}
                  onOpenUser={onOpenUser}
                />
              ))}
            </>
          )}
        </>
      )}
    </div>
  );
}
