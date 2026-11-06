// Mirrors the backend DTOs in app/internal/controller/dto.

export interface UserResponse {
  id: number;
  name: string;
  age: number;
  login: string;
  created_at: string;
  city: string;
  telegram_login?: string;
  avatar_id?: number | null;
  role: string;
  created_events_count: number;
  joined_events_count: number;
  organizer_rating: number | null;
  organizer_reviews_count: number;
}

export interface CreateUserRequest {
  name: string;
  age: number;
  login: string;
  password: string;
  city: string;
  telegram_login?: string;
  avatar_id?: number;
}

export interface UpdateUserRequest {
  name: string;
  age: number;
  city: string;
  telegram_login?: string;
  avatar_id?: number;
}

export interface EventLocation {
  name: string;
  longitude: number;
  latitude: number;
  address: string;
}

export interface EventResponse {
  id: number;
  creator_id: number;
  name: string;
  desc?: string;
  created_at: string;
  updated_at: string;
  event_time: string;
  telegram_chat_url?: string;
  city: string;
  creator?: EventMember;
  members?: EventMember[];
  categories?: Category[];
  location: EventLocation | null;
  image_id?: number | null;
  deleted: boolean;
}

export interface EventMember {
  id: number;
  name: string;
  avatar_id?: number | null;
}

export interface CreateEventRequest {
  name: string;
  desc?: string;
  event_time: string;
  telegram_chat_url?: string;
  city: string;
  location: EventLocation;
  image_id?: number;
}

export interface UpdateEventRequest extends CreateEventRequest {
  id: number;
  category_ids?: number[];
}

export interface EventSearchParams {
  name?: string;
  event_from?: string;
  event_to?: string;
  city?: string;
  category_id?: number[];
}

export interface AuthResponse {
  token: string;
}

export interface ImageResponse {
  id: number;
}

export interface Category {
  id: number;
  slug?: string;
  name: string;
  hue?: number;
}
