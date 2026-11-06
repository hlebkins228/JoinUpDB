import { request } from './client';
import type {
  CreateEventRequest,
  Category,
  EventResponse,
  EventSearchParams,
  ImageResponse,
  UpdateEventRequest,
} from '../types';

interface EventsListResponse {
  events: EventResponse[];
}

interface CategoriesListResponse {
  categories: Array<{ id: number; name: string }>;
}

function toSlug(value: string, id: number): string {
  return value
    .trim()
    .toLowerCase()
    .replace(/\s+/g, '-')
    .replace(/[^a-zа-яё0-9-]/gi, '') || `category-${id}`;
}

function toCategory(raw: { id: number; name: string }): Category {
  return {
    id: raw.id,
    name: raw.name,
    slug: toSlug(raw.name, raw.id),
    hue: (raw.id * 47) % 360,
  };
}

export async function searchEvents(
  params: EventSearchParams = {},
): Promise<EventResponse[]> {
  const data = await request<EventsListResponse>('/user/event/search', {
    method: 'GET',
    searchParams: {
      name: params.name,
      event_from: params.event_from,
      event_to: params.event_to,
      city: params.city,
      category_id: params.category_id,
    },
  });
  return data.events ?? [];
}

export async function getJoinedEvents(): Promise<EventResponse[]> {
  const data = await request<EventsListResponse>('/user/event/joined', {
    method: 'GET',
  });
  return data.events ?? [];
}

export async function getCreatedEvents(): Promise<EventResponse[]> {
  const data = await request<EventsListResponse>('/user/event/created', {
    method: 'GET',
  });
  return data.events ?? [];
}

export async function getCategories(): Promise<Category[]> {
  const data = await request<CategoriesListResponse>('/user/category', {
    method: 'GET',
  });
  return (data.categories ?? []).map(toCategory);
}

export async function getEvent(id: number): Promise<EventResponse> {
  return request<EventResponse>(`/user/event/${id}`, { method: 'GET' });
}

export async function createEvent(body: CreateEventRequest): Promise<{ id: number }> {
  return request<{ id: number }>('/user/event', {
    method: 'POST',
    body,
  });
}

export async function updateEvent(body: UpdateEventRequest): Promise<void> {
  await request<void>(`/user/event/${body.id}`, {
    method: 'PUT',
    body,
  });
}

export async function deleteEvent(id: number): Promise<void> {
  await request<void>(`/user/event/${id}`, { method: 'DELETE' });
}

export async function joinEvent(id: number): Promise<void> {
  await request<void>(`/user/event/${id}/join`, { method: 'POST' });
}

export async function leaveEvent(id: number): Promise<void> {
  await request<void>(`/user/event/${id}/join`, { method: 'DELETE' });
}

export async function uploadEventImage(
  eventId: number,
  file: File,
): Promise<ImageResponse> {
  const form = new FormData();
  form.append('image', file, file.name);
  return request<ImageResponse>(`/user/event/${eventId}/image`, {
    method: 'PUT',
    body: form,
  });
}

export async function addEventCategory(
  eventId: number,
  categoryId: number,
): Promise<void> {
  await request<void>(`/user/event/${eventId}/category`, {
    method: 'POST',
    body: { category_id: categoryId },
  });
}
