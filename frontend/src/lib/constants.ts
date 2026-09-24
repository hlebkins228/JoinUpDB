import type { Category } from '../types';

// Hard-coded city list — the backend stores cities as plain strings and does
// not expose a directory endpoint, so we replicate the prototype's options.
export const CITIES = [
  'Москва',
  'Санкт-Петербург',
  'Казань',
  'Новосибирск',
  'Екатеринбург',
  'Нижний Новгород',
  'Краснодар',
  'Самара',
  'Ростов-на-Дону',
  'Уфа',
  'Воронеж',
  'Пермь',
  'Волгоград',
  'Калининград',
  'Сочи',
];

export function findCategory(idOrSlug: number | string): Category | undefined {
  void idOrSlug;
  return undefined;
}

export function categoryColor(c: Category): { fg: string; bg: string; border: string } {
  const hue = c.hue ?? (c.id * 47) % 360;
  return {
    fg: `oklch(0.30 0.10 ${hue})`,
    bg: `oklch(0.95 0.04 ${hue})`,
    border: `oklch(0.80 0.08 ${hue})`,
  };
}
