import { useState, type FormEvent } from 'react';
import {
  Card, CardContent, CardHeader, CardTitle, Button, Input, Label,
  StatusPill, LoadingState, ErrorState, EmptyState, InlineError, ZonePicker,
} from '@skykin/ui';
import { MapPin } from 'lucide-react';
import { useAuth } from '../context/AuthContext';
import { useZones, useCreateZone } from '../lib/queries';
import { formatDate } from '../lib/campaignUtils';

// Addis Ababa — a sane starting view rather than the middle of the ocean.
const DEFAULT_CENTRE = { latitude: 9.022736, longitude: 38.746799 };
const DEFAULT_RADIUS = 150;

// Server-side bounds, mirrored so the error appears before the request.
const MIN_RADIUS = 1;
const MAX_RADIUS = 50000;

export default function Zones() {
  const { canWrite } = useAuth();
  const zonesQ = useZones();
  const createZone = useCreateZone();

  const [latitude, setLatitude] = useState(String(DEFAULT_CENTRE.latitude));
  const [longitude, setLongitude] = useState(String(DEFAULT_CENTRE.longitude));
  const [radius, setRadius] = useState(String(DEFAULT_RADIUS));
  const [formError, setFormError] = useState('');

  // The map needs numbers even while the inputs hold half-typed text.
  const lat = Number(latitude);
  const lng = Number(longitude);
  const rad = Number(radius);
  const mapLat = Number.isFinite(lat) ? lat : DEFAULT_CENTRE.latitude;
  const mapLng = Number.isFinite(lng) ? lng : DEFAULT_CENTRE.longitude;
  const mapRadius = Number.isFinite(rad) && rad > 0 ? rad : DEFAULT_RADIUS;

  function handleSubmit(e: FormEvent) {
    e.preventDefault();
    if (!Number.isFinite(lat) || lat < -90 || lat > 90) return setFormError('Latitude must be between -90 and 90');
    if (!Number.isFinite(lng) || lng < -180 || lng > 180) return setFormError('Longitude must be between -180 and 180');
    if (!Number.isFinite(rad) || rad < MIN_RADIUS || rad > MAX_RADIUS) {
      return setFormError(`Radius must be between ${MIN_RADIUS} and ${MAX_RADIUS.toLocaleString()} metres`);
    }
    // The backend binds latitude/longitude with `required`, which rejects an
    // exact 0 — worth saying so rather than surfacing a confusing 400.
    if (lat === 0 || lng === 0) return setFormError('Latitude and longitude cannot be exactly 0');

    setFormError('');
    createZone.mutate({ latitude: lat, longitude: lng, radius_metres: rad });
  }

  if (zonesQ.isPending) return <LoadingState label="Loading store zones…" />;
  if (zonesQ.isError) {
    return (
      <ErrorState
        title="Could not load zones"
        message={(zonesQ.error as Error)?.message ?? 'Request failed'}
        onRetry={() => zonesQ.refetch()}
      />
    );
  }

  const zones = zonesQ.data ?? [];
  const error = formError || (createZone.isError ? (createZone.error as Error).message : '');

  return (
    <div className="space-y-6">
      <div>
        <h2 className="font-display text-xl font-bold">Store zones</h2>
        <p className="mt-1 text-sm text-muted-foreground">
          Draw a circle around a store. Link zones to a campaign and customers who walk
          into the circle receive the offer. New zones stay inactive until an operator
          approves the campaign they're linked to.
        </p>
      </div>

      {!canWrite && (
        <div className="rounded-lg border border-border bg-muted/50 px-4 py-3 text-sm text-muted-foreground">
          Your role is read-only, so you can view zones but not create them.
        </div>
      )}

      <div className="grid gap-6 lg:grid-cols-2">
        {canWrite && (
          <Card>
            <CardHeader><CardTitle>New zone</CardTitle></CardHeader>
            <CardContent>
              <form onSubmit={handleSubmit} className="space-y-4">
                {error && <InlineError message={error} />}

                <ZonePicker
                  latitude={mapLat}
                  longitude={mapLng}
                  radiusMetres={mapRadius}
                  onChange={({ latitude: nextLat, longitude: nextLng }) => {
                    setLatitude(nextLat.toFixed(6));
                    setLongitude(nextLng.toFixed(6));
                  }}
                />
                <p className="text-xs text-muted-foreground">
                  Click the map to place the store, or paste coordinates below.
                </p>

                <div className="grid gap-4 sm:grid-cols-2">
                  <div className="space-y-1.5">
                    <Label htmlFor="lat">Latitude</Label>
                    <Input id="lat" inputMode="decimal" value={latitude} onChange={e => setLatitude(e.target.value)} />
                  </div>
                  <div className="space-y-1.5">
                    <Label htmlFor="lng">Longitude</Label>
                    <Input id="lng" inputMode="decimal" value={longitude} onChange={e => setLongitude(e.target.value)} />
                  </div>
                </div>
                <div className="space-y-1.5">
                  <Label htmlFor="radius">Radius (metres)</Label>
                  <Input id="radius" type="number" min={MIN_RADIUS} max={MAX_RADIUS} value={radius} onChange={e => setRadius(e.target.value)} />
                </div>

                <Button type="submit" disabled={createZone.isPending} className="w-full">
                  {createZone.isPending ? 'Creating…' : 'Create zone'}
                </Button>
              </form>
            </CardContent>
          </Card>
        )}

        <Card>
          <CardHeader><CardTitle>Your zones ({zones.length})</CardTitle></CardHeader>
          <CardContent>
            {zones.length === 0 ? (
              <EmptyState
                icon={MapPin}
                title="No zones yet"
                description="Create a zone for each store location you want to advertise around."
              />
            ) : (
              <ul className="space-y-3">
                {zones.map(zone => (
                  <li key={zone.id} className="rounded-lg border border-border p-3">
                    <div className="flex items-start justify-between gap-3">
                      <div className="min-w-0">
                        <p className="font-mono text-sm tabular-nums">
                          {zone.latitude.toFixed(5)}, {zone.longitude.toFixed(5)}
                        </p>
                        <p className="mt-1 text-xs text-muted-foreground">
                          {zone.radius_metres.toLocaleString()} m radius
                          {zone.created_at && ` · created ${formatDate(zone.created_at)}`}
                        </p>
                      </div>
                      <StatusPill status={zone.is_active ? 'active' : 'pending'} />
                    </div>
                    {!zone.is_active && (
                      <p className="mt-2 text-xs text-muted-foreground">
                        Goes live when an operator approves a campaign this zone is linked to.
                      </p>
                    )}
                  </li>
                ))}
              </ul>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
