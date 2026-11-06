import { request } from './client';

export interface ReviewAuthor {
  id: number;
  name: string;
  avatar_id: number | null;
}

export interface Review {
  id: number;
  event_id: number;
  author_id: number;
  rating: number;
  text: string;
  created_at: string;
  author: ReviewAuthor;
}

interface ReviewsResponse {
  reviews: Review[];
}

export async function listEventReviews(eventId: number): Promise<Review[]> {
  const data = await request<ReviewsResponse>(`/event/${eventId}/review`, {
    method: 'GET',
  });
  return data.reviews ?? [];
}

export async function createEventReview(
  eventId: number,
  body: { rating: number; text: string },
): Promise<Review> {
  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), 10000);
  try {
    return await request<Review>(`/user/event/${eventId}/review`, {
      method: 'POST',
      body,
      signal: controller.signal,
    });
  } finally {
    clearTimeout(timer);
  }
}
