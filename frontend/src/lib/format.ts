const MONTHS = [
  'янв', 'фев', 'мар', 'апр', 'мая', 'июн',
  'июл', 'авг', 'сен', 'окт', 'ноя', 'дек',
];

export function fmtDate(iso?: string | null): string {
  if (!iso) return '—';
  const d = parseServerDate(iso);
  if (Number.isNaN(d.getTime())) return '—';
  return `${d.getDate()} ${MONTHS[d.getMonth()]} ${d.getFullYear()}`;
}

export function fmtTime(iso?: string | null): string {
  if (!iso) return '';
  const d = parseServerDate(iso);
  if (Number.isNaN(d.getTime())) return '';
  const hh = String(d.getHours()).padStart(2, '0');
  const mm = String(d.getMinutes()).padStart(2, '0');
  return `${hh}:${mm}`;
}

export function fmtDateTime(iso?: string | null): string {
  if (!iso) return '—';
  return `${fmtDate(iso)} · ${fmtTime(iso)}`;
}

export function toInputDate(iso?: string | null): string {
  if (!iso) return '';
  const d = parseServerDate(iso);
  if (Number.isNaN(d.getTime())) return '';
  const yyyy = d.getFullYear();
  const mm = String(d.getMonth() + 1).padStart(2, '0');
  const dd = String(d.getDate()).padStart(2, '0');
  return `${yyyy}-${mm}-${dd}`;
}

export function toInputDateTime(iso?: string | null): string {
  if (!iso) return '';
  const d = parseServerDate(iso);
  if (Number.isNaN(d.getTime())) return '';
  const yyyy = d.getFullYear();
  const mm = String(d.getMonth() + 1).padStart(2, '0');
  const dd = String(d.getDate()).padStart(2, '0');
  const hh = String(d.getHours()).padStart(2, '0');
  const mi = String(d.getMinutes()).padStart(2, '0');
  return `${yyyy}-${mm}-${dd}T${hh}:${mi}`;
}

export function fromInputDateTime(value: string): string {
  const d = new Date(value);
  if (Number.isNaN(d.getTime())) return toLocalOffsetIso(new Date());
  return `${value.length === 16 ? `${value}:00` : value}${formatOffset(d)}`;
}

export function fromInputDate(value: string): string | undefined {
  if (!value) return undefined;
  const d = new Date(`${value}T00:00:00`);
  if (Number.isNaN(d.getTime())) return undefined;
  return `${value}T00:00:00${formatOffset(d)}`;
}

export function currentInputDateTime(): string {
  return toInputDateTime(toLocalOffsetIso(new Date()));
}

export function isPastDate(iso?: string | null): boolean {
  if (!iso) return false;
  const d = parseServerDate(iso);
  return !Number.isNaN(d.getTime()) && d.getTime() < Date.now();
}

export function initials(name?: string | null): string {
  if (!name) return '??';
  const parts = name.trim().split(/\s+/).slice(0, 2);
  return parts
    .map((p) => p[0])
    .filter(Boolean)
    .join('')
    .toUpperCase();
}

function toLocalOffsetIso(date: Date): string {
  const yyyy = date.getFullYear();
  const mm = String(date.getMonth() + 1).padStart(2, '0');
  const dd = String(date.getDate()).padStart(2, '0');
  const hh = String(date.getHours()).padStart(2, '0');
  const mi = String(date.getMinutes()).padStart(2, '0');
  const ss = String(date.getSeconds()).padStart(2, '0');
  return `${yyyy}-${mm}-${dd}T${hh}:${mi}:${ss}${formatOffset(date)}`;
}

function parseServerDate(iso: string): Date {
  const hasTimezone = /(?:z|[+-]\d{2}:?\d{2})$/i.test(iso);
  return new Date(hasTimezone ? iso : `${iso}Z`);
}

function formatOffset(date: Date): string {
  const offset = -date.getTimezoneOffset();
  const sign = offset >= 0 ? '+' : '-';
  const abs = Math.abs(offset);
  const hh = String(Math.floor(abs / 60)).padStart(2, '0');
  const mm = String(abs % 60).padStart(2, '0');
  return `${sign}${hh}:${mm}`;
}
