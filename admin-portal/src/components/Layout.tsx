import { Outlet, Link, useLocation } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import ThemeToggle from './ThemeToggle';
import {
  LayoutDashboard,
  DollarSign,
  TrendingUp,
  Users,
  Megaphone,
  ShieldCheck,
  UserPlus,
  FileCheck,
  Settings,
  LogOut,
  type LucideIcon,
} from 'lucide-react';

type NavItem = {
  name: string;
  path: string;
  icon: LucideIcon;
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
      items: [{ name: 'Dashboard', path: '/', exact: true, icon: LayoutDashboard }],
    },
    {
      title: 'Analytics',
      items: [
        { name: 'Revenue', path: '/revenue', icon: DollarSign },
        { name: 'Delivery', path: '/delivery', icon: TrendingUp },
        { name: 'Advertisers', path: '/advertisers', icon: Users },
        { name: 'Campaigns', path: '/campaigns', exact: true, icon: Megaphone },
      ],
    },
    {
      title: 'Management',
      items: [
        { name: 'Moderation', path: '/campaigns/pending', icon: ShieldCheck },
        { name: 'Segment Candidates', path: '/segment-candidates', icon: FileCheck },
        { name: 'Plans & Billing', path: '/plans', icon: Settings },
        { name: 'Operator Team', path: '/users', icon: UserPlus },
      ],
    },
  ];

  const allItems = navCategories.flatMap(c => c.items);

  const currentItem = allItems.find(item =>
    item.exact ? location.pathname === item.path : location.pathname.startsWith(item.path),
  );

  return (
    <div className="admin-shell min-h-screen flex">
      <aside className="admin-sidebar w-60 flex flex-col shrink-0">
        <div className="p-5 border-b border-[var(--sidebar-border)]">
          <Link to="/" className="flex items-center gap-2.5">
            <div className="admin-sidebar-logo">Sk</div>
            <div>
              <span className="font-semibold text-sm tracking-tight text-[var(--sidebar-text)]">Skykin Admin</span>
              <p className="text-[9px] uppercase tracking-wider text-[var(--sidebar-muted)] mt-0.5">Operator console</p>
            </div>
          </Link>
        </div>

        <nav className="flex-1 px-3 py-4 space-y-5 overflow-y-auto">
          {navCategories.map((category, idx) => (
            <div key={idx}>
              <h3 className="px-3 text-[9px] font-semibold text-[var(--sidebar-muted)] uppercase tracking-wider mb-1.5">
                {category.title}
              </h3>
              <div className="space-y-0.5">
                {category.items.map(item => {
                  const isActive = item.exact
                    ? location.pathname === item.path
                    : location.pathname.startsWith(item.path);
                  const Icon = item.icon;
                  return (
                    <Link
                      key={item.path}
                      to={item.path}
                      className={`admin-nav-link text-[13px] ${isActive ? 'admin-nav-link-active' : ''}`}
                    >
                      <Icon size={15} strokeWidth={2} />
                      {item.name}
                    </Link>
                  );
                })}
              </div>
            </div>
          ))}
        </nav>

        <div className="p-3 border-t border-[var(--sidebar-border)] shrink-0">
          <div className="px-3 py-2 mb-1.5 rounded-lg bg-[var(--sidebar-hover)]">
            <p className="text-xs font-medium text-[var(--sidebar-text)] truncate">{user?.name}</p>
            <p className="text-[10px] text-[var(--sidebar-muted)] truncate">{user?.email}</p>
          </div>
          <button type="button" onClick={logout} className="admin-nav-link w-full justify-center text-[13px]">
            <LogOut size={14} strokeWidth={2} />
            Sign out
          </button>
        </div>
      </aside>

      <main className="flex-1 flex flex-col min-w-0 overflow-hidden bg-[var(--bg-subtle)]">
        <header className="admin-topbar h-14 flex items-center justify-between px-6 lg:px-8 shrink-0">
          <div>
            <h2 className="text-sm font-semibold text-primary">{currentItem?.name || 'Admin Area'}</h2>
            <p className="text-[11px] text-muted hidden sm:block">Platform administration</p>
          </div>
          <div className="flex items-center gap-3">
            <ThemeToggle variant="header" />
            <button type="button" onClick={logout} className="header-btn text-xs">
              <LogOut size={13} strokeWidth={2} />
              Sign out
            </button>
          </div>
        </header>

        <div className="flex-1 overflow-auto p-5 lg:p-7">
          <div className="max-w-6xl mx-auto">
            <Outlet />
          </div>
        </div>
      </main>
    </div>
  );
}
