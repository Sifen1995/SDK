import { Outlet, useNavigate } from 'react-router-dom';
import { AppShell, SkykinMark, Avatar, AvatarFallback, Button, type NavGroup } from '@skykin/ui';
import { LayoutGrid, PlusCircle, LogOut } from 'lucide-react';
import { useAuth } from '../context/AuthContext';

function initials(name?: string, email?: string) {
  const src = name || email || '';
  if (!src) return 'DV';
  return src.trim().split(/\s+/).slice(0, 2).map(w => w[0]?.toUpperCase() ?? '').join('') || 'DV';
}

const groups: NavGroup[] = [
  {
    label: 'Developer',
    items: [
      { label: 'Applications', to: '/', end: true, icon: LayoutGrid },
      { label: 'New Application', to: '/applications/new', icon: PlusCircle },
    ],
  },
];

export default function Layout() {
  const { developer, logout } = useAuth();
  const navigate = useNavigate();

  // Auth screens own the full viewport — render them without the app shell.
  if (!developer) return <Outlet />;

  function handleLogout() {
    logout();
    navigate('/login');
  }

  const firstName = developer.name?.trim().split(/\s+/)[0] || 'developer';

  return (
    <AppShell
      brand={{ name: 'Skykin', sub: 'Developer', mark: <SkykinMark className="size-5 text-identity" /> }}
      groups={groups}
      title={
        <div>
          <h1 className="font-display text-base font-semibold text-foreground">Hey {firstName}, welcome back</h1>
          <p className="hidden text-[11px] text-muted-foreground sm:block">Manage your applications and SDK credentials</p>
        </div>
      }
      topbarRight={
        <Avatar>
          <AvatarFallback>{initials(developer.name, developer.email)}</AvatarFallback>
        </Avatar>
      }
      sidebarFooter={
        <div className="space-y-1.5">
          <div className="flex items-center gap-2.5 rounded-md px-2 py-1.5">
            <Avatar className="size-8">
              <AvatarFallback className="text-[10px]">{initials(developer.name, developer.email)}</AvatarFallback>
            </Avatar>
            <div className="min-w-0">
              <p className="truncate text-xs font-medium text-sidebar-foreground">{developer.name}</p>
              <p className="truncate text-[10px] text-sidebar-muted">{developer.email}</p>
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
