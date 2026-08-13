import { useState } from 'react';
import {
  Card, CardContent, Button, StatusPill, LoadingState, ErrorState, EmptyState,
  InlineError, ZonePicker,
} from '@skykin/ui';
import { MapPin, Check } from 'lucide-react';
import { usePendingZones, useActivateZone } from '../lib/queries';
import { formatDate } from '../lib/campaignUtils';

/**
 * Draft store zones awaiting activation.
 *
 * Most zones never need to be touched here — approving the campaign a zone is
 * linked to activates it automatically. This page exists for the leftovers:
 * zones linked to a campaign *after* it was approved, and zones created without
 * a campaign yet.
 */
export default function AdminPendingZones() {
  const zonesQ = usePendingZones();
  const activate = useActivateZone();
  const [expandedId, setExpandedId] = useState<string | null>(null);

  if (zonesQ.isPending) return <LoadingState label="Loading draft zones…" />;
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

  return (
    <div className="space-y-6">
      <div>
        <h2 className="font-display text-xl font-bold">Draft store zones</h2>
        <p className="mt-1 text-sm text-muted-foreground">
          Advertiser store zones that are not yet live. Approving a campaign activates
          the zones linked to it, so anything left here was created standalone or linked
          after its campaign was approved.
        </p>
      </div>

      {activate.isError && (
        <InlineError message={(activate.error as Error)?.message ?? 'Activation failed'} />
      )}

      {zones.length === 0 ? (
        <EmptyState
          icon={Check}
          title="Nothing waiting"
          description="Every store zone is active."
        />
      ) : (
        <div className="space-y-3">
          {zones.map(zone => {
            const expanded = expandedId === zone.id;
            return (
              <Card key={zone.id}>
                <CardContent className="p-4">
                  <div className="flex flex-wrap items-start justify-between gap-3">
                    <div className="min-w-0">
                      <div className="flex items-center gap-2">
                        <MapPin className="size-4 shrink-0 text-muted-foreground" />
                        <span className="font-mono text-sm tabular-nums">
                          {zone.latitude.toFixed(5)}, {zone.longitude.toFixed(5)}
                        </span>
                        <StatusPill status="pending" />
                      </div>
                      <p className="mt-1.5 text-xs text-muted-foreground">
                        {zone.radius_metres.toLocaleString()} m radius
                        {zone.advertiser_id && ` · advertiser ${zone.advertiser_id}`}
                        {zone.created_at && ` · created ${formatDate(zone.created_at)}`}
                      </p>
                    </div>
                    <div className="flex gap-2">
                      <Button
                        variant="secondary"
                        size="sm"
                        onClick={() => setExpandedId(expanded ? null : zone.id)}
                      >
                        {expanded ? 'Hide map' : 'View on map'}
                      </Button>
                      <Button
                        size="sm"
                        disabled={activate.isPending}
                        onClick={() => activate.mutate(zone.id)}
                      >
                        {activate.isPending ? 'Activating…' : 'Activate'}
                      </Button>
                    </div>
                  </div>

                  {expanded && (
                    <div className="mt-4">
                      <ZonePicker
                        latitude={zone.latitude}
                        longitude={zone.longitude}
                        radiusMetres={zone.radius_metres}
                        heightClassName="h-64"
                      />
                    </div>
                  )}
                </CardContent>
              </Card>
            );
          })}
        </div>
      )}
    </div>
  );
}
