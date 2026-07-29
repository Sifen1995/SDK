import { useQueryStates, parseAsInteger, parseAsString } from 'nuqs';

/**
 * URL-persisted list state (search + pagination) shared by list/table pages,
 * so filters and page survive refresh and are shareable. Sorting can be added
 * per-page with the same pattern.
 */
export function useListUrlState(defaults?: { pageSize?: number }) {
  const [state, setState] = useQueryStates({
    q: parseAsString.withDefault(''),
    page: parseAsInteger.withDefault(1),
    status: parseAsString.withDefault('all'),
  });

  const pageSize = defaults?.pageSize ?? 10;
  return {
    q: state.q,
    page: state.page,
    status: state.status,
    pageSize,
    setQ: (q: string) => setState({ q, page: 1 }),
    setPage: (page: number) => setState({ page }),
    setStatus: (status: string) => setState({ status, page: 1 }),
  };
}

/** Download an array of flat records as a CSV file. */
export function exportToCsv<T extends Record<string, unknown>>(
  rows: T[],
  filename: string,
  columns?: { key: keyof T; label: string }[],
) {
  if (!rows.length) return;
  const cols = columns ?? (Object.keys(rows[0]) as (keyof T)[]).map(key => ({ key, label: String(key) }));
  const esc = (v: unknown) => {
    const s = v == null ? '' : String(v);
    return /[",\n]/.test(s) ? `"${s.replace(/"/g, '""')}"` : s;
  };
  const header = cols.map(c => esc(c.label)).join(',');
  const body = rows.map(r => cols.map(c => esc(r[c.key])).join(',')).join('\n');
  const blob = new Blob([`${header}\n${body}`], { type: 'text/csv;charset=utf-8;' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = filename.endsWith('.csv') ? filename : `${filename}.csv`;
  a.click();
  URL.revokeObjectURL(url);
}
