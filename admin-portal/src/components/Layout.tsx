import { Outlet, useLocation } from 'react-router-dom';
import {
  AppShell,
  SkykinMark,
  ThemeToggle,
  Avatar,
  AvatarFallback,
  Button,
  type NavGroup,
} from '@skykin/ui';
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
  Contact,
  KeyRound,
  LogOut,
} from 'lucide-react';
import { useAuth } from '../context/AuthContext';
import { useTheme } from '../context/ThemeContext';

const groups: NavGroup[] = [
  { label: 'Overview', items: [{ label: 'Dashboard', to: '/', end: true, icon: LayoutDashboard }] },
  {
    label: 'Analytics',
    items: [
      { label: 'Revenue', to: '/revenue', icon: DollarSign },
      { label: 'Delivery', to: '/delivery', icon: TrendingUp },
      { label: 'Advertisers', to: '/advertisers', icon: Users },
      { label: 'SDK Users', to: '/sdk-users', icon: Contact },
      { label: 'Campaigns', to: '/campaigns', end: true, icon: Megaphone },
    ],
  },
  {
    label: 'Management',
    items: [
      { label: 'Moderation', to: '/campaigns/pending', icon: ShieldCheck },
      { label: 'Segment Candidates', to: '/segment-candidates', icon: FileCheck },
      { label: 'Plans & Billing', to: '/plans', icon: Settings },
      { label: 'Operator Team', to: '/users', icon: UserPlus },
      { label: 'Roles & Permissions', to: '/roles', icon: KeyRound },
    ],
  },
];

const titles: { match: (p: string) => boolean; label: string }[] = [
  { match: p => p === '/', label: 'Platform Overview' },
  { match: p => p.startsWith('/revenue'), label: 'Revenue Analytics' },
  { match: p => p.startsWith('/delivery'), label: 'Delivery Analytics' },
  { match: p => p.startsWith('/advertisers'), label: 'Advertisers' },
  { match: p => p.startsWith('/campaigns/pending'), label: 'Moderation Queue' },
  { match: p => p.startsWith('/campaigns'), label: 'Campaigns' },
  { match: p => p.startsWith('/segment-candidates'), label: 'Segment Candidates' },
  { match: p => p.startsWith('/plans'), label: 'Plans & Billing' },
  { match: p => p.startsWith('/users'), label: 'Operator Team' },
];

function initials(name?: string) {
  if (!name) return 'OP';
  return name.trim().split(/\s+/).slice(0, 2).map(w => w[0]?.toUpperCase() ?? '').join('') || 'OP';
}

export default function Layout() {
  const { user, logout } = useAuth();
  const { theme, toggleTheme } = useTheme();
  const { pathname } = useLocation();
  const title = titles.find(t => t.match(pathname))?.label ?? 'Admin';

  return (
    <AppShell
      brand={{ name: 'Skykin', sub: 'Operator', mark: <SkykinMark className="size-5 text-identity" /> }}
      groups={groups}
      title={
        <div>
          <h1 className="font-display text-base font-semibold text-foreground">{title}</h1>
          <p className="hidden text-[11px] text-muted-foreground sm:block">Platform administration</p>
        </div>
      }
      topbarRight={
        <>
          <ThemeToggle isDark={theme === 'dark'} onToggle={toggleTheme} />
          <Avatar>
            <AvatarFallback>{initials(user?.name)}</AvatarFallback>
          </Avatar>
        </>
      }
      sidebarFooter={
        <div className="space-y-1.5">
          <div className="flex items-center gap-2.5 rounded-md px-2 py-1.5">
            <Avatar className="size-8">
              <AvatarFallback className="text-[10px]">{initials(user?.name)}</AvatarFallback>
            </Avatar>
            <div className="min-w-0">
              <p className="truncate text-xs font-medium text-sidebar-foreground">{user?.name}</p>
              <p className="truncate text-[10px] text-sidebar-muted">{user?.email}</p>
            </div>
          </div>
          <Button variant="ghost" size="sm" className="w-full justify-start gap-2 text-sidebar-muted" onClick={logout}>
            <LogOut className="size-4" />
            Sign out
          </Button>
        </div>
      }
    >
      <Outlet />
    </AppShell>
  );
}
