import { Link, NavLink, Outlet, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import RoleBadge from './RoleBadge';
import ThemeToggle from './ThemeToggle';

export default function Layout() {
  const { user, logout, isAdmin } = useAuth();
  const navigate = useNavigate();

  function handleLogout() {
    logout();
    navigate('/login');
  }

  const navClass = ({ isActive }: { isActive: boolean }) =>
    `px-3 py-2 rounded-lg text-sm font-medium transition ${
      isActive ? 'nav-link-active' : 'nav-link'
    }`;

  return (
    <div className="min-h-screen flex flex-col">
      <header className="sticky top-0 z-20 app-header">
        <div className="mx-auto max-w-6xl px-4 sm:px-6 h-16 flex items-center justify-between gap-4">
          <Link to="/" className="flex items-center gap-2.5 shrink-0">
            <div className="logo-mark h-9 w-9 rounded-xl flex items-center justify-center font-bold text-sm shadow-sm">
              Sk
            </div>
            <div>
              <p className="font-semibold text-primary leading-tight">Skykin</p>
              <p className="text-[11px] text-muted font-medium uppercase tracking-wider">Ad Portal</p>
            </div>
          </Link>

          {user && (
            <nav className="hidden md:flex items-center gap-1">
              <NavLink to="/" end className={navClass}>Campaigns</NavLink>
              <NavLink to="/profile" className={navClass}>Profile</NavLink>
              {isAdmin && <NavLink to="/team" className={navClass}>Team</NavLink>}
            </nav>
          )}

          <div className="flex items-center gap-2 sm:gap-3">
            <ThemeToggle />

            {user ? (
              <>
                <div className="hidden sm:block text-right">
                  <p className="text-sm font-medium text-primary truncate max-w-[140px]">{user.name}</p>
                  <RoleBadge role={user.role} size="sm" />
                </div>
                <button
                  type="button"
                  onClick={handleLogout}
                  className="text-sm text-muted hover:text-primary transition cursor-pointer"
                >
                  Sign out
                </button>
              </>
            ) : (
              <div className="flex items-center gap-2 text-sm">
                <Link to="/login" className="text-muted hover:text-primary">Sign in</Link>
                <Link to="/register" className="btn-primary px-3 py-1.5 text-sm">
                  Register
                </Link>
              </div>
            )}
          </div>
        </div>
      </header>

      <main className="flex-1 mx-auto w-full max-w-6xl px-4 sm:px-6 py-8">
        <Outlet />
      </main>

      <footer className="app-footer py-6 text-center text-xs">
        Skykin Advertiser Portal · Campaign management for intent-driven ads
      </footer>
    </div>
  );
}
