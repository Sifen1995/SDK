import { Outlet, Link, useLocation } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { useTheme } from '../context/ThemeContext';

export default function AdminLayout() {
  const { user, logout } = useAuth();
  const { theme, toggleTheme } = useTheme();
  const location = useLocation();

  const navItems = [
    { name: 'Dashboard', path: '/admin', exact: true },
    { name: 'Pending Campaigns', path: '/admin/campaigns/pending' },
    { name: 'User Management', path: '/admin/users' },
    { name: 'Audience Segments', path: '/admin/segments' },
    { name: 'Operator Team', path: '/admin/team' },
  ];

  return (
    <div className="min-h-screen flex bg-[#f8f9fa] dark:bg-[#0f0e13]">
      {/* Sidebar - Dark Purple Brand Color */}
      <aside className="w-64 bg-[#2b193d] dark:bg-[#1a0f2e] text-white flex flex-col shrink-0">
        <div className="p-6">
          <Link to="/" className="flex items-center gap-2">
            <div className="h-8 w-8 rounded-lg bg-white/10 flex items-center justify-center font-bold text-white shadow-inner">
              ◆
            </div>
            <span className="font-bold text-lg tracking-tight text-white">Skykin Admin</span>
          </Link>
        </div>

        <nav className="flex-1 px-4 space-y-1 mt-6">
          {navItems.map(item => {
            const isActive = item.exact ? location.pathname === item.path : location.pathname.startsWith(item.path);
            return (
              <Link
                key={item.path}
                to={item.path}
                className={`flex items-center px-4 py-3 text-sm font-medium rounded-xl transition ${
                  isActive
                    ? 'bg-white/10 text-white shadow-sm'
                    : 'text-white/70 hover:bg-white/5 hover:text-white'
                }`}
              >
                {item.name}
              </Link>
            );
          })}
        </nav>

        <div className="p-4 mt-auto border-t border-white/10">
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

      {/* Main content */}
      <main className="flex-1 flex flex-col min-w-0 overflow-hidden">
        {/* Top Header */}
        <header className="h-16 flex items-center justify-between px-8 bg-white dark:bg-[#1c1b22] border-b border-[var(--border)] shrink-0">
          <h2 className="text-sm font-medium text-muted">
            {navItems.find(i => (i.exact ? location.pathname === i.path : location.pathname.startsWith(i.path)))?.name || 'Admin Area'}
          </h2>
          <div className="flex items-center gap-4">
            <button
              onClick={toggleTheme}
              className="p-2 rounded-lg hover:bg-[var(--bg-subtle)] text-muted transition"
              title="Toggle theme"
            >
              {theme === 'dark' ? '☀️ Light' : '🌙 Dark'}
            </button>
            <Link to="/" className="text-sm font-medium text-brand-600 hover:text-brand-700 transition">
              Exit Admin
            </Link>
          </div>
        </header>

        {/* Scrollable Content Area */}
        <div className="flex-1 overflow-auto p-8">
          <div className="max-w-6xl mx-auto">
            <Outlet />
          </div>
        </div>
      </main>
    </div>
  );
}
