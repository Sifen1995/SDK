import type { ReactNode } from 'react';

interface FilterBarProps {
  children: ReactNode;
  resultCount?: number;
  totalCount?: number;
}

export default function FilterBar({ children, resultCount, totalCount }: FilterBarProps) {
  return (
    <div className="filter-bar mb-6">
      <div className="flex flex-wrap items-center gap-3">{children}</div>
      {resultCount !== undefined && totalCount !== undefined && (
        <p className="text-xs text-muted mt-3">
          Showing {resultCount} of {totalCount}
        </p>
      )}
    </div>
  );
}

interface FilterInputProps {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  className?: string;
}

export function FilterSearch({ value, onChange, placeholder = 'Search…', className = '' }: FilterInputProps) {
  return (
    <input
      type="search"
      value={value}
      onChange={e => onChange(e.target.value)}
      placeholder={placeholder}
      className={`filter-input ${className}`}
    />
  );
}

interface FilterSelectProps {
  value: string;
  onChange: (value: string) => void;
  options: { value: string; label: string }[];
  className?: string;
}

export function FilterSelect({ value, onChange, options, className = '' }: FilterSelectProps) {
  return (
    <select value={value} onChange={e => onChange(e.target.value)} className={`filter-select ${className}`}>
      {options.map(opt => (
        <option key={opt.value} value={opt.value}>
          {opt.label}
        </option>
      ))}
    </select>
  );
}
