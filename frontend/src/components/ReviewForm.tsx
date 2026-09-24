import { useState, type CSSProperties, type FormEvent } from 'react';
import { StarRating } from './StarRating';
import { createEventReview, type Review } from '../api/reviews';
import { ApiError } from '../api/client';

interface ReviewFormProps {
  eventId: number;
  onCreated(review: Review): void;
}

const MAX_TEXT_LENGTH = 2000;

const wrapStyle: CSSProperties = {
  display: 'flex',
  flexDirection: 'column',
  gap: 12,
  marginBottom: 16,
};

const ratingRowStyle: CSSProperties = {
  display: 'flex',
  alignItems: 'center',
  gap: 12,
};

const fieldErrorStyle: CSSProperties = {
  color: 'var(--danger, #d33)',
  fontSize: 12,
};

const counterStyle: CSSProperties = {
  alignSelf: 'flex-end',
  color: 'var(--fg-50)',
  fontSize: 12,
};

const actionsStyle: CSSProperties = {
  display: 'flex',
  justifyContent: 'flex-end',
};

function isAbortError(err: unknown): boolean {
  return (
    typeof err === 'object' &&
    err !== null &&
    'name' in err &&
    (err as { name?: string }).name === 'AbortError'
  );
}

export function ReviewForm({ eventId, onCreated }: ReviewFormProps) {
  const [rating, setRating] = useState<number | null>(null);
  const [text, setText] = useState<string>('');
  const [submitting, setSubmitting] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const [ratingError, setRatingError] = useState<string | null>(null);
  const [textError, setTextError] = useState<string | null>(null);

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    if (submitting) return;

    // Per-field validation. Do not call the API on invalid input.
    const trimmed = text.trim();
    let invalid = false;

    if (rating === null || !Number.isInteger(rating) || rating < 1 || rating > 5) {
      setRatingError('выберите оценку от 1 до 5');
      invalid = true;
    } else {
      setRatingError(null);
    }

    if (trimmed.length < 1) {
      setTextError('напишите комментарий');
      invalid = true;
    } else if (trimmed.length > MAX_TEXT_LENGTH) {
      setTextError(`не более ${MAX_TEXT_LENGTH} символов`);
      invalid = true;
    } else {
      setTextError(null);
    }

    if (invalid) return;

    setError(null);
    setSubmitting(true);
    try {
      const created = await createEventReview(eventId, {
        rating: rating as number,
        text: trimmed,
      });
      // Reset form. Parent unmounts the form within ≤ 1s.
      setRating(null);
      setText('');
      onCreated(created);
    } catch (err) {
      if (isAbortError(err)) {
        setError('превышено время ожидания, попробуйте ещё раз');
      } else if (err instanceof ApiError) {
        setError(err.message);
      } else {
        setError(err instanceof Error ? err.message : String(err));
      }
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <form className="review-form" onSubmit={onSubmit} noValidate style={wrapStyle}>
      {error && <div className="alert">{error}</div>}

      <div>
        <div style={ratingRowStyle}>
          <span className="muted">оценка</span>
          <StarRating
            value={rating}
            onChange={(n) => {
              setRating(n);
              if (ratingError) setRatingError(null);
            }}
            size="lg"
          />
        </div>
        {ratingError && <span style={fieldErrorStyle}>{ratingError}</span>}
      </div>

      <div style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
        <textarea
          className="input"
          value={text}
          onChange={(e) => {
            setText(e.target.value);
            if (textError) setTextError(null);
          }}
          maxLength={MAX_TEXT_LENGTH}
          rows={4}
          placeholder="расскажите о впечатлениях"
        />
        <span style={counterStyle}>
          {text.length}/{MAX_TEXT_LENGTH}
        </span>
        {textError && <span style={fieldErrorStyle}>{textError}</span>}
      </div>

      <div style={actionsStyle}>
        <button type="submit" className="btn" disabled={submitting}>
          {submitting ? 'отправляем…' : 'отправить отзыв'}
        </button>
      </div>
    </form>
  );
}
