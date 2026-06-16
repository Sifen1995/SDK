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
        { name: 'Pending Approvals', path: '/campaigns/pending', icon: '⚡' },
        { name: 'Operator Team', path: '/users', icon: '🛡️' },
      ],
    },
  ];

  const allItems = navCategories.flatMap(c => c.items);

  const currentItem = allItems.find(item =>
    item.exact ? location.pathname === item.path : location.pathname.startsWith(item.path),
  );

  return (
    <div className="min-h-screen flex bg-[#f8f9fa] dark:bg-[#0f0e13]">
      <aside className="w-64 bg-[#2b193d] dark:bg-[#1a0f2e] text-white flex flex-col shrink-0 shadow-xl">
        <div className="p-6">
          <Link to="/" className="flex items-center gap-2">
            <div className="h-8 w-8 rounded-lg bg-white/10 flex items-center justify-center font-bold text-white shadow-inner">
              ◆
            </div>
            <span className="font-bold text-lg tracking-tight text-white">Skykin Admin</span>
          </Link>
        </div>

        <nav className="flex-1 px-4 space-y-6 mt-2 overflow-y-auto pb-6">
          {navCategories.map((category, idx) => (
            <div key={idx}>
              <h3 className="px-4 text-xs font-bold text-white/50 uppercase tracking-wider mb-2">
                {category.title}
              </h3>
              <div className="space-y-1">
                {category.items.map(item => {
                  const isActive = item.exact
                    ? location.pathname === item.path
                    : location.pathname.startsWith(item.path);
                  return (
                    <Link
                      key={item.path}
                      to={item.path}
                      className={`flex items-center gap-3 px-4 py-2.5 text-sm font-medium rounded-xl transition ${
                        isActive
                          ? 'bg-white/15 text-white shadow-sm ring-1 ring-white/10'
                          : 'text-white/70 hover:bg-white/5 hover:text-white'
                      }`}
                    >
                      <span className="text-base">{item.icon}</span>
                      {item.name}
                    </Link>
                  );
                })}
              </div>
            </div>
          ))}
        </nav>

        <div className="p-4 mt-auto border-t border-white/10 shrink-0">
          <div className="px-4 py-3 mb-2">
            <p className="text-sm font-medium text-white truncate">{user?.name}</p>
            <p className="text-xs text-white/50 truncate">{user?.email}</p>
          </div>
          <button
            onClick={logout}
            className="w-full flex items-center px-4 py-2 text-sm font-medium rounded-lg text-white/70 hover:bg-white/5 hover:text-white transition"
          >
            Sign out
          </button>
        </div>
      </aside>

      <main className="flex-1 flex flex-col min-w-0 overflow-hidden">
        <header className="admin-topbar h-16 flex items-center justify-between px-6 lg:px-8 shrink-0">
          <div>
            <h2 className="text-base font-semibold text-primary">{currentItem?.name || 'Admin Area'}</h2>
            <p className="text-xs text-muted hidden sm:block">Operator administration console</p>
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
