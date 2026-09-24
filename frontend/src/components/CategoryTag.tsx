import type { CSSProperties } from 'react';
import { categoryColor, findCategory } from '../lib/constants';
import type { Category } from '../types';

interface CategoryTagProps {
  category: Category | number | string;
  active?: boolean;
  toggle?: boolean;
  onClick?(): void;
}

export function CategoryTag({ category, active = true, toggle, onClick }: CategoryTagProps) {
  const cat: Category | undefined =
    typeof category === 'object' ? category : findCategory(category);
  if (!cat) return null;

  const colors = categoryColor(cat);
  const style: CSSProperties = active
    ? { background: colors.bg, color: colors.fg, borderColor: colors.border }
    : { background: 'transparent', color: 'var(--fg-50)', borderColor: 'var(--fg-30)' };

  const cls = ['tag', toggle ? 'tag-toggle' : '', !active ? 'is-off' : '']
    .filter(Boolean)
    .join(' ');

  return (
    <span className={cls} style={style} onClick={onClick} role={onClick ? 'button' : undefined}>
      {cat.name}
    </span>
  );
}
