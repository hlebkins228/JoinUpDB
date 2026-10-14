import { useRef, useState } from 'react';
import { CategoryTag } from './CategoryTag';
import { TextInput } from './inputs';
import type { Category } from '../types';

interface CategoryPickerProps {
  categories: Category[];
  activeCats: number[];
  onChange(ids: number[]): void;
}

export function CategoryPicker({ categories, activeCats, onChange }: CategoryPickerProps) {
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState('');
  const [draftCats, setDraftCats] = useState<number[]>([]);
  const triggerRef = useRef<HTMLButtonElement>(null);
  const search = query.trim().toLowerCase();
  const filteredCategories = categories.filter((category) =>
    category.name.toLowerCase().includes(search),
  );

  function close() {
    setOpen(false);
    triggerRef.current?.focus();
  }

  function toggleCat(id: number) {
    setDraftCats((prev) =>
      prev.includes(id) ? prev.filter((value) => value !== id) : [...prev, id],
    );
  }

  return (
    <div
      className="category-picker"
      onKeyDown={(e) => {
        if (open && e.key === 'Escape') {
          e.preventDefault();
          e.stopPropagation();
          close();
        }
      }}
    >
      <button
        ref={triggerRef}
        type="button"
        className="select category-picker-trigger"
        aria-label="выбрать категории"
        aria-expanded={open}
        onClick={() => {
          if (open) {
            close();
          } else {
            setDraftCats(activeCats);
            setQuery('');
            setOpen(true);
          }
        }}
      >
        <span>{activeCats.length ? `выбрано: ${activeCats.length}` : 'без категорий'}</span>
        <span aria-hidden="true">{open ? '▴' : '▾'}</span>
      </button>
      {open && (
        <div className="category-picker-panel">
          <TextInput
            value={query}
            onChange={setQuery}
            placeholder="найти категорию"
            aria-label="поиск категорий"
            autoFocus
            onKeyDown={(e) => {
              if (e.key === 'Enter') e.preventDefault();
            }}
          />
          <div className="filters-cats category-picker-list" role="group" aria-label="категории">
            {filteredCategories.map((category) => (
              <button
                key={category.id}
                type="button"
                className="category-picker-option"
                aria-pressed={draftCats.includes(category.id)}
                onClick={() => toggleCat(category.id)}
              >
                <CategoryTag category={category} toggle active={draftCats.includes(category.id)} />
              </button>
            ))}
            {filteredCategories.length === 0 && <span className="muted">ничего не найдено</span>}
          </div>
          <span className="muted">выбрано: {draftCats.length}</span>
          <div className="row-end">
            <button type="button" className="btn btn-sm btn-ghost" onClick={close}>
              отмена
            </button>
            <button
              type="button"
              className="btn btn-sm"
              onClick={() => {
                onChange(draftCats);
                close();
              }}
            >
              ок
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
