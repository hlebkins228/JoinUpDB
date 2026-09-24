import type { CSSProperties, MouseEventHandler } from 'react';
import { initials } from '../lib/format';

type Size = 'sm' | 'md' | 'lg' | 'event';

interface AvatarProps {
  src?: string | null;
  name?: string;
  size?: Size;
  onClick?: MouseEventHandler<HTMLSpanElement>;
  className?: string;
  style?: CSSProperties;
  title?: string;
}

const SIZE_CLASS: Record<Size, string> = {
  sm: 'avatar',
  md: 'avatar',
  lg: 'avatar avatar-lg',
  event: 'avatar avatar-event',
};

export function Avatar({
  src,
  name,
  size = 'md',
  onClick,
  className = '',
  style,
  title,
}: AvatarProps) {
  const cls = `${SIZE_CLASS[size]} ${className}`.trim();
  const finalStyle: CSSProperties = {
    cursor: onClick ? 'pointer' : 'default',
    ...style,
  };
  return (
    <span
      className={cls}
      onClick={onClick}
      style={finalStyle}
      title={title ?? name ?? ''}
    >
      {src ? <img src={src} alt={name ?? ''} /> : <span>{initials(name)}</span>}
    </span>
  );
}
