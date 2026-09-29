import { useEffect, useMemo, useState } from 'react';
import { Field } from '../components/Field';
import { Select, TextInput } from '../components/inputs';
import { CategoryDropdown } from '../components/CategoryDropdown';
import { PaginatedEvents } from '../components/PaginatedEvents';
import { CITIES } from '../lib/constants';
import { getCategories, searchEvents } from '../api/events';
import { fromInputDate } from '../lib/format';
import { ApiError } from '../api/client';
import type { Category, EventResponse } from '../types';

interface SearchScreenProps {
  onOpenEvent(id: number): void;
  onOpenUser(id: number): void;
  onCreateEvent(): void;
  refreshKey?: number;
}

export function SearchScreen({
  onOpenEvent,
  onOpenUser,
  onCreateEvent,
  refreshKey,
}: SearchScreenProps) {
  const [query, setQuery] = useState('');
  const [city, setCity] = useState('');
  const [from, setFrom] = useState('');
  const [to, setTo] = useState('');
  const [activeCats, setActiveCats] = useState<number[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [events, setEvents] = useState<EventResponse[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const cityOptions = useMemo(
    () => [{ value: '', label: 'все города' }, ...CITIES.map((c) => ({ value: c, label: c }))],
    [],
  );

  async function runSearch() {
    setLoading(true);
    setError(null);
    try {
      const data = await searchEvents({
        name: query.trim() || undefined,
        city: city || undefined,
        event_from: fromInputDate(from),
        event_to: fromInputDate(to),
        category_id: activeCats.length ? activeCats : undefined,
      });
      setEvents(data);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : String(err));
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    runSearch();
    getCategories()
      .then(setCategories)
      .catch((err) => setError(err instanceof ApiError ? err.message : String(err)));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [refreshKey]);

  function toggleCat(id: number) {
    setActiveCats((prev) =>
      prev.includes(id) ? prev.filter((x) => x !== id) : [...prev, id],
    );
  }

  function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    runSearch();
  }

  return (
    <div className="container">
      <form className="search-bar" onSubmit={onSubmit}>
        <div className="search-input-wrap">
          <span className="search-icon">⌕</span>
          <TextInput
            value={query}
            onChange={setQuery}
            placeholder="искать по названию"
          />
        </div>
        <button type="submit" className="btn">
          найти
        </button>
      </form>

      <div className="filters">
        <Field label="город">
          <Select
            value={city}
            onChange={setCity}
            options={cityOptions}
          />
        </Field>
        <Field label="когда">
          <div className="date-row">
            <input
              className="input"
              type="date"
              value={from}
              onChange={(e) => setFrom(e.target.value)}
              aria-label="с"
            />
            <input
              className="input"
              type="date"
              value={to}
              onChange={(e) => setTo(e.target.value)}
              aria-label="по"
            />
          </div>
        </Field>
        <Field label="категории">
          <CategoryDropdown categories={categories} activeCats={activeCats} onToggle={toggleCat} />
        </Field>
      </div>

      <div className="section-head">
        <h2>события</h2>
        <button type="button" className="btn btn-sm btn-accent" onClick={onCreateEvent}>
          + создать
        </button>
      </div>

      {error && <div className="alert">{error}</div>}
      {loading ? (
        <div className="empty-state">загружаем…</div>
      ) : events.length === 0 ? (
        <div className="empty-state">ничего не нашли</div>
      ) : (
        <PaginatedEvents
          events={events}
          onOpen={onOpenEvent}
          onOpenUser={onOpenUser}
          showParticipantUsers={false}
        />
      )}
    </div>
  );
}
