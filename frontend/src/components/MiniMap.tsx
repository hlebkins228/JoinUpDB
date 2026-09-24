import { useEffect, useRef, useState } from 'react';
import {
  YandexMapsConfigError,
  loadYandexMaps,
} from '../lib/yandexMaps';

interface MiniMapProps {
  longitude: number;
  latitude: number;
  zoom?: number;
}

// Read-only map preview used on the event detail screen. Rendered with the
// Yandex Maps JS API; falls back to a coordinates label if the API key is
// missing or the script fails to load.
export function MiniMap({ longitude, latitude, zoom = 15 }: MiniMapProps) {
  const containerRef = useRef<HTMLDivElement | null>(null);
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const mapRef = useRef<any>(null);
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const placemarkRef = useRef<any>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    loadYandexMaps()
      .then((ymaps) => {
        if (cancelled || !containerRef.current) return;
        const center: [number, number] = [latitude, longitude];
        const map = new ymaps.Map(
          containerRef.current,
          {
            center,
            zoom,
            controls: ['zoomControl'],
          },
          { suppressMapOpenBlock: true },
        );
        mapRef.current = map;
        const placemark = new ymaps.Placemark(
          center,
          {},
          { preset: 'islands#redIcon' },
        );
        map.geoObjects.add(placemark);
        placemarkRef.current = placemark;
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
      if (mapRef.current) {
        mapRef.current.destroy();
        mapRef.current = null;
        placemarkRef.current = null;
      }
    };
    // Intentionally only run once: subsequent coordinate changes are handled
    // by the effect below so the map instance is reused.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    const map = mapRef.current;
    const placemark = placemarkRef.current;
    if (!map || !placemark) return;
    const center: [number, number] = [latitude, longitude];
    placemark.geometry.setCoordinates(center);
    map.setCenter(center, zoom, { duration: 200 });
  }, [latitude, longitude, zoom]);

  if (error) {
    return (
      <div className="map" aria-label="карта">
        <span className="pin" />
        <span className="map-coords">
          {latitude.toFixed(4)}, {longitude.toFixed(4)}
        </span>
        <span className="map-error">{error}</span>
      </div>
    );
  }

  return <div className="map map-live" ref={containerRef} aria-label="карта" />;
}
