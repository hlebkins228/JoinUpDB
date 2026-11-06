import { request } from './client';
import type { AuthResponse } from '../types';

// `GET /api/v1/auth` reads the credentials from custom headers.
export async function login(loginValue: string, password: string): Promise<AuthResponse> {
  return request<AuthResponse>('/auth', {
    method: 'GET',
    rawHeaders: {
      login: loginValue,
      password,
    },
  });
}
