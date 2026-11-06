import type { EventResponse } from '../types';
import { Avatar } from './Avatar';
import { CategoryTag } from './CategoryTag';
import { UserChip } from './UserChip';
import { fmtDateTime, isPastDate } from '../lib/format';
import { imageUrl } from '../lib/imageUrl';

interface EventCardProps {
  event: EventResponse;
  onOpen(id: number): void;
  onOpenUser?: (id: number) => void;
  showParticipantUsers?: boolean;
}

export function EventCard({
  event,
  onOpen,
  onOpenUser,
  showParticipantUsers = true,
}: EventCardProps) {
  const cover = imageUrl(event.image_id);
  const isPast = isPastDate(event.event_time);
  const creator = event.creator ?? { id: event.creator_id, name: `#${event.creator_id}` };
  const membersCount = event.members?.length ?? 0;

  function openCard() {
    onOpen(event.id);
  }

  return (
    <div
      className={`event-card ${isPast ? 'is-past' : ''}`}
      onClick={openCard}
      onKeyDown={(e) => {
        if (e.key === 'Enter' || e.key === ' ') {
          e.preventDefault();
          openCard();
        }
      }}
      role="button"
      tabIndex={0}
      aria-label={event.name}
    >
      <Avatar src={cover} name={event.name} size="event" />
      <div className="event-card-meta">
        <h3 className="event-card-title">{event.name}</h3>
        <div className="event-card-line">
          <span>{fmtDateTime(event.event_time)}</span>
          <span>·</span>
          <span>{event.city}</span>
          {event.location?.name ? (
            <>
              <span>·</span>
              <span>{event.location.name}</span>
            </>
          ) : null}
        </div>
        {event.categories?.length ? (
          <div className="event-card-tags">
            {event.categories.map((category) => (
              <CategoryTag key={category.id} category={category} />
            ))}
          </div>
        ) : null}
      </div>
      <div className="event-card-side">
        <div className="event-card-users">
          <span className="label">создатель</span>
          <UserChip user={creator} onOpenUser={onOpenUser} />
          <span className="label">участники</span>
          {showParticipantUsers && membersCount > 0 ? (
            <div className="event-card-members">
              {event.members?.slice(0, 3).map((member) => (
                <UserChip key={member.id} user={member} onOpenUser={onOpenUser} />
              ))}
              {membersCount > 3 && (
                <span className="muted">+{membersCount - 3}</span>
              )}
            </div>
          ) : (
            <span>{membersCount}</span>
          )}
        </div>
      </div>
    </div>
  );
}
