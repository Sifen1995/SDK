import { useEffect, useState } from 'react';
import { Layers } from 'lucide-react';
import { api } from '../lib/api';
import type { AudienceSegment } from '../types';

export default function AdminSegments() {
  const [segments, setSegments] = useState<AudienceSegment[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    api.listSegments()
      .then(setSegments)
      .catch((err: Error) => setError(err.message))
      .finally(() => setLoading(false));
  }, []);

  return (
    <div>
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-xl font-bold text-primary mb-1">Audience Segments</h1>
          <p className="text-sm text-muted">Manage the purchasable audience segments available to advertisers.</p>
        </div>
        <button className="btn-primary opacity-50 cursor-not-allowed" title="Coming soon">
          + New Segment
        </button>
      </div>

      {error && <div className="alert-error mb-6">{error}</div>}

      {loading ? (
        <div className="card p-6">
          <div className="animate-pulse flex items-center gap-3">
            <div className="h-10 w-10 rounded-full bg-[var(--bg-subtle)]" />
            <div className="flex-1 space-y-2">
              <div className="h-4 w-32 rounded bg-[var(--bg-subtle)]" />
              <div className="h-3 w-48 rounded bg-[var(--bg-subtle)]" />
            </div>
          </div>
        </div>
      ) : segments.length === 0 ? (
        <div className="card p-12 text-center border-dashed">
          <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-brand-50 text-brand-600 dark:bg-brand-900/20 dark:text-brand-400">
            <Layers size={20} />
          </div>
          <h3 className="mt-4 text-sm font-semibold text-primary">No audience segments available yet</h3>
          <p className="mx-auto mt-2 max-w-md text-sm text-muted">
            Segments will appear here as soon as they are created or approved for use.
          </p>
        </div>
      ) : (
        <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-4">
          {segments.map(seg => (
            <div key={seg.id} className="card p-5 bg-white dark:bg-[#1c1b22] border-[var(--border)] hover:border-brand-300 transition">
              <h3 className="font-bold text-lg text-primary">{seg.name}</h3>
              <p className="text-xs font-mono text-muted mt-1 break-all">{seg.id}</p>
              
              <div className="mt-4 pt-4 border-t border-[var(--border)] flex justify-between items-center">
                <span className="text-sm font-medium text-muted">Price (ETB)</span>
                <span className="text-lg font-bold text-brand-600 dark:text-brand-400">{seg.price_etb}</span>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
