import { useState } from 'react';
import { AuthProvider, useAuth } from './auth/AuthContext';
import { Header } from './components/Header';
import { LoginScreen } from './screens/LoginScreen';
import { RegisterScreen } from './screens/RegisterScreen';
import { SearchScreen } from './screens/SearchScreen';
import { MyEventsScreen } from './screens/MyEventsScreen';
import { JoiningScreen } from './screens/JoiningScreen';
import { ProfileScreen } from './screens/ProfileScreen';
import { EventDetailScreen } from './screens/EventDetailScreen';
import type { Route, RouteState } from './routes';

function Shell() {
  const { token, loading, user } = useAuth();
  const [authRoute, setAuthRoute] = useState<'login' | 'register'>('login');
  const [state, setState] = useState<RouteState>({ route: 'search', refreshKey: 0 });

  if (loading) {
    return <div className="empty-state">загружаем…</div>;
  }

  if (!token || !user) {
    if (authRoute === 'register') {
      return <RegisterScreen onGoLogin={() => setAuthRoute('login')} />;
    }
    return <LoginScreen onGoRegister={() => setAuthRoute('register')} />;
  }

  function navigate(route: Route, eventId: number | null = null, userId: number | null = null) {
    setState((prev) => ({
      route,
      eventId,
      userId,
      refreshKey: (prev.refreshKey ?? 0) + 1,
    }));
  }

  let body: JSX.Element;
  switch (state.route) {
    case 'search':
      body = (
        <SearchScreen
          onOpenEvent={(id) => navigate('event-detail', id)}
          onOpenUser={(id) => navigate('profile', null, id)}
          onCreateEvent={() => navigate('event-new', null)}
          refreshKey={state.refreshKey}
        />
      );
      break;
    case 'mine':
      body = (
        <MyEventsScreen
          onOpenEvent={(id) => navigate('event-detail', id)}
          onOpenUser={(id) => navigate('profile', null, id)}
          onCreateEvent={() => navigate('event-new', null)}
          refreshKey={state.refreshKey}
        />
      );
      break;
    case 'joining':
      body = (
        <JoiningScreen
          onOpenEvent={(id) => navigate('event-detail', id)}
          onOpenUser={(id) => navigate('profile', null, id)}
          refreshKey={state.refreshKey}
        />
      );
      break;
    case 'profile':
      body = (
        <ProfileScreen
          userId={state.userId ?? null}
          onBack={() => navigate('search')}
          refreshKey={state.refreshKey}
        />
      );
      break;
    case 'event-detail':
      body = (
        <EventDetailScreen
          eventId={state.eventId ?? null}
          onOpenUser={(id) => navigate('profile', null, id)}
          onBack={() => navigate('search')}
          onAfterDelete={() => navigate('mine')}
          refreshKey={state.refreshKey}
        />
      );
      break;
    case 'event-new':
      body = (
        <EventDetailScreen
          eventId={null}
          onOpenUser={(id) => navigate('profile', null, id)}
          onBack={() => navigate('search')}
          onAfterDelete={() => navigate('mine')}
          refreshKey={state.refreshKey}
        />
      );
      break;
    default:
      body = (
        <SearchScreen
          onOpenEvent={(id) => navigate('event-detail', id)}
          onOpenUser={(id) => navigate('profile', null, id)}
          onCreateEvent={() => navigate('event-new', null)}
          refreshKey={state.refreshKey}
        />
      );
  }

  return (
    <div className="app-shell">
      <Header
        route={state.route}
        user={user}
        onNavigate={(r) => navigate(r)}
      />
      <main style={{ flex: 1, paddingBottom: 64 }}>{body}</main>
    </div>
  );
}

export default function App() {
  return (
    <AuthProvider>
      <Shell />
    </AuthProvider>
  );
}
