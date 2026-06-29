const PAGE_SIZE = 5;

export function paginateSlice<T>(items: T[], page: number, pageSize = PAGE_SIZE): T[] {
  const offset = (page - 1) * pageSize;
  return items.slice(offset, offset + pageSize);
}

export function totalPages(count: number, pageSize = PAGE_SIZE): number {
  return Math.max(1, Math.ceil(count / pageSize));
}

interface OffsetPaginationProps {
  page: number;
  totalItems: number;
  pageSize?: number;
  onPageChange: (page: number) => void;
}

export default function OffsetPagination({
  page,
  totalItems,
  pageSize = PAGE_SIZE,
  onPageChange,
}: OffsetPaginationProps) {
  const pages = totalPages(totalItems, pageSize);
  if (totalItems <= pageSize) return null;

  const start = (page - 1) * pageSize + 1;
  const end = Math.min(page * pageSize, totalItems);

  return (
    <div className="pagination-bar">
      <p className="text-sm text-muted">
        Showing <span className="font-medium text-primary">{start}</span>–
        <span className="font-medium text-primary">{end}</span> of{' '}
        <span className="font-medium text-primary">{totalItems}</span>
      </p>
      <div className="flex items-center gap-2">
        <button
          type="button"
          className="btn-secondary px-4 py-2 text-sm"
          disabled={page <= 1}
          onClick={() => onPageChange(page - 1)}
        >
          ← Previous
        </button>
        <span className="text-sm text-muted px-2">
          Page {page} of {pages}
        </span>
        <button
          type="button"
          className="btn-secondary px-4 py-2 text-sm"
          disabled={page >= pages}
          onClick={() => onPageChange(page + 1)}
        >
          Next →
        </button>
      </div>
    </div>
  );
}

export { PAGE_SIZE };
