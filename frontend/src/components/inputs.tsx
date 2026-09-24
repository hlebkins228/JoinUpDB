import type {
  ChangeEvent,
  InputHTMLAttributes,
  SelectHTMLAttributes,
  TextareaHTMLAttributes,
} from 'react';

type StringInputProps = Omit<InputHTMLAttributes<HTMLInputElement>, 'value' | 'onChange'> & {
  value: string;
  onChange(value: string): void;
};

export function TextInput({ value, onChange, className = '', ...rest }: StringInputProps) {
  return (
    <input
      className={`input ${className}`.trim()}
      value={value}
      onChange={(e: ChangeEvent<HTMLInputElement>) => onChange(e.target.value)}
      {...rest}
    />
  );
}

type NumberInputProps = Omit<InputHTMLAttributes<HTMLInputElement>, 'value' | 'onChange' | 'type'> & {
  value: number | '';
  onChange(value: number | ''): void;
};

export function NumberInput({ value, onChange, className = '', ...rest }: NumberInputProps) {
  return (
    <input
      type="number"
      className={`input ${className}`.trim()}
      value={value === '' ? '' : value}
      onChange={(e: ChangeEvent<HTMLInputElement>) => {
        const raw = e.target.value;
        if (raw === '') {
          onChange('');
          return;
        }
        const parsed = Number(raw);
        onChange(Number.isFinite(parsed) ? parsed : '');
      }}
      {...rest}
    />
  );
}

type TextAreaProps = Omit<
  TextareaHTMLAttributes<HTMLTextAreaElement>,
  'value' | 'onChange'
> & {
  value: string;
  onChange(value: string): void;
};

export function TextArea({ value, onChange, className = '', ...rest }: TextAreaProps) {
  return (
    <textarea
      className={`textarea ${className}`.trim()}
      value={value}
      onChange={(e: ChangeEvent<HTMLTextAreaElement>) => onChange(e.target.value)}
      {...rest}
    />
  );
}

type SelectProps = Omit<SelectHTMLAttributes<HTMLSelectElement>, 'value' | 'onChange'> & {
  value: string;
  onChange(value: string): void;
  options: Array<{ value: string; label: string }>;
  placeholder?: string;
};

export function Select({
  value,
  onChange,
  options,
  placeholder,
  className = '',
  ...rest
}: SelectProps) {
  return (
    <select
      className={`select ${className}`.trim()}
      value={value}
      onChange={(e: ChangeEvent<HTMLSelectElement>) => onChange(e.target.value)}
      {...rest}
    >
      {placeholder !== undefined && (
        <option value="" disabled={!value}>
          {placeholder}
        </option>
      )}
      {options.map((opt) => (
        <option key={opt.value} value={opt.value}>
          {opt.label}
        </option>
      ))}
    </select>
  );
}
