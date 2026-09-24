import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
  type ReactNode,
} from 'react';
import type { UserResponse } from '../types';
import { getToken, setToken } from '../api/client';
import { getUser } from '../api/users';
import { decodeJwt, isExpired } from './jwt';

interface AuthContextValue {
  token: string | null;
  user: UserResponse | null;
  loading: boolean;
  signIn(token: string): Promise<void>;
  signOut(): void;
  refreshUser(): Promise<void>;
}

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [token, setTokenState] = useState<string | null>(() => {
    const stored = getToken();
    if (!stored) return null;
    if (isExpired(decodeJwt(stored))) {
      setToken(null);
      return null;
    }
    return stored;
  });
  const [user, setUser] = useState<UserResponse | null>(null);
  const [loading, setLoading] = useState<boolean>(!!token);
  // The bootstrap effect should only fetch the user once for the token that
  // was restored from localStorage. Subsequent transitions (signIn / signOut)
  // manage `user` directly, so we skip the effect after the first run to
  // avoid double-fetching when signIn calls setTokenState.
  const bootstrapped = useRef(false);

  // Returns `true` when the user was loaded for the given token. On any
  // failure (expired token, rejected request) it cleans up persisted state
  // and returns `false`, leaving the caller to decide what to do.
  const loadUser = useCallback(async (rawToken: string): Promise<boolean> => {
    const payload = decodeJwt(rawToken);
    if (!payload || isExpired(payload)) {
      setToken(null);
      setTokenState(null);
      setUser(null);
      return false;
    }
    try {
      const fetched = await getUser(payload.UserID);
      setUser(fetched);
      return true;
    } catch (err) {
      console.error('failed to load user', err);
      setToken(null);
      setTokenState(null);
      setUser(null);
      return false;
    }
  }, []);

  useEffect(() => {
    if (bootstrapped.current) return;
    bootstrapped.current = true;
    if (!token) {
      setLoading(false);
      return;
    }
    setLoading(true);
    loadUser(token).finally(() => setLoading(false));
  }, [token, loadUser]);

  const signIn = useCallback(
    async (newToken: string) => {
      setToken(newToken);
      setLoading(true);
      try {
        const ok = await loadUser(newToken);
        // Only commit the token to React state when the user was actually
        // loaded — otherwise loadUser has already cleared persistent state
        // and we must not resurrect a failed token.
        if (ok) setTokenState(newToken);
      } finally {
        setLoading(false);
      }
    },
    [loadUser],
  );

  const signOut = useCallback(() => {
    setToken(null);
    setTokenState(null);
    setUser(null);
  }, []);

  const refreshUser = useCallback(async () => {
    if (!token) return;
    await loadUser(token);
  }, [token, loadUser]);

  const value = useMemo<AuthContextValue>(
    () => ({ token, user, loading, signIn, signOut, refreshUser }),
    [token, user, loading, signIn, signOut, refreshUser],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

// eslint-disable-next-line react-refresh/only-export-components
export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used within AuthProvider');
  return ctx;
}
