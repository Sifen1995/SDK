import { useEffect, useState, type ReactNode } from 'react';

interface ModalProps {
  open: boolean;
  onClose: () => void;
  title: string;
  description?: string;
  children?: ReactNode;
  footer?: ReactNode;
  size?: 'sm' | 'md' | 'lg';
}

export default function Modal({
  open,
  onClose,
  title,
  description,
  children,
  footer,
  size = 'md',
}: ModalProps) {
  useEffect(() => {
    if (!open) return;
    function onKey(e: KeyboardEvent) {
      if (e.key === 'Escape') onClose();
    }
    document.addEventListener('keydown', onKey);
    document.body.style.overflow = 'hidden';
    return () => {
      document.removeEventListener('keydown', onKey);
      document.body.style.overflow = '';
    };
  }, [open, onClose]);

  if (!open) return null;

  const widthClass = size === 'sm' ? 'max-w-sm' : size === 'lg' ? 'max-w-2xl' : 'max-w-md';

  return (
    <div className="modal-backdrop" onClick={onClose} role="presentation">
      <div
        className={`modal-card ${widthClass}`}
        onClick={e => e.stopPropagation()}
        role="dialog"
        aria-modal="true"
        aria-labelledby="modal-title"
      >
        <div className="modal-header">
          <h3 id="modal-title" className="text-lg font-semibold text-primary">{title}</h3>
          {description && <p className="text-sm text-muted mt-1">{description}</p>}
        </div>
        {children && <div className="modal-body">{children}</div>}
        {footer && <div className="modal-footer">{footer}</div>}
      </div>
    </div>
  );
}

interface ConfirmModalProps {
  open: boolean;
  title: string;
  description: string;
  confirmLabel?: string;
  cancelLabel?: string;
  variant?: 'primary' | 'danger' | 'success';
  loading?: boolean;
  onConfirm: () => void;
  onCancel: () => void;
}

export function ConfirmModal({
  open,
  title,
  description,
  confirmLabel = 'Confirm',
  cancelLabel = 'Cancel',
  variant = 'primary',
  loading = false,
  onConfirm,
  onCancel,
}: ConfirmModalProps) {
  const confirmClass =
    variant === 'danger' ? 'btn-danger' : variant === 'success' ? 'btn-success' : 'btn-primary';

  return (
    <Modal
      open={open}
      onClose={onCancel}
      title={title}
      description={description}
      footer={
        <>
          <button type="button" className="btn-secondary" onClick={onCancel} disabled={loading}>
            {cancelLabel}
          </button>
          <button type="button" className={confirmClass} onClick={onConfirm} disabled={loading}>
            {loading ? 'Please wait…' : confirmLabel}
          </button>
        </>
      }
    />
  );
}

interface NotesModalProps {
  open: boolean;
  title: string;
  description: string;
  label?: string;
  placeholder?: string;
  required?: boolean;
  confirmLabel?: string;
  variant?: 'primary' | 'danger';
  loading?: boolean;
  onConfirm: (notes: string) => void;
  onCancel: () => void;
}

export function NotesModal({
  open,
  title,
  description,
  label = 'Notes',
  placeholder = 'Add optional notes…',
  required = false,
  confirmLabel = 'Submit',
  variant = 'primary',
  loading = false,
  onConfirm,
  onCancel,
}: NotesModalProps) {
  const [notes, setNotes] = useState('');

  useEffect(() => {
    if (open) setNotes('');
  }, [open]);

  const confirmClass = variant === 'danger' ? 'btn-danger' : 'btn-primary';

  return (
    <Modal
      open={open}
      onClose={onCancel}
      title={title}
      description={description}
      footer={
        <>
          <button type="button" className="btn-secondary" onClick={onCancel} disabled={loading}>
            Cancel
          </button>
          <button
            type="button"
            className={confirmClass}
            disabled={loading || (required && !notes.trim())}
            onClick={() => onConfirm(notes.trim())}
          >
            {loading ? 'Please wait…' : confirmLabel}
          </button>
        </>
      }
    >
      <label className="block text-sm font-medium text-primary mb-1.5">{label}</label>
      <textarea
        value={notes}
        onChange={e => setNotes(e.target.value)}
        rows={3}
        className="field-input resize-y"
        placeholder={placeholder}
      />
    </Modal>
  );
}
