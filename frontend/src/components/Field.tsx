import type { ReactNode } from 'react';

interface FieldProps {
  label: string;
  optional?: boolean;
  error?: string;
  hint?: string;
  children: ReactNode;
}

export function Field({ label, optional, error, hint, children }: FieldProps) {
  const cls = ['field', error ? 'field-error' : ''].filter(Boolean).join(' ');
  return (
    <label className={cls}>
      <span className="field-label">
        <span>{label}</span>
        {optional && <span className="opt">опционально</span>}
      </span>
      {children}
      {error && <span className="input-error">{error}</span>}
      {!error && hint && <span className="muted">{hint}</span>}
    </label>
  );
}
