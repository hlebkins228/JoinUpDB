import { useEffect, useState } from 'react';
import { PaginatedEvents } from '../components/PaginatedEvents';
import { getJoinedEvents } from '../api/events';
import type { EventResponse } from '../types';
import { ApiError } from '../api/client';

interface JoiningScreenProps {
  onOpenEvent(id: number): void;
  onOpenUser(id: number): void;
  refreshKey?: number;
}

export function JoiningScreen({ onOpenEvent, onOpenUser, refreshKey }: JoiningScreenProps) {
  const [events, setEvents] = useState<EventResponse[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    async function load() {
      setLoading(true);
      setError(null);
      try {
        const joined = await getJoinedEvents();
        if (cancelled) return;
        setEvents(joined);
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
  }, [refreshKey]);

  return (
    <div className="container">
      <div className="section-head">
        <h2>я участвую</h2>
      </div>

      {error && <div className="alert">{error}</div>}
      {loading ? (
        <div className="empty-state">загружаем…</div>
      ) : events.length === 0 ? (
        <div className="empty-state">вы ещё ни к чему не присоединились</div>
      ) : (
        <PaginatedEvents events={events} onOpen={onOpenEvent} onOpenUser={onOpenUser} />
      )}
    </div>
  );
}
