import { Outlet, Link, useLocation } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import ThemeToggle from './ThemeToggle';

type NavItem = {
  name: string;
  path: string;
  icon: string;
  exact?: boolean;
};

type NavCategory = {
  title: string;
  items: NavItem[];
};

export default function AdminLayout() {
  const { user, logout } = useAuth();
  const location = useLocation();

  const navCategories: NavCategory[] = [
    {
      title: 'Overview',
      items: [{ name: 'Dashboard', path: '/', exact: true, icon: '📊' }],
    },
    {
      title: 'Analytics',
      items: [
        { name: 'Revenue', path: '/revenue', icon: '💰' },
        { name: 'Delivery Analytics', path: '/delivery', icon: '📈' },
        { name: 'Advertisers', path: '/advertisers', icon: '👥' },
        { name: 'Campaigns', path: '/campaigns', exact: true, icon: '📢' },
      ],
    },
    {
      title: 'Management',
      items: [
        { name: 'Campaign Moderation', path: '/campaigns/pending', icon: '⚡' },
        { name: 'Operator Team', path: '/users', icon: '🛡️' },
      ],
    },
  ];

  const allItems = navCategories.flatMap(c => c.items);

  const currentItem = allItems.find(item =>
    item.exact ? location.pathname === item.path : location.pathname.startsWith(item.path),
  );

  return (
    <div className="admin-shell min-h-screen flex">
      <aside className="admin-sidebar w-64 flex flex-col shrink-0">
        <div className="p-6 border-b border-[var(--sidebar-border)]">
          <Link to="/" className="flex items-center gap-2.5">
            <div className="admin-sidebar-logo">◆</div>
            <div>
              <span className="font-bold text-base tracking-tight text-[var(--sidebar-text)]">Skykin Admin</span>
              <p className="text-[10px] uppercase tracking-wider text-[var(--sidebar-muted)] mt-0.5">Operator console</p>
            </div>
          </Link>
        </div>

        <nav className="flex-1 px-3 py-5 space-y-6 overflow-y-auto">
          {navCategories.map((category, idx) => (
            <div key={idx}>
              <h3 className="px-3 text-[10px] font-bold text-[var(--sidebar-muted)] uppercase tracking-wider mb-2">
                {category.title}
              </h3>
              <div className="space-y-0.5">
                {category.items.map(item => {
                  const isActive = item.exact
                    ? location.pathname === item.path
                    : location.pathname.startsWith(item.path);
                  return (
                    <Link
                      key={item.path}
                      to={item.path}
                      className={`admin-nav-link ${isActive ? 'admin-nav-link-active' : ''}`}
                    >
                      <span className="text-base opacity-90">{item.icon}</span>
                      {item.name}
                    </Link>
                  );
                })}
              </div>
            </div>
          ))}
        </nav>

        <div className="p-4 border-t border-[var(--sidebar-border)] shrink-0">
          <div className="px-3 py-2 mb-2 rounded-lg bg-[var(--sidebar-hover)]">
            <p className="text-sm font-medium text-[var(--sidebar-text)] truncate">{user?.name}</p>
            <p className="text-xs text-[var(--sidebar-muted)] truncate">{user?.email}</p>
          </div>
          <button type="button" onClick={logout} className="admin-nav-link w-full justify-center">
            Sign out
          </button>
        </div>
      </aside>

      <main className="flex-1 flex flex-col min-w-0 overflow-hidden bg-[var(--bg-subtle)]">
        <header className="admin-topbar h-16 flex items-center justify-between px-6 lg:px-8 shrink-0">
          <div>
            <h2 className="text-base font-semibold text-primary">{currentItem?.name || 'Admin Area'}</h2>
            <p className="text-xs text-muted hidden sm:block">Platform administration</p>
          </div>
          <div className="flex items-center gap-3">
            <ThemeToggle variant="header" />
            <button type="button" onClick={logout} className="header-btn">
              Sign out
            </button>
          </div>
        </header>

        <div className="flex-1 overflow-auto p-6 lg:p-8">
          <div className="max-w-6xl mx-auto">
            <Outlet />
          </div>
        </div>
      </main>
    </div>
  );
}
