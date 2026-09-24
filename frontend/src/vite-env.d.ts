/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_API_BASE_URL?: string;
  readonly VITE_API_PROXY?: string;
  readonly VITE_YANDEX_MAPS_API_KEY?: string;
  readonly VITE_YANDEX_SUGGEST_API_KEY?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}

// The Yandex Maps JS API attaches `ymaps` to `window` after the script
// loads. The API surface is large and dynamically typed, so we expose a
// permissive `any` alias here and use it through the loader.
interface Window {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  ymaps?: any;
}
