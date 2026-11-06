import type { CSSProperties } from 'react';

type Size = 'sm' | 'md' | 'lg';

interface StarRatingProps {
  value: number | null;
  onChange?: (n: number) => void;
  readOnly?: boolean;
  size?: Size;
}

const SIZE_PX: Record<Size, number> = {
  sm: 14,
  md: 18,
  lg: 24,
};

export function StarRating({
  value,
  onChange,
  readOnly = false,
  size = 'md',
}: StarRatingProps) {
  const interactive = !readOnly && typeof onChange === 'function';
  const fontSize = SIZE_PX[size];
  const current = value ?? 0;

  const wrapStyle: CSSProperties = {
    display: 'inline-flex',
    alignItems: 'center',
    gap: 2,
    lineHeight: 1,
    fontSize,
  };

  return (
    <span
      className="star-rating"
      style={wrapStyle}
      role={interactive ? 'radiogroup' : 'img'}
      aria-label={value === null ? 'нет оценки' : `оценка ${current} из 5`}
    >
      {[1, 2, 3, 4, 5].map((i) => {
        const filled = i <= current;
        const starStyle: CSSProperties = {
          cursor: interactive ? 'pointer' : 'default',
          color: filled ? 'var(--accent)' : 'var(--fg-30)',
          background: 'transparent',
          border: 'none',
          padding: 0,
          fontSize,
          lineHeight: 1,
          userSelect: 'none',
        };
        if (interactive) {
          return (
            <button
              key={i}
              type="button"
              onClick={() => onChange?.(i)}
              style={starStyle}
              role="radio"
              aria-checked={i === current}
              aria-label={`${i} из 5`}
            >
              {filled ? '★' : '☆'}
            </button>
          );
        }
        return (
          <span key={i} style={starStyle} aria-hidden="true">
            {filled ? '★' : '☆'}
          </span>
        );
      })}
    </span>
  );
}
