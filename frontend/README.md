# JoinUp · frontend

React + TypeScript + Vite implementation of the JoinUp UI. The visual system
(colors, typography, layout) is ported from the bundled HTML mock that lived
at `frontend/index.html` previously — the original prototype was a single
self-contained HTML/Babel page; this project replaces it with a real React
app that talks to the Go backend in `app/`.

## Stack

- React 18 + TypeScript
- Vite (dev server + production build)
- ESLint with `@typescript-eslint` and `react-hooks` plugins
- No external UI/CSS library — the visual system is a single `styles.css`
  ported 1:1 from the original prototype

## Local development

```bash
cd frontend
npm install
npm run dev    # starts Vite on http://localhost:5173
```

The Vite dev server proxies any request to `/api/**` to the Go backend
(default `http://localhost:8080`). Override the target with
`VITE_API_PROXY=http://example.org` (see `.env.example`).

In production the bundle is expected to be served from the same origin as
the API, so all fetches use relative `/api/v1/...` URLs by default. Set
`VITE_API_BASE_URL` if the frontend needs to point at a different host.

`VITE_YANDEX_MAPS_API_KEY` is required for the address picker (event
creation/edit) and the map preview on the event detail screen. Get a key
in the [Yandex developer cabinet](https://developer.tech.yandex.ru/) for
the **JavaScript API and Geocoder** product. The address suggest dropdown
also needs a separate **Geosuggest API** key — set it as
`VITE_YANDEX_SUGGEST_API_KEY`. Put both into `.env.local`:

```
VITE_YANDEX_MAPS_API_KEY=your-js-api-key
VITE_YANDEX_SUGGEST_API_KEY=your-suggest-key
```

Without the keys the picker and the preview render a fallback message but
the rest of the UI keeps working.

## Scripts

| command         | description                              |
| --------------- | ---------------------------------------- |
| `npm run dev`   | Vite dev server with HMR                 |
| `npm run build` | TypeScript type-check + production build |
| `npm run lint`  | ESLint over `src/**`                     |

## Layout

```
src/
├── main.tsx                # bootstrap
├── App.tsx                 # auth gate + tiny in-memory router
├── styles.css              # design system from the prototype
├── api/                    # fetch wrappers per resource
├── auth/                   # JWT decode + AuthProvider/useAuth
├── components/             # Header, Avatar, EventCard, inputs, …
├── lib/                    # constants (cities/categories), formatters
└── screens/                # Login, Register, Search, MyEvents, Joining,
                            # Profile, EventDetail (view + edit)
```

## Backend endpoints used

All requests live under `/api/v1`:

| method | path                              | usage                           |
| ------ | --------------------------------- | ------------------------------- |
| GET    | `/auth`                           | login (login/password headers)  |
| POST   | `/user`                           | register                        |
| GET    | `/user/:id`                       | profile lookup (after JWT)      |
| PUT    | `/user`                           | profile update                  |
| POST   | `/user/image`                     | upload user avatar              |
| GET    | `/user/event/search`              | search events                   |
| GET    | `/user/event/:id`                 | event detail                    |
| POST   | `/user/event`                     | create event                    |
| PUT    | `/user/event/:id`                 | update event                    |
| DELETE | `/user/event/:id`                 | delete event                    |
| POST   | `/user/event/:id/join`            | join event                      |
| PUT    | `/user/event/:id/image`           | upload event cover              |
| POST   | `/user/event/:id/category`        | attach a category               |

## Missing backend endpoints

Some flows in the UI cannot be fully implemented because the matching
backend handlers do not exist yet. The frontend keeps the UI in place and
falls back to placeholder messaging where data is missing — none of these
endpoints have been implemented in this PR (per the task instructions).

- `GET /image/:id` — there is no public route to read uploaded images. The
  app stores the `image_id` returned from upload but cannot fetch the
  binary back, so avatars and event covers fall back to initials /
  placeholder until the user re-uploads in the same session.
- Listing endpoints scoped to a user. `searchEvents` is the only listing
  route, so `Мои события` (events I created) and `Я участвую` (events I
  joined) are filtered on the client. There is no creator filter on the
  search route, and the response does not populate `members`, which means
  the «participates» list cannot be reconstructed from the API today.
- `DELETE /user/event/:id/join` — there is no leave-event handler, so the
  UI omits the «отписаться» action.
- `GET /user/event/categories` (and `GET /user/event/:id/categories`) —
  the categories directory is not exposed and the event response does not
  contain attached category ids, so the UI uses a hard-coded list and
  cannot show which categories are already attached to an event.
- `GET /user/me` — the JWT carries `UserID`, which is decoded in the
  browser, but there is no first-class «who am I» endpoint. `GET /user/:id`
  is used as a workaround.
