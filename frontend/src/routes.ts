export type Route =
  | 'login'
  | 'register'
  | 'search'
  | 'mine'
  | 'joining'
  | 'profile'
  | 'event-detail'
  | 'event-new';

export interface RouteState {
  route: Route;
  eventId?: number | null;
  userId?: number | null;
  refreshKey?: number;
}
