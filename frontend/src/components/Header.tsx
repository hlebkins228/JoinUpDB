import { Avatar } from './Avatar';
import type { Route } from '../routes';
import type { UserResponse } from '../types';
import { imageUrl } from '../lib/imageUrl';

interface HeaderProps {
  route: Route;
  user: UserResponse | null;
  onNavigate(route: Route): void;
}

const TABS: Array<{ id: Route; label: string }> = [
  { id: 'search', label: 'поиск' },
  { id: 'mine', label: 'мои' },
  { id: 'joining', label: 'участвую' },
];

export function Header({ route, user, onNavigate }: HeaderProps) {
  return (
    <header className="app-header">
      <div className="container app-header-inner">
        <button
          type="button"
          className="brand"
          onClick={() => onNavigate('search')}
        >
          <span className="brand-dot" />
          <span>joinup</span>
        </button>
        <nav className="tabs">
          {TABS.filter((tab) => user || tab.id === 'search').map((tab) => (
            <button
              key={tab.id}
              type="button"
              className={`tab ${route === tab.id ? 'is-active' : ''}`}
              onClick={() => onNavigate(tab.id)}
            >
              {tab.label}
            </button>
          ))}
        </nav>
        <div className="header-right">
          {user ? (
            <Avatar
              src={imageUrl(user.avatar_id)}
              name={user.name}
              onClick={() => onNavigate('profile')}
            />
          ) : (
            <button type="button" className="btn btn-sm" onClick={() => onNavigate('login')}>
              войти
            </button>
          )}
        </div>
      </div>
    </header>
  );
}
