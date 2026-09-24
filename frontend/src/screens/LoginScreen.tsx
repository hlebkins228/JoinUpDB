import { useState, type FormEvent } from 'react';
import { Field } from '../components/Field';
import { TextInput } from '../components/inputs';
import { login as loginApi } from '../api/auth';
import { ApiError } from '../api/client';
import { useAuth } from '../auth/AuthContext';

interface LoginScreenProps {
  onGoRegister(): void;
}

export function LoginScreen({ onGoRegister }: LoginScreenProps) {
  const { signIn } = useAuth();
  const [loginValue, setLoginValue] = useState('');
  const [password, setPassword] = useState('');
  const [errors, setErrors] = useState<{ login?: string; password?: string }>({});
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    const next: { login?: string; password?: string } = {};
    if (!loginValue.trim()) next.login = 'введите логин';
    if (!password) next.password = 'введите пароль';
    setErrors(next);
    if (Object.keys(next).length) return;

    setBusy(true);
    setSubmitError(null);
    try {
      const res = await loginApi(loginValue.trim(), password);
      if (!res?.token) {
        setSubmitError('сервер не вернул токен');
        return;
      }
      await signIn(res.token);
    } catch (err) {
      if (err instanceof ApiError) {
        setSubmitError(err.status === 401 ? 'неверный логин или пароль' : err.message);
      } else {
        setSubmitError(err instanceof Error ? err.message : String(err));
      }
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="auth-shell">
      <div className="auth-card">
        <div className="auth-brand">
          <span className="auth-brand-dot" />
          <span>joinup</span>
        </div>
        <h1 className="t-display">войти</h1>
        <p className="auth-sub">встречайтесь оффлайн</p>

        {submitError && <div className="alert">{submitError}</div>}

        <form onSubmit={onSubmit} noValidate>
          <Field label="логин" error={errors.login}>
            <TextInput
              value={loginValue}
              onChange={setLoginValue}
              placeholder="anna_k"
              autoComplete="username"
            />
          </Field>
          <Field label="пароль" error={errors.password}>
            <TextInput
              type="password"
              value={password}
              onChange={setPassword}
              placeholder="••••••"
              autoComplete="current-password"
            />
          </Field>

          <button type="submit" className="btn btn-block" disabled={busy}>
            {busy ? 'входим…' : 'войти'}
          </button>
        </form>

        <div className="auth-foot">
          ещё нет аккаунта?{' '}
          <button type="button" className="auth-link" onClick={onGoRegister}>
            создать
          </button>
        </div>
      </div>
    </div>
  );
}
