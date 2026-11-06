import { useState, type FormEvent } from 'react';
import { Field } from '../components/Field';
import { NumberInput, Select, TextInput } from '../components/inputs';
import { CITIES } from '../lib/constants';
import { createUser } from '../api/users';
import { login as loginApi } from '../api/auth';
import { ApiError } from '../api/client';
import { useAuth } from '../auth/AuthContext';

interface RegisterScreenProps {
  onGoLogin(): void;
}

interface FormErrors {
  name?: string;
  login?: string;
  password?: string;
  password2?: string;
  age?: string;
  city?: string;
}

export function RegisterScreen({ onGoLogin }: RegisterScreenProps) {
  const { signIn } = useAuth();
  const [name, setName] = useState('');
  const [loginValue, setLoginValue] = useState('');
  const [password, setPassword] = useState('');
  const [password2, setPassword2] = useState('');
  const [age, setAge] = useState<number | ''>('');
  const [city, setCity] = useState('');
  const [tg, setTg] = useState('');
  const [errors, setErrors] = useState<FormErrors>({});
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  function validate(): FormErrors {
    const next: FormErrors = {};
    if (!name.trim()) next.name = 'введите имя';
    if (!loginValue.trim()) next.login = 'введите логин';
    if (!password) next.password = 'введите пароль';
    else if (password.length < 6) next.password = 'минимум 6 символов';
    if (password !== password2) next.password2 = 'пароли не совпадают';
    if (age === '' || Number(age) <= 0) next.age = 'укажите возраст';
    if (!city) next.city = 'выберите город';
    return next;
  }

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    const next = validate();
    setErrors(next);
    if (Object.keys(next).length) return;

    setBusy(true);
    setSubmitError(null);
    try {
      await createUser({
        name: name.trim(),
        age: Number(age),
        login: loginValue.trim(),
        password,
        city,
        telegram_login: tg.trim() || undefined,
      });
      const auth = await loginApi(loginValue.trim(), password);
      if (!auth?.token) {
        setSubmitError('регистрация прошла, но не удалось войти автоматически');
        return;
      }
      await signIn(auth.token);
    } catch (err) {
      if (err instanceof ApiError) {
        setSubmitError(err.message);
      } else {
        setSubmitError(err instanceof Error ? err.message : String(err));
      }
    } finally {
      setBusy(false);
    }
  }

  const cityOptions = CITIES.map((c) => ({ value: c, label: c }));

  return (
    <div className="auth-shell">
      <div className="auth-card">
        <div className="auth-brand">
          <span className="auth-brand-dot" />
          <span>joinup</span>
        </div>
        <h1 className="t-display">создать аккаунт</h1>
        <p className="auth-sub">найдите свою компанию</p>

        {submitError && <div className="alert">{submitError}</div>}

        <form onSubmit={onSubmit} noValidate>
          <Field label="имя" error={errors.name}>
            <TextInput value={name} onChange={setName} placeholder="Анна К." />
          </Field>
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
              autoComplete="new-password"
            />
          </Field>
          <Field label="пароль ещё раз" error={errors.password2}>
            <TextInput
              type="password"
              value={password2}
              onChange={setPassword2}
              autoComplete="new-password"
            />
          </Field>
          <Field label="возраст" error={errors.age}>
            <NumberInput value={age} onChange={setAge} min={1} max={120} />
          </Field>
          <Field label="город" error={errors.city}>
            <Select
              value={city}
              onChange={setCity}
              options={cityOptions}
              placeholder="выберите город"
            />
          </Field>
          <Field label="telegram" optional hint="без @">
            <TextInput value={tg} onChange={setTg} placeholder="anna_k" />
          </Field>

          <button type="submit" className="btn btn-block" disabled={busy}>
            {busy ? 'создаём…' : 'создать аккаунт'}
          </button>
        </form>

        <div className="auth-foot">
          уже есть аккаунт?{' '}
          <button type="button" className="auth-link" onClick={onGoLogin}>
            войти
          </button>
        </div>
      </div>
    </div>
  );
}
