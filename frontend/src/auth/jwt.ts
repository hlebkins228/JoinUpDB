// The auth handler returns an HS256 JWT signed by the Go backend. We can't
// verify the signature in the browser without the secret, but we can read
// the payload to extract `UserID` / `Role`. This is fine for client-side
// gating: every protected request still goes through the backend with the
// raw token, which is the only place the signature is checked.
export interface JwtPayload {
  UserID: number;
  Role: string;
  exp?: number;
  [key: string]: unknown;
}

function base64UrlDecode(input: string): string {
  const pad = input.length % 4 === 0 ? '' : '='.repeat(4 - (input.length % 4));
  const base64 = (input + pad).replace(/-/g, '+').replace(/_/g, '/');
  if (typeof atob === 'function') {
    return atob(base64);
  }
  return Buffer.from(base64, 'base64').toString('binary');
}

export function decodeJwt(token: string): JwtPayload | null {
  try {
    const [, payload] = token.split('.');
    if (!payload) return null;
    const decoded = base64UrlDecode(payload);
    const json = decodeURIComponent(
      decoded
        .split('')
        .map((c) => `%${c.charCodeAt(0).toString(16).padStart(2, '0')}`)
        .join(''),
    );
    return JSON.parse(json) as JwtPayload;
  } catch {
    return null;
  }
}

export function isExpired(payload: JwtPayload | null): boolean {
  if (!payload?.exp) return false;
  return payload.exp * 1000 <= Date.now();
}
