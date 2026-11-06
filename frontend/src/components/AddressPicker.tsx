import { useCallback, useEffect, useId, useRef, useState } from 'react';
import {
  YandexMapsConfigError,
  loadYandexMaps,
} from '../lib/yandexMaps';

export interface PickedAddress {
  name: string;
  address: string;
  latitude: number;
  longitude: number;
}

interface AddressPickerProps {
  // The current address text (controlled — matches `form.address`).
  address: string;
  // Coordinates of the current selection. Drives the marker.
  latitude: number | null;
  longitude: number | null;
  // Optional bias for suggestions / geocoding (e.g. the chosen city).
  city?: string;
  // Fired when the user picks a suggestion or interacts with the map.
  onPick(value: PickedAddress): void;
  // Fired on every keystroke so the parent stays in sync.
  onAddressChange(value: string): void;
  placeholder?: string;
}

// Default centre for new sessions. Roughly Moscow centre.
const DEFAULT_CENTER: [number, number] = [55.7558, 37.6173];
const DEFAULT_ZOOM = 11;
const PICKED_ZOOM = 16;

// Reads the formatted address from a Yandex GeoObject returned by the
// geocoder. `getAddressLine()` is the canonical helper; we keep a fallback to
// `properties.text` for safety across API versions.
// eslint-disable-next-line @typescript-eslint/no-explicit-any
function readAddress(geoObject: any): string {
  if (!geoObject) return '';
  try {
    if (typeof geoObject.getAddressLine === 'function') {
      const line = geoObject.getAddressLine();
      if (typeof line === 'string' && line.trim()) return line.trim();
    }
    const props = geoObject.properties;
    const text = props?.get?.('text');
    if (typeof text === 'string' && text.trim()) return text.trim();
  } catch {
    /* fall through */
  }
  return '';
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function readKind(geoObject: any): string {
  try {
    return (
      geoObject?.properties?.get?.('metaDataProperty.GeocoderMetaData.kind') ??
      ''
    );
  } catch {
    return '';
  }
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function readShortName(geoObject: any): string {
  try {
    const name = geoObject?.properties?.get?.('name');
    if (typeof name === 'string' && name.trim()) return name.trim();
  } catch {
    /* fall through */
  }
  return '';
}

// Pick the best label to seed the editable "name" field with. For house
// results that's the full address; for venues / districts / metro the short
// Yandex name (e.g. "ТЦ Авиапарк").
// eslint-disable-next-line @typescript-eslint/no-explicit-any
function deriveName(geoObject: any, address: string): string {
  const kind = readKind(geoObject);
  const short = readShortName(geoObject);
  if (short && kind && kind !== 'house' && kind !== 'street') return short;
  if (address) return address;
  return short;
}

export function AddressPicker({
  address,
  latitude,
  longitude,
  city,
  onPick,
  onAddressChange,
  placeholder = 'Москва, ул. Тверская, 1',
}: AddressPickerProps) {
  const inputId = useId();
  const inputRef = useRef<HTMLInputElement | null>(null);
  const mapContainerRef = useRef<HTMLDivElement | null>(null);
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const mapRef = useRef<any>(null);
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const placemarkRef = useRef<any>(null);
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const ymapsRef = useRef<any>(null);
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const suggestRef = useRef<any>(null);
  const onPickRef = useRef(onPick);

  const [error, setError] = useState<string | null>(null);
  const [ready, setReady] = useState(false);
  const [diagnostic, setDiagnostic] = useState<string | null>(null);

  useEffect(() => {
    onPickRef.current = onPick;
  }, [onPick]);

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const reverseGeocode = useCallback((ymaps: any, coords: [number, number]) => {
    ymaps
      .geocode(coords, { results: 1 })
      .then((res: { geoObjects: { get: (i: number) => unknown } }) => {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        const obj = res.geoObjects.get(0) as any;
        const addressText = readAddress(obj);
        const nameText = deriveName(obj, addressText);
        onPickRef.current({
          address: addressText,
          name: nameText || addressText,
          latitude: coords[0],
          longitude: coords[1],
        });
      })
      .catch(() => {
        onPickRef.current({
          address: '',
          name: '',
          latitude: coords[0],
          longitude: coords[1],
        });
      });
  }, []);

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const forwardGeocode = useCallback((ymaps: any, query: string) => {
    ymaps
      .geocode(query, { results: 1 })
      .then((res: { geoObjects: { get: (i: number) => unknown } }) => {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        const obj = res.geoObjects.get(0) as any;
        if (!obj) return;
        const coords = obj.geometry.getCoordinates() as [number, number];
        const addressText = readAddress(obj) || query;
        const nameText = deriveName(obj, addressText) || query;
        onPickRef.current({
          address: addressText,
          name: nameText,
          latitude: coords[0],
          longitude: coords[1],
        });
      })
      .catch(() => {
        /* swallow — fields stay as they were */
      });
  }, []);

  // Bootstrap the map + the SuggestView. Only mounts once; coordinate / city
  // changes flow in via the effects below.
  useEffect(() => {
    let cancelled = false;
    loadYandexMaps()
      .then((ymaps) => {
        if (cancelled) return;
        if (!mapContainerRef.current || !inputRef.current) return;
        ymapsRef.current = ymaps;

        const initialCenter: [number, number] =
          latitude != null && longitude != null
            ? [latitude, longitude]
            : DEFAULT_CENTER;
        const initialZoom =
          latitude != null && longitude != null ? PICKED_ZOOM : DEFAULT_ZOOM;

        const map = new ymaps.Map(
          mapContainerRef.current,
          {
            center: initialCenter,
            zoom: initialZoom,
            controls: ['zoomControl'],
          },
          { suppressMapOpenBlock: true },
        );
        mapRef.current = map;

        if (latitude != null && longitude != null) {
          const placemark = new ymaps.Placemark(
            initialCenter,
            {},
            { preset: 'islands#redIcon', draggable: true },
          );
          placemark.events.add('dragend', () => {
            const coords = placemark.geometry.getCoordinates() as [
              number,
              number,
            ];
            reverseGeocode(ymaps, coords);
          });
          map.geoObjects.add(placemark);
          placemarkRef.current = placemark;
        }

        map.events.add('click', (e: { get: (k: string) => unknown }) => {
          const coords = e.get('coords') as [number, number];
          reverseGeocode(ymaps, coords);
        });

        // The native Yandex SuggestView. It binds to the <input id="..."> and
        // injects its own floating dropdown via `document.body`. The event
        // we care about is `select` — when the user picks an item we kick off
        // a forward geocode to resolve coordinates and the canonical address.
        if (typeof ymaps.SuggestView !== 'function') {
          setDiagnostic(
            'модуль SuggestView не загружен — проверьте параметр `load` в URL загрузки API',
          );
        } else {
          try {
            const suggest = new ymaps.SuggestView(inputRef.current, {
              results: 7,
            });
            suggest.events.add(
              'select',
              (e: { get: (k: string) => unknown }) => {
                const item = e.get('item') as { value?: string } | undefined;
                const value = item?.value;
                if (typeof value !== 'string' || !value.trim()) return;
                forwardGeocode(ymaps, value.trim());
              },
            );
            // eslint-disable-next-line no-console
            console.info(
              '[AddressPicker] SuggestView attached to input',
              inputRef.current,
            );
            suggestRef.current = suggest;
          } catch (err) {
            // eslint-disable-next-line no-console
            console.error('[AddressPicker] SuggestView ctor failed', err);
            setDiagnostic(
              'не удалось инициализировать SuggestView: ' +
                (err instanceof Error ? err.message : String(err)),
            );
          }
        }

        setReady(true);
      })
      .catch((err) => {
        if (cancelled) return;
        if (err instanceof YandexMapsConfigError) {
          setError('карта недоступна: не настроен ключ Yandex Maps');
        } else {
          setError(err instanceof Error ? err.message : String(err));
        }
      });

    return () => {
      cancelled = true;
      if (suggestRef.current) {
        suggestRef.current.destroy();
        suggestRef.current = null;
      }
      if (mapRef.current) {
        mapRef.current.destroy();
        mapRef.current = null;
        placemarkRef.current = null;
      }
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // Bias the SuggestView toward the chosen city using a bounding box. We
  // recompute it every time the city changes — geocoding the city centre and
  // expanding by ~30 km in each direction is good enough for typical events.
  useEffect(() => {
    const ymaps = ymapsRef.current;
    const suggest = suggestRef.current;
    if (!ymaps || !suggest) return;
    const trimmed = city?.trim();
    if (!trimmed) {
      suggest.options.unset('boundedBy');
      suggest.options.unset('strictBounds');
      return;
    }
    let cancelled = false;
    ymaps
      .geocode(trimmed, { results: 1, kind: 'locality' })
      .then((res: { geoObjects: { get: (i: number) => unknown } }) => {
        if (cancelled) return;
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        const obj = res.geoObjects.get(0) as any;
        if (!obj) return;
        const center = obj.geometry.getCoordinates() as [number, number];
        // ~30 km radius. 1° latitude ≈ 111 km, so ~0.27°. Longitude scales
        // with cos(latitude), but we keep the box generous and don't bother
        // with the projection.
        const dLat = 0.3;
        const dLon = 0.5;
        const bbox: [[number, number], [number, number]] = [
          [center[0] - dLat, center[1] - dLon],
          [center[0] + dLat, center[1] + dLon],
        ];
        suggest.options.set('boundedBy', bbox);
        suggest.options.set('strictBounds', false);
      })
      .catch(() => {
        /* leave previous bounds in place */
      });
    return () => {
      cancelled = true;
    };
  }, [city, ready]);

  // Keep the marker / map in sync with props (e.g. on initial load of an
  // existing event into the form).
  useEffect(() => {
    if (!ready) return;
    const map = mapRef.current;
    const ymaps = ymapsRef.current;
    if (!map || !ymaps) return;
    if (latitude == null || longitude == null) {
      if (placemarkRef.current) {
        map.geoObjects.remove(placemarkRef.current);
        placemarkRef.current = null;
      }
      return;
    }
    const coords: [number, number] = [latitude, longitude];
    if (!placemarkRef.current) {
      const placemark = new ymaps.Placemark(
        coords,
        {},
        { preset: 'islands#redIcon', draggable: true },
      );
      placemark.events.add('dragend', () => {
        const c = placemark.geometry.getCoordinates() as [number, number];
        reverseGeocode(ymaps, c);
      });
      map.geoObjects.add(placemark);
      placemarkRef.current = placemark;
    } else {
      placemarkRef.current.geometry.setCoordinates(coords);
    }
    map.setCenter(coords, PICKED_ZOOM, { duration: 200 });
  }, [latitude, longitude, ready, reverseGeocode]);

  return (
    <div className="address-picker">
      <input
        id={inputId}
        ref={inputRef}
        className="input"
        type="text"
        autoComplete="off"
        placeholder={placeholder}
        value={address}
        onChange={(e) => onAddressChange(e.target.value)}
      />
      {error ? (
        <div className="map" aria-label="карта">
          <span className="map-error">{error}</span>
        </div>
      ) : (
        <div
          className="map map-live address-picker-map"
          ref={mapContainerRef}
          aria-label="карта"
        />
      )}
      <p className="muted address-picker-hint">
        выберите подсказку или кликните на карте, чтобы поставить точку
      </p>
      {diagnostic && (
        <p className="address-picker-diagnostic">{diagnostic}</p>
      )}
    </div>
  );
}
