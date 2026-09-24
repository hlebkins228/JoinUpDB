import type { CSSProperties } from 'react';
import { Avatar } from './Avatar';
import { StarRating } from './StarRating';
import { imageUrl } from '../lib/imageUrl';
import { fmtDate } from '../lib/format';
import type { Review } from '../api/reviews';

interface ReviewListProps {
  reviews: Review[];
  loading: boolean;
  error: string | null;
  onOpenUser(id: number): void;
}

const listStyle: CSSProperties = {
  display: 'flex',
  flexDirection: 'column',
  gap: 16,
};

const itemStyle: CSSProperties = {
  display: 'flex',
  flexDirection: 'column',
  gap: 8,
  padding: 12,
  borderRadius: 8,
  background: 'var(--bg-2, transparent)',
  border: '1px solid var(--fg-15, rgba(0,0,0,0.08))',
};

const headerStyle: CSSProperties = {
  display: 'flex',
  alignItems: 'center',
  gap: 8,
  flexWrap: 'wrap',
};

const dateStyle: CSSProperties = {
  marginLeft: 'auto',
  color: 'var(--fg-50)',
  fontSize: 12,
};

const textStyle: CSSProperties = {
  margin: 0,
  whiteSpace: 'pre-wrap',
  wordBreak: 'break-word',
};

const authorButtonStyle: CSSProperties = {
  display: 'inline-flex',
  alignItems: 'center',
  gap: 6,
  background: 'transparent',
  border: 'none',
  padding: 0,
  cursor: 'pointer',
  font: 'inherit',
  color: 'inherit',
};

export function ReviewList({ reviews, loading, error, onOpenUser }: ReviewListProps) {
  if (loading) {
    return <div className="empty-state">загружаем…</div>;
  }

  return (
    <div className="review-list" style={listStyle}>
      {error && <div className="alert">{error}</div>}

      {!error && reviews.length === 0 && (
        <p className="muted">пока нет отзывов</p>
      )}

      {reviews.map((review) => (
        <article key={review.id} className="review-item" style={itemStyle}>
          <header style={headerStyle}>
            <button
              type="button"
              className="review-author"
              style={authorButtonStyle}
              onClick={() => onOpenUser(review.author.id)}
            >
              <Avatar
                src={imageUrl(review.author.avatar_id)}
                name={review.author.name}
                size="sm"
              />
              <span>{review.author.name}</span>
            </button>
            <StarRating value={review.rating} readOnly size="sm" />
            <span style={dateStyle}>{fmtDate(review.created_at)}</span>
          </header>
          <p style={textStyle}>{review.text}</p>
        </article>
      ))}
    </div>
  );
}
