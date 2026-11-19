import { request } from './client';
import type { AuthResponse } from '../types';

export async function login(loginValue: string, password: string): Promise<AuthResponse> {
  return request<AuthResponse>('/tokens', {
    method: 'POST',
    body: {
      login: loginValue,
      password,
    },
  });
}
