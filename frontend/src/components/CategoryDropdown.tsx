import { CategoryTag } from './CategoryTag';
import type { Category } from '../types';

interface CategoryDropdownProps {
  categories: Category[];
  activeCats: number[];
  onToggle(id: number): void;
}

export function CategoryDropdown({ categories, activeCats, onToggle }: CategoryDropdownProps) {
  return (
    <details className="category-dropdown">
      <summary className="select">
        {activeCats.length ? `выбрано: ${activeCats.length}` : 'все категории'}
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
