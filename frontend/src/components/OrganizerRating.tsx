interface OrganizerRatingProps {
  rating: number | null;
  reviewsCount: number;
}

export function OrganizerRating({ rating, reviewsCount }: OrganizerRatingProps) {
  const ratingDisplay = rating === null ? '—' : rating.toFixed(1);
  const countDisplay = rating === null ? 0 : reviewsCount;

  return (
    <div className="profile-stats">
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
        <span className="profile-stat-value">{countDisplay}</span>
        <span className="muted">кол-во отзывов</span>
      </div>
    </div>
  );
}
