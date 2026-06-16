import { createContext, useContext, useMemo, useState, type ReactNode } from 'react';
import type { PortalUser } from '../types';
import { canWriteCampaigns, isOperatorAdmin } from '../types';

interface AuthContextValue {
  token: string | null;
  user: PortalUser | null;
  login: (token: string, user: PortalUser) => void;
  logout: () => void;
  refreshUser: (user: PortalUser) => void;
  canWrite: boolean;
  isAdmin: boolean;
}

const AuthContext = createContext<AuthContextValue | null>(null);

function loadUser(): PortalUser | null {
  const raw = localStorage.getItem('adminPortalUser');
  if (!raw) return null;
  try {
    return JSON.parse(raw) as PortalUser;
  } catch {
    return null;
  }
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [token, setToken] = useState<string | null>(() => localStorage.getItem('adminPortalToken'));
  const [user, setUser] = useState<PortalUser | null>(() => loadUser());

  const value = useMemo<AuthContextValue>(() => {
    const role = user?.role ?? 'read_only_analyst';
    return {
      token,
      user,
      login(nextToken, nextUser) {
        localStorage.setItem('adminPortalToken', nextToken);
        localStorage.setItem('adminPortalUser', JSON.stringify(nextUser));
        setToken(nextToken);
        setUser(nextUser);
      },
      logout() {
        localStorage.removeItem('adminPortalToken');
        localStorage.removeItem('adminPortalUser');
        setToken(null);
        setUser(null);
      },
      refreshUser(nextUser) {
        localStorage.setItem('adminPortalUser', JSON.stringify(nextUser));
        setUser(nextUser);
      },
      canWrite: canWriteCampaigns(role),
      isAdmin: isOperatorAdmin(role),
    };
  }, [token, user]);

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used within AuthProvider');
  return ctx;
}
