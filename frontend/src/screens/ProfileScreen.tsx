import { useEffect, useMemo, useState, type FormEvent } from 'react';
import { Field } from '../components/Field';
import { NumberInput, Select, TextInput } from '../components/inputs';
import { UploadAvatar } from '../components/UploadAvatar';
import { Avatar } from '../components/Avatar';
import { CITIES } from '../lib/constants';
import { getUser, updateUser } from '../api/users';
import { ApiError } from '../api/client';
import { useAuth } from '../auth/AuthContext';
import { imageUrl } from '../lib/imageUrl';
import type { UserResponse } from '../types';

interface ProfileScreenProps {
  userId?: number | null;
  refreshKey?: number;
  onBack(): void;
}

function ProfileStats({ user }: { user: UserResponse }) {
  const ratingDisplay =
    user.organizer_rating === null ? '—' : user.organizer_rating.toFixed(1);
  const reviewsCount =
    user.organizer_rating === null ? 0 : user.organizer_reviews_count;
  return (
    <div className="profile-stats">
      <div>
        <span className="profile-stat-value">{user.created_events_count}</span>
        <span className="muted">создано событий</span>
      </div>
      <div>
        <span className="profile-stat-value">{user.joined_events_count}</span>
        <span className="muted">участий</span>
      </div>
      <div>
        <span className="profile-stat-value">
          {ratingDisplay}
          <span aria-hidden="true" style={{ color: 'var(--accent)', marginLeft: 6 }}>
            ★
          </span>
        </span>
        <span className="muted">рейтинг</span>
      </div>
      <div>
        <span className="profile-stat-value">{reviewsCount}</span>
        <span className="muted">кол-во отзывов</span>
      </div>
    </div>
  );
}

export function ProfileScreen({ userId, refreshKey, onBack }: ProfileScreenProps) {
  const { user: currentUser, refreshUser, signOut } = useAuth();
  const isOwnProfile = !userId || userId === currentUser?.id;
  const [profileUser, setProfileUser] = useState<UserResponse | null>(currentUser);
  const [name, setName] = useState('');
  const [age, setAge] = useState<number | ''>('');
  const [city, setCity] = useState('');
  const [tg, setTg] = useState('');
  const [avatarId, setAvatarId] = useState<number | null>(null);
  const [avatarPreview, setAvatarPreview] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);

  useEffect(() => {
    if (isOwnProfile) {
      setProfileUser(currentUser);
      return;
    }

    let cancelled = false;
    async function load() {
      if (!userId) return;
      setLoading(true);
      setError(null);
      try {
        const fetched = await getUser(userId);
        if (!cancelled) setProfileUser(fetched);
      } catch (err) {
        if (!cancelled) setError(err instanceof ApiError ? err.message : String(err));
      } finally {
        if (!cancelled) setLoading(false);
      }
    }
    load();
    return () => {
      cancelled = true;
    };
  }, [currentUser, isOwnProfile, userId, refreshKey]);

  useEffect(() => {
    if (!isOwnProfile) return;
    refreshUser().catch((err) => {
      setError(err instanceof ApiError ? err.message : String(err));
    });
  }, [isOwnProfile, refreshKey, refreshUser]);

  useEffect(() => {
    if (!profileUser || !isOwnProfile) return;
    setName(profileUser.name);
    setAge(profileUser.age);
    setCity(profileUser.city);
    setTg(profileUser.telegram_login ?? '');
    setAvatarId(profileUser.avatar_id ?? null);
    setAvatarPreview(null);
  }, [profileUser, isOwnProfile]);

  const cityOptions = useMemo(
    () => CITIES.map((c) => ({ value: c, label: c })),
    [],
  );

  if (loading) {
    return (
      <div className="container">
        <div className="empty-state">загружаем…</div>
      </div>
    );
  }

  if (!profileUser) return null;

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setError(null);
    setSuccess(false);
    if (!name.trim() || age === '' || !city) {
      setError('заполните обязательные поля');
      return;
    }
    setBusy(true);
    try {
      await updateUser({
        name: name.trim(),
        age: Number(age),
        city,
        telegram_login: tg.trim() || undefined,
        avatar_id: avatarId ?? undefined,
      });
      await refreshUser();
      setSuccess(true);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : String(err));
    } finally {
      setBusy(false);
    }
  }

  if (!isOwnProfile) {
    return (
      <div className="container">
        <div className="section-head">
          <h2>профиль</h2>
          <button type="button" className="btn btn-sm btn-ghost" onClick={onBack}>
            назад
          </button>
        </div>

        {error && <div className="alert">{error}</div>}
        <div className="profile-grid">
          <div>
            <Avatar
              src={imageUrl(profileUser.avatar_id)}
              name={profileUser.name}
              size="lg"
            />
          </div>
          <div className="detail-meta">
            <div className="detail-meta-row">
              <span className="label">имя</span>
              <span>{profileUser.name}</span>
            </div>
            <div className="detail-meta-row">
              <span className="label">город</span>
              <span>{profileUser.city}</span>
            </div>
            {profileUser.telegram_login && (
              <div className="detail-meta-row">
                <span className="label">telegram</span>
                <span>{profileUser.telegram_login}</span>
              </div>
            )}
          </div>
        </div>
        <div className="profile-foreign-stats">
          <ProfileStats user={profileUser} />
        </div>
      </div>
    );
  }

  return (
    <div className="container">
      <div className="section-head">
        <h2>профиль</h2>
        <button type="button" className="btn btn-sm btn-ghost" onClick={onBack}>
          назад
        </button>
      </div>

      <div className="profile-grid">
        <div>
          <UploadAvatar
            src={avatarPreview ?? imageUrl(avatarId)}
            name={profileUser.name}
            onUploaded={(id, preview) => {
              setAvatarId(id);
              setAvatarPreview(preview);
            }}
            onError={(message) => setError(message)}
          />
        </div>

        <form onSubmit={onSubmit} noValidate>
          {error && <div className="alert">{error}</div>}
          {success && (
            <div
              className="alert"
              style={{
                borderColor: 'var(--fg-30)',
                color: 'var(--fg)',
                background: 'transparent',
              }}
            >
              сохранено
            </div>
          )}

          <ProfileStats user={profileUser} />

          <Field label="имя">
            <TextInput value={name} onChange={setName} />
          </Field>
          <Field label="логин" hint="логин нельзя поменять">
            <TextInput
              value={profileUser.login}
              onChange={() => undefined}
              disabled
            />
          </Field>
          <Field label="возраст">
            <NumberInput value={age} onChange={setAge} min={1} max={120} />
          </Field>
          <Field label="город">
            <Select value={city} onChange={setCity} options={cityOptions} />
          </Field>
          <Field label="telegram" optional>
            <TextInput value={tg} onChange={setTg} placeholder="anna_k" />
          </Field>

          <div className="row-end" style={{ marginTop: 16 }}>
            <button
              type="button"
              className="btn btn-danger btn-sm"
              onClick={signOut}
            >
              выйти
            </button>
            <button type="submit" className="btn" disabled={busy}>
              {busy ? 'сохраняем…' : 'сохранить'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
