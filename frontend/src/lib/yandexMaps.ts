// Single-shot loader for the Yandex Maps JS API. The script is injected on
// first use, deduplicated across components, and resolves once `ymaps.ready`
// fires. Subsequent calls return the cached promise (and therefore the same
// `ymaps` namespace), which keeps multiple `<MiniMap>` / `<AddressPicker>`
// instances from racing each other.

// eslint-disable-next-line @typescript-eslint/no-explicit-any
type Ymaps = any;

let pending: Promise<Ymaps> | null = null;

export class YandexMapsConfigError extends Error {
  constructor(message: string) {
    super(message);
    this.name = 'YandexMapsConfigError';
  }
}

export function getYandexApiKey(): string {
  return (import.meta.env.VITE_YANDEX_MAPS_API_KEY ?? '').trim();
}

// Yandex requires a separate "suggest" key to enable the SuggestView widget.
// Without it the JS API loads but `new ymaps.SuggestView(...)` produces no
// suggestions. Keys are issued in the same developer cabinet as the JS API
// key but for the "Geosuggest" product.
export function getYandexSuggestApiKey(): string {
  return (import.meta.env.VITE_YANDEX_SUGGEST_API_KEY ?? '').trim();
}

export function loadYandexMaps(): Promise<Ymaps> {
  if (typeof window === 'undefined') {
    return Promise.reject(
      new Error('Yandex Maps API доступен только в браузере'),
    );
  }

  if (window.ymaps && typeof window.ymaps.ready === 'function') {
    return new Promise((resolve, reject) => {
      window.ymaps.ready(() => resolve(window.ymaps), reject);
    });
  }

  if (pending) return pending;

  const apiKey = getYandexApiKey();
  if (!apiKey) {
    return Promise.reject(
      new YandexMapsConfigError(
        'не задан ключ Yandex Maps (VITE_YANDEX_MAPS_API_KEY)',
      ),
    );
  }
  const suggestKey = getYandexSuggestApiKey();
  if (!suggestKey) {
    // Hard fail with a clear message — without this key SuggestView is silent
    // (no errors thrown, no requests made) which makes the missing dropdown
    // very confusing to debug.
    // eslint-disable-next-line no-console
    console.warn(
      '[yandexMaps] VITE_YANDEX_SUGGEST_API_KEY is not set — the address suggest dropdown will be empty. ' +
        'Add the key to .env.local and restart `npm run dev`.',
    );
  }

  pending = new Promise<Ymaps>((resolve, reject) => {
    const existing = document.querySelector<HTMLScriptElement>(
      'script[data-yandex-maps-loader="1"]',
    );
    const script = existing ?? document.createElement('script');
    if (!existing) {
      // The exact URL format documented for SuggestView. We deliberately keep
      // it minimal — adding `package.full` here triggered "module already
      // loaded" warnings in some setups.
      const params = new URLSearchParams({
        lang: 'ru_RU',
        apikey: apiKey,
        load: 'SuggestView,Map,Placemark,geocode,geoObject.addon.balloon,geoObject.addon.hint,control.ZoomControl',
      });
      if (suggestKey) params.set('suggest_apikey', suggestKey);
      script.src = `https://api-maps.yandex.ru/2.1/?${params.toString()}`;
      script.async = true;
      script.dataset.yandexMapsLoader = '1';
      // eslint-disable-next-line no-console
      console.info('[yandexMaps] loading script:', script.src);
    }
    script.addEventListener('load', () => {
      if (!window.ymaps) {
        pending = null;
        reject(new Error('Yandex Maps API не инициализирован'));
        return;
      }
      window.ymaps.ready(() => {
        // eslint-disable-next-line no-console
        console.info(
          '[yandexMaps] ready, SuggestView available:',
          typeof window.ymaps.SuggestView === 'function',
        );
        resolve(window.ymaps);
      }, reject);
    });
    script.addEventListener('error', () => {
      pending = null;
      reject(new Error('не удалось загрузить Yandex Maps API'));
    });
    if (!existing) document.head.appendChild(script);
  });

  return pending;
}
