import { Avatar } from './Avatar';
import { imageUrl } from '../lib/imageUrl';

interface UserChipProps {
  user: {
    id: number;
    name: string;
    avatar_id?: number | null;
  };
  onOpenUser?: (id: number) => void;
}

export function UserChip({ user, onOpenUser }: UserChipProps) {
  const content = (
    <>
      <Avatar src={imageUrl(user.avatar_id)} name={user.name} size="sm" />
      <span>{user.name}</span>
    </>
  );

  if (!onOpenUser) {
    return <span className="user-chip">{content}</span>;
  }

  return (
    <button
      type="button"
      className="user-chip user-chip-button"
      onClick={(e) => {
        e.stopPropagation();
        onOpenUser(user.id);
      }}
    >
      {content}
    </button>
  );
}
