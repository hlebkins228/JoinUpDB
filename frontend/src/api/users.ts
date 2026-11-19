import { request } from './client';
import type {
  CreateUserRequest,
  ImageResponse,
  UpdateUserRequest,
  UserResponse,
} from '../types';

export async function createUser(body: CreateUserRequest): Promise<{ id: number }> {
  return request<{ id: number }>('/users', {
    method: 'POST',
    body,
  });
}

export async function getUser(id: number): Promise<UserResponse> {
  return request<UserResponse>(`/users/${id}`, { method: 'GET' });
}

export async function updateUser(body: UpdateUserRequest): Promise<void> {
  await request<void>('/users/me', {
    method: 'PUT',
    body,
  });
}

export async function uploadUserImage(file: File): Promise<ImageResponse> {
  const form = new FormData();
  form.append('image', file, file.name);
  return request<ImageResponse>('/images', {
    method: 'POST',
    body: form,
  });
}
