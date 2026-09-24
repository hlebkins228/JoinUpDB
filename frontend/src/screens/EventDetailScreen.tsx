import { useEffect, useMemo, useState, type FormEvent } from 'react';
import { Field } from '../components/Field';
import { Select, TextArea, TextInput } from '../components/inputs';
import { CategoryTag } from '../components/CategoryTag';
import { UploadEventCover } from '../components/UploadEventCover';
import { MiniMap } from '../components/MiniMap';
import { AddressPicker } from '../components/AddressPicker';
import { UserChip } from '../components/UserChip';
import { ReviewForm } from '../components/ReviewForm';
import { ReviewList } from '../components/ReviewList';
import { CITIES } from '../lib/constants';
import {
  addEventCategory,
  createEvent,
  deleteEvent,
  getCategories,
  getEvent,
  joinEvent,
  leaveEvent,
  updateEvent,
  uploadEventImage,
} from '../api/events';
import { listEventReviews, type Review } from '../api/reviews';
import { ApiError } from '../api/client';
import {
  fmtDate,
  fmtDateTime,
  currentInputDateTime,
  fromInputDateTime,
  isPastDate,
  toInputDateTime,
} from '../lib/format';
import { imageUrl } from '../lib/imageUrl';
import { useAuth } from '../auth/AuthContext';
import type { Category, EventResponse } from '../types';

interface EventDetailScreenProps {
  eventId: number | null;
  onOpenUser(id: number): void;
  onBack(): void;
  onAfterDelete(): void;
  refreshKey?: number;
}

interface FormState {
  name: string;
  desc: string;
  city: string;
  eventTime: string;
  tgChat: string;
  locName: string;
  address: string;
  latitude: number | null;
  longitude: number | null;
}

const EMPTY_FORM: FormState = {
  name: '',
  desc: '',
  city: '',
  eventTime: '',
  tgChat: '',
  locName: '',
  address: '',
  latitude: null,
  longitude: null,
};

function eventToForm(ev: EventResponse): FormState {
  return {
    name: ev.name,
    desc: ev.desc ?? '',
    city: ev.city,
    eventTime: toInputDateTime(ev.event_time),
    tgChat: ev.telegram_chat_url ?? '',
    locName: ev.location?.name ?? '',
    address: ev.location?.address ?? '',
    latitude: ev.location?.latitude ?? null,
    longitude: ev.location?.longitude ?? null,
  };
}

export function EventDetailScreen({
  eventId,
  onOpenUser,
  onBack,
  onAfterDelete,
  refreshKey,
}: EventDetailScreenProps) {
  const { user } = useAuth();
  const [event, setEvent] = useState<EventResponse | null>(null);
  const [createdId, setCreatedId] = useState<number | null>(null);
  const isNew = eventId === null && createdId === null;
  const [form, setForm] = useState<FormState>(EMPTY_FORM);
  const [editing, setEditing] = useState<boolean>(eventId === null);
  const [activeCats, setActiveCats] = useState<number[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);
  const [coverPreview, setCoverPreview] = useState<string | null>(null);
  const [coverImageId, setCoverImageId] = useState<number | null>(null);
  const [pendingCover, setPendingCover] = useState<File | null>(null);
  // Tracks the value the picker last auto-suggested for `name`. We only
  // overwrite the name field while it still matches this value (i.e. the
  // user hasn't customised it). Once the user edits the name we leave it
  // alone on subsequent picks — they can clear it to opt back in.
  const [autoFilledName, setAutoFilledName] = useState<string>('');

  const [reviews, setReviews] = useState<Review[]>([]);
  const [reviewsLoading, setReviewsLoading] = useState<boolean>(false);
  const [reviewsError, setReviewsError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    getCategories()
      .then((fetched) => {
        if (!cancelled) setCategories(fetched);
      })
      .catch((err) => {
        if (!cancelled) setError(err instanceof ApiError ? err.message : String(err));
      });
    return () => {
      cancelled = true;
    };
  }, [refreshKey]);

  useEffect(() => {
    if (eventId == null) {
      setEvent(null);
      setCreatedId(null);
      setForm(EMPTY_FORM);
      setEditing(true);
      setActiveCats([]);
      setError(null);
      setCoverPreview(null);
      setCoverImageId(null);
      setPendingCover(null);
      setAutoFilledName('');
      setReviews([]);
      setReviewsLoading(false);
      setReviewsError(null);
      return;
    }
    let cancelled = false;
    async function load() {
      setBusy(true);
      setError(null);
      try {
        const fetched = await getEvent(eventId as number);
        if (cancelled) return;
        setEvent(fetched);
        setForm(eventToForm(fetched));
        setCoverImageId(fetched.image_id ?? null);
        setCoverPreview(null);
        setEditing(false);
      } catch (err) {
        if (cancelled) return;
        setError(err instanceof ApiError ? err.message : String(err));
      } finally {
        if (!cancelled) setBusy(false);
      }
    }
    load();
    return () => {
      cancelled = true;
    };
  }, [eventId, refreshKey]);

  useEffect(() => {
    if (!event) {
      setReviews([]);
      setReviewsLoading(false);
      setReviewsError(null);
      return;
    }
    let cancelled = false;
    setReviewsLoading(true);
    setReviewsError(null);
    listEventReviews(event.id)
      .then((fetched) => {
        if (!cancelled) setReviews(fetched);
      })
      .catch((err) => {
        if (cancelled) return;
        setReviewsError(err instanceof ApiError ? err.message : String(err));
      })
      .finally(() => {
        if (!cancelled) setReviewsLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [event?.id, refreshKey]);

  const isCreator = useMemo(
    () => !!user && !!event && event.creator_id === user.id,
    [user, event],
  );
  const isParticipant = useMemo(
    () => !!user && !!event?.members?.some((member) => member.id === user.id),
    [user, event],
  );
  const isPast = !!event && isPastDate(event.event_time);
  const canEdit = isCreator && !isPast;
  const hasOwnReview = useMemo(
    () => !!user && reviews.some((r) => r.author.id === user.id),
    [user, reviews],
  );
  const creator = event?.creator ?? (event ? { id: event.creator_id, name: `#${event.creator_id}` } : null);

  const cityOptions = useMemo(
    () => CITIES.map((c) => ({ value: c, label: c })),
    [],
  );

  function setField<K extends keyof FormState>(key: K, value: FormState[K]) {
    setForm((prev) => ({ ...prev, [key]: value }));
  }

  function toggleCat(id: number) {
    setActiveCats((prev) =>
      prev.includes(id) ? prev.filter((x) => x !== id) : [...prev, id],
    );
  }

  function validate(): string | null {
    if (form.name.trim().length < 5 || form.name.trim().length > 100) {
      return 'название от 5 до 100 символов';
    }
    if (!form.eventTime) return 'укажите дату и время';
    if (form.eventTime < currentInputDateTime()) {
      return 'нельзя создать или сохранить событие в прошлом';
    }
    if (!form.city) return 'выберите город';
    if (!form.locName.trim()) return 'укажите название места';
    if (!form.address.trim()) return 'укажите адрес';
    if (form.latitude == null || form.longitude == null) {
      return 'выберите точку на карте';
    }
    return null;
  }

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setError(null);
    const validation = validate();
    if (validation) {
      setError(validation);
      return;
    }

    setBusy(true);
    try {
      const payload = {
        name: form.name.trim(),
        desc: form.desc.trim() || undefined,
        event_time: fromInputDateTime(form.eventTime),
        telegram_chat_url: form.tgChat.trim() || undefined,
        city: form.city,
        location: {
          name: form.locName.trim(),
          address: form.address.trim(),
          latitude: form.latitude as number,
          longitude: form.longitude as number,
        },
        image_id: coverImageId ?? undefined,
      };

      let nextId: number;
      if (event) {
        nextId = event.id;
        await updateEvent({ ...payload, id: event.id });
      } else if (createdId != null) {
        nextId = createdId;
        await updateEvent({ ...payload, id: createdId });
      } else {
        const created = await createEvent(payload);
        nextId = created.id;
        setCreatedId(created.id);
      }

      if (pendingCover) {
        try {
          const { id } = await uploadEventImage(nextId, pendingCover);
          setCoverImageId(id);
          setPendingCover(null);
        } catch (err) {
          setError(
            'событие сохранено, но обложку загрузить не удалось: ' +
              (err instanceof Error ? err.message : String(err)),
          );
        }
      }

      for (const catId of activeCats) {
        try {
          await addEventCategory(nextId, catId);
        } catch (err) {
          if (err instanceof ApiError && err.status !== 409) {
            console.warn('add category failed', err);
          }
        }
      }

      const refreshed = await getEvent(nextId);
      setEvent(refreshed);
      setForm(eventToForm(refreshed));
      setCoverImageId(refreshed.image_id ?? null);
      setEditing(false);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : String(err));
    } finally {
      setBusy(false);
    }
  }

  async function onDelete() {
    if (!event) return;
    if (!window.confirm('удалить событие?')) return;
    setBusy(true);
    try {
      await deleteEvent(event.id);
      onAfterDelete();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : String(err));
    } finally {
      setBusy(false);
    }
  }

  async function onJoin() {
    if (!event) return;
    setBusy(true);
    setError(null);
    try {
      await joinEvent(event.id);
      const refreshed = await getEvent(event.id);
      setEvent(refreshed);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : String(err));
    } finally {
      setBusy(false);
    }
  }

  async function onLeave() {
    if (!event) return;
    setBusy(true);
    setError(null);
    try {
      await leaveEvent(event.id);
      const refreshed = await getEvent(event.id);
      setEvent(refreshed);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : String(err));
    } finally {
      setBusy(false);
    }
  }

  if (busy && !event && !isNew) {
    return (
      <div className="container">
        <div className="empty-state">загружаем…</div>
      </div>
    );
  }

  if (!editing && event) {
    const cover = coverPreview ?? imageUrl(event.image_id);
    return (
      <div className="container">
        <div className="section-head">
          <button type="button" className="btn btn-sm btn-ghost" onClick={onBack}>
            ← назад
          </button>
          <div className="row">
            {isPast && <span className="muted">событие прошло</span>}
            {canEdit && (
              <>
                <button
                  type="button"
                  className="btn btn-sm btn-ghost"
                  onClick={() => setEditing(true)}
                >
                  редактировать
                </button>
                <button
                  type="button"
                  className="btn btn-sm btn-danger"
                  onClick={onDelete}
                  disabled={busy}
                >
                  удалить
                </button>
              </>
            )}
            {!isPast && !isCreator && !isParticipant && (
              <button
                type="button"
                className="btn btn-accent btn-sm"
                onClick={onJoin}
                disabled={busy}
              >
                {busy ? '…' : 'присоединиться'}
              </button>
            )}
            {!isPast && !isCreator && isParticipant && (
              <button
                type="button"
                className="btn btn-sm btn-ghost"
                onClick={onLeave}
                disabled={busy}
              >
                {busy ? '…' : 'выйти'}
              </button>
            )}
          </div>
        </div>

        {error && <div className="alert">{error}</div>}

        <UploadEventCover
          src={cover}
          editable={false}
          onPicked={() => undefined}
        />

        <div className="detail-title-row">
          <h1 className="detail-title">{event.name}</h1>
        </div>

        <div className="detail-grid">
          <div>
            <section className="detail-section">
              <h3>описание</h3>
              <p>{event.desc?.trim() || 'без описания'}</p>
            </section>

            {event.location && (
              <section className="detail-section">
                <h3>место</h3>
                <p>
                  <strong>{event.location.name}</strong>
                  <br />
                  {event.location.address}
                </p>
                <div style={{ marginTop: 12 }}>
                  <MiniMap
                    longitude={event.location.longitude}
                    latitude={event.location.latitude}
                  />
                </div>
              </section>
            )}

            <section className="detail-section">
              <h3>участники: {event.members?.length ?? 0}</h3>
              {event.members && event.members.length > 0 ? (
                <div className="participants">
                  {event.members.map((member) => (
                    <UserChip
                      key={member.id}
                      user={member}
                      onOpenUser={onOpenUser}
                    />
                  ))}
                </div>
              ) : (
                <p className="muted">пока никто не присоединился</p>
              )}
            </section>

            <section className="detail-section">
              <h3>отзывы</h3>
              {!isCreator && isPast && !hasOwnReview && user && (
                <ReviewForm
                  eventId={event.id}
                  onCreated={(review) =>
                    setReviews((prev) => [review, ...prev])
                  }
                />
              )}
              <ReviewList
                reviews={reviews}
                loading={reviewsLoading}
                error={reviewsError}
                onOpenUser={onOpenUser}
              />
            </section>
          </div>

          <aside className="detail-meta">
            <div className="detail-meta-row">
              <span className="label">когда</span>
              <span>{fmtDateTime(event.event_time)}</span>
            </div>
            <div className="detail-meta-row">
              <span className="label">город</span>
              <span>{event.city}</span>
            </div>
            {creator && (
              <div className="detail-meta-row">
                <span className="label">создатель</span>
                <UserChip user={creator} onOpenUser={onOpenUser} />
              </div>
            )}
            {event.telegram_chat_url && (
              <div className="detail-meta-row">
                <span className="label">чат</span>
                <a
                  href={event.telegram_chat_url}
                  target="_blank"
                  rel="noopener noreferrer"
                >
                  {event.telegram_chat_url}
                </a>
              </div>
            )}
            <div className="detail-meta-row">
              <span className="label">создано</span>
              <span>{fmtDate(event.created_at)}</span>
            </div>
            <div className="detail-meta-row">
              <span className="label">обновлено</span>
              <span>{fmtDate(event.updated_at)}</span>
            </div>
          </aside>
        </div>
      </div>
    );
  }

  return (
    <div className="container">
      <div className="section-head">
        <button
          type="button"
          className="btn btn-sm btn-ghost"
          onClick={() => {
            if (event) {
              setEditing(false);
              setForm(eventToForm(event));
              setActiveCats([]);
            } else {
              onBack();
            }
          }}
        >
          ← назад
        </button>
        <h2>{isNew ? 'новое событие' : 'редактировать'}</h2>
      </div>

      <form onSubmit={onSubmit} noValidate>
        {error && <div className="alert">{error}</div>}

        <div style={{ marginBottom: 16 }}>
          <UploadEventCover
            src={coverPreview ?? imageUrl(coverImageId)}
            editable
            eventId={event?.id ?? null}
            onPicked={(file, preview) => {
              if (event) {
                setCoverPreview(preview);
              } else {
                setPendingCover(file);
                setCoverPreview(preview);
              }
            }}
            onUploaded={(id, preview) => {
              setCoverImageId(id);
              setCoverPreview(preview);
            }}
            onError={(message) => setError(message)}
          />
        </div>

        <Field label="название">
          <TextInput
            value={form.name}
            onChange={(v) => setField('name', v)}
            placeholder="Утренняя пробежка"
          />
        </Field>
        <Field label="описание" optional>
          <TextArea
            value={form.desc}
            onChange={(v) => setField('desc', v)}
            placeholder="о чём событие, что взять с собой…"
          />
        </Field>
        <Field label="когда">
          <input
            className="input"
            type="datetime-local"
            value={form.eventTime}
            onChange={(e) => setField('eventTime', e.target.value)}
            min={currentInputDateTime()}
            max="9999-12-31T23:59"
          />
        </Field>
        <Field label="город">
          <Select
            value={form.city}
            onChange={(v) => setField('city', v)}
            options={cityOptions}
            placeholder="выберите город"
          />
        </Field>
        <Field label="место (название)" hint="по умолчанию подставляется адрес или название из подсказки — поле можно отредактировать">
          <TextInput
            value={form.locName}
            onChange={(v) => setField('locName', v)}
            placeholder="Парк Горького"
          />
        </Field>
        <div className="field">
          <span className="field-label">
            <span>адрес</span>
          </span>
          <AddressPicker
            address={form.address}
            latitude={form.latitude}
            longitude={form.longitude}
            city={form.city}
            onAddressChange={(value) => setField('address', value)}
            onPick={(picked) => {
              setForm((prev) => {
                const shouldAutoFillName =
                  prev.locName.trim() === '' ||
                  prev.locName === autoFilledName;
                return {
                  ...prev,
                  address: picked.address,
                  latitude: picked.latitude,
                  longitude: picked.longitude,
                  locName: shouldAutoFillName ? picked.name : prev.locName,
                };
              });
              setAutoFilledName(picked.name);
            }}
          />
        </div>
        <Field label="ссылка на чат" optional>
          <TextInput
            value={form.tgChat}
            onChange={(v) => setField('tgChat', v)}
            placeholder="https://t.me/…"
          />
        </Field>

        {isNew && (
          <Field label="категории" optional>
            <div className="filters-cats">
              {categories.map((c) => (
                <CategoryTag
                  key={c.id}
                  category={c}
                  toggle
                  active={activeCats.includes(c.id)}
                  onClick={() => toggleCat(c.id)}
                />
              ))}
            </div>
          </Field>
        )}

        <div className="row-end" style={{ marginTop: 24 }}>
          <button type="submit" className="btn" disabled={busy}>
            {busy ? 'сохраняем…' : isNew ? 'создать' : 'сохранить'}
          </button>
        </div>
      </form>
    </div>
  );
}
