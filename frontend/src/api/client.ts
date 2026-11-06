// Thin fetch wrapper. The Vite dev server proxies `/api` to the Go backend
// (see vite.config.ts), so relative URLs are sufficient in both dev and prod.

const ENV_BASE_URL = (import.meta.env.VITE_API_BASE_URL ?? '').replace(/\/$/, '');
const API_PREFIX = '/api/v1';

const TOKEN_KEY = 'joinup.token';

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY);
}

export function setToken(token: string | null): void {
  if (token === null) {
    localStorage.removeItem(TOKEN_KEY);
  } else {
    localStorage.setItem(TOKEN_KEY, token);
  }
}

export class ApiError extends Error {
  status: number;
  body?: unknown;
  constructor(status: number, message: string, body?: unknown) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.body = body;
  }
}

export interface RequestOptions extends Omit<RequestInit, 'body'> {
  body?: unknown;
  searchParams?: Record<string, string | number | undefined | null | string[] | number[]>;
  rawHeaders?: Record<string, string>;
}

function buildUrl(
  path: string,
  searchParams?: RequestOptions['searchParams'],
): string {
  const base = ENV_BASE_URL ? `${ENV_BASE_URL}${API_PREFIX}` : API_PREFIX;
  const url = `${base}${path.startsWith('/') ? path : `/${path}`}`;
  if (!searchParams) return url;

  const sp = new URLSearchParams();
  for (const [key, value] of Object.entries(searchParams)) {
    if (value === undefined || value === null || value === '') continue;
    if (Array.isArray(value)) {
      for (const item of value) sp.append(key, String(item));
    } else {
      sp.append(key, String(value));
    }
  }
  const query = sp.toString();
  return query ? `${url}?${query}` : url;
}

async function parseBody(res: Response): Promise<unknown> {
  const text = await res.text();
  if (!text) return null;
  try {
    return JSON.parse(text);
  } catch {
    return text;
  }
}

function extractMessage(body: unknown, fallback: string): string {
  if (typeof body === 'string') return body || fallback;
  if (body && typeof body === 'object') {
    const candidate = body as Record<string, unknown>;
    // The Go backend serializes errors as `dto.Msg` ({"msg": "..."}); other
    // shapes are kept for resilience against handlers that don't follow the
    // convention.
    if (typeof candidate.msg === 'string') return candidate.msg;
    if (typeof candidate.message === 'string') return candidate.message;
    if (typeof candidate.error === 'string') return candidate.error;
    if (typeof candidate.detail === 'string') return candidate.detail;
  }
  return fallback;
}

export async function request<T>(
  path: string,
  { body, searchParams, rawHeaders, headers, ...rest }: RequestOptions = {},
): Promise<T> {
  const finalHeaders: Record<string, string> = {
    Accept: 'application/json',
    ...(headers as Record<string, string> | undefined),
  };

  let finalBody: BodyInit | undefined;
  if (body instanceof FormData) {
    finalBody = body;
  } else if (body !== undefined) {
    finalBody = JSON.stringify(body);
    finalHeaders['Content-Type'] = 'application/json';
  }

  const token = getToken();
  if (token && !finalHeaders.Authorization) {
    finalHeaders.Authorization = `Bearer ${token}`;
  }

  if (rawHeaders) {
    for (const [key, value] of Object.entries(rawHeaders)) {
      finalHeaders[key] = value;
    }
  }

  const res = await fetch(buildUrl(path, searchParams), {
    ...rest,
    headers: finalHeaders,
    body: finalBody,
  });

  if (res.status === 401) {
    setToken(null);
  }

  // Some endpoints return a token only via the Authorization header.
  const authHeader = res.headers.get('Authorization');
  const responseToken = authHeader?.startsWith('Bearer ')
    ? authHeader.slice('Bearer '.length)
    : null;

  if (!res.ok) {
    const errBody = await parseBody(res);
    throw new ApiError(
      res.status,
      extractMessage(errBody, `${res.status} ${res.statusText}`),
      errBody,
    );
  }

  if (res.status === 204) {
    return (responseToken ? ({ token: responseToken } as unknown as T) : (null as unknown as T));
  }

  const data = (await parseBody(res)) as T;
  if (responseToken && (data === null || typeof data !== 'object')) {
    return { token: responseToken } as unknown as T;
  }
  if (responseToken && data && typeof data === 'object' && !('token' in (data as object))) {
    (data as Record<string, unknown>).token = responseToken;
  }
  return data;
}
