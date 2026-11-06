import { CategoryTag } from './CategoryTag';
import type { Category } from '../types';

interface CategoryDropdownProps {
  categories: Category[];
  activeCats: number[];
  emptyLabel?: string;
  onToggle(id: number): void;
}

export function CategoryDropdown({ categories, activeCats, emptyLabel = 'все категории', onToggle }: CategoryDropdownProps) {
  return (
    <details className="category-dropdown">
      <summary className="select">
        {activeCats.length ? `выбрано: ${activeCats.length}` : emptyLabel}
      </summary>
      <div className="category-dropdown-list filters-cats">
        {categories.map((category) => (
          <CategoryTag
            key={category.id}
            category={category}
            toggle
            active={activeCats.includes(category.id)}
            onClick={() => onToggle(category.id)}
          />
        ))}
      </div>
    </details>
  );
}
