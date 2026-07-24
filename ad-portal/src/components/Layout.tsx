import { Outlet, useNavigate } from 'react-router-dom';
import {
  AppShell,
  SkykinMark,
  ThemeToggle,
  Avatar,
  AvatarFallback,
  Button,
  type NavGroup,
} from '@skykin/ui';
import { Megaphone, CreditCard, User, Users, PlusCircle, LogOut } from 'lucide-react';
import { useAuth } from '../context/AuthContext';
import { useSubscription } from '../context/SubscriptionContext';
import { useTheme } from '../context/ThemeContext';

function initials(name?: string) {
  if (!name) return 'AD';
  return name.trim().split(/\s+/).slice(0, 2).map(w => w[0]?.toUpperCase() ?? '').join('') || 'AD';
}

export default function Layout() {
  const { user, logout, isAdmin, canWrite } = useAuth();
  const { subscribed, subscription } = useSubscription();
  const { theme, toggleTheme } = useTheme();
  const navigate = useNavigate();

  function handleLogout() {
    logout();
    navigate('/login');
  }

  const workspace = [
    { label: 'Campaigns', to: '/', end: true, icon: Megaphone },
    ...(canWrite && subscribed ? [{ label: 'New Campaign', to: '/campaigns/new', icon: PlusCircle }] : []),
    {
      label: 'Subscription',
      to: '/subscription',
      icon: CreditCard,
      badge: !subscribed ? <span className="size-2 rounded-full bg-warning" title="Action needed" /> : undefined,
    },
  ];
  const account = [
    { label: 'Profile', to: '/profile', icon: User },
    ...(isAdmin ? [{ label: 'Team', to: '/team', icon: Users }] : []),
  ];
  const groups: NavGroup[] = [
    { label: 'Workspace', items: workspace },
    { label: 'Account', items: account },
  ];

  const firstName = user?.name?.trim().split(/\s+/)[0] || 'there';

  return (
    <AppShell
      brand={{ name: 'Skykin', sub: 'Ad Portal', mark: <SkykinMark className="size-5 text-identity" /> }}
      groups={groups}
      title={
        <div>
          <h1 className="font-display text-base font-semibold text-foreground">Hey {firstName}, welcome back</h1>
          <p className="hidden text-[11px] text-muted-foreground sm:block">
            {user?.company_name ? `${user.company_name} · ` : ''}Manage your intent-targeted campaigns
          </p>
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
              <p className="truncate text-[10px] text-sidebar-muted">
                {subscribed && subscription ? `${subscription.plan.name} plan` : user?.email}
              </p>
            </div>
          </div>
          <Button variant="ghost" size="sm" className="w-full justify-start gap-2 text-sidebar-muted" onClick={handleLogout}>
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
