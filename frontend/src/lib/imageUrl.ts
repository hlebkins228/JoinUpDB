const ENV_BASE_URL = (import.meta.env.VITE_API_BASE_URL ?? '').replace(/\/$/, '');
const API_PREFIX = '/api/v1';

export function imageUrl(id: number | null | undefined): string | null {
  if (id == null) return null;
  const base = ENV_BASE_URL ? `${ENV_BASE_URL}${API_PREFIX}` : API_PREFIX;
  return `${base}/image/${id}`;
}
