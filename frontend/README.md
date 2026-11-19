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

## API contract

The canonical OpenAPI 3.1 specification is [openapi.yaml](../app/api/openapi.yaml).
The SPA uses resource routes under `/api/v1`: `/tokens`, `/users`, `/events`,
`/categories`, `/images`, and nested event participants and reviews.
Authentication uses `POST /tokens` with JSON credentials. Mutations use Bearer JWT.
The backend serves the contract at `/openapi.yaml` and Swagger UI at `/swagger/index.html`.
See the root [README](../README.md) for laboratory 2 and offline contract validation.
