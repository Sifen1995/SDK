import { useEffect, useMemo } from 'react';
import { MapContainer, TileLayer, Circle, Marker, useMap, useMapEvents } from 'react-leaflet';
import L from 'leaflet';
import 'leaflet/dist/leaflet.css';
import { cn } from '../lib/utils';

/**
 * Leaflet's default marker icon is referenced by a relative image path that
 * bundlers rewrite, leaving a broken image. Resolve the packaged assets through
 * the bundler explicitly instead.
 */
const markerIcon = L.icon({
  iconUrl: new URL('leaflet/dist/images/marker-icon.png', import.meta.url).href,
  iconRetinaUrl: new URL('leaflet/dist/images/marker-icon-2x.png', import.meta.url).href,
  shadowUrl: new URL('leaflet/dist/images/marker-shadow.png', import.meta.url).href,
  iconSize: [25, 41],
  iconAnchor: [12, 41],
  shadowSize: [41, 41],
});

export interface ZonePickerProps {
  latitude: number;
  longitude: number;
  radiusMetres: number;
  /** Omit to render read-only — clicks no longer move the marker. */
  onChange?: (next: { latitude: number; longitude: number }) => void;
  className?: string;
  heightClassName?: string;
}

/** Click-to-place. No-op when the picker is read-only. */
function ClickHandler({ onChange }: { onChange?: ZonePickerProps['onChange'] }) {
  useMapEvents({
    click(e) {
      onChange?.({ latitude: e.latlng.lat, longitude: e.latlng.lng });
    },
  });
  return null;
}

/**
 * Keeps the view on the marker when the coordinate changes from outside — e.g.
 * the advertiser pastes coordinates into the numeric inputs instead of clicking.
 */
function Recenter({ latitude, longitude }: { latitude: number; longitude: number }) {
  const map = useMap();
  useEffect(() => {
    map.setView([latitude, longitude], map.getZoom(), { animate: true });
  }, [map, latitude, longitude]);
  return null;
}

/**
 * Map picker for a circular geofence zone.
 *
 * Two-way bound: click to place the marker, or type coordinates into the numeric
 * inputs alongside it and the map follows. The numeric inputs remain the source
 * of truth, so a zone can still be created by pasting coordinates from any map
 * service.
 *
 * Tiles are fetched from OpenStreetMap at runtime — the only external network
 * dependency in the portals.
 */
export function ZonePicker({
  latitude,
  longitude,
  radiusMetres,
  onChange,
  className,
  heightClassName = 'h-72',
}: ZonePickerProps) {
  const center = useMemo<[number, number]>(() => [latitude, longitude], [latitude, longitude]);
  const readOnly = !onChange;

  return (
    <div
      className={cn(
        'overflow-hidden rounded-lg border border-border',
        heightClassName,
        className,
      )}
    >
      <MapContainer
        center={center}
        zoom={15}
        scrollWheelZoom={!readOnly}
        dragging={!readOnly}
        className="size-full"
        // Leaflet's own panes sit at z-index 400+, which would cover dialogs.
        style={{ zIndex: 0 }}
      >
        <TileLayer
          attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
          url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
        />
        <Marker position={center} icon={markerIcon} />
        <Circle
          center={center}
          radius={Math.max(radiusMetres, 1)}
          pathOptions={{ color: '#0083d5', fillColor: '#0083d5', fillOpacity: 0.15 }}
        />
        <Recenter latitude={latitude} longitude={longitude} />
        {!readOnly && <ClickHandler onChange={onChange} />}
      </MapContainer>
    </div>
  );
}
