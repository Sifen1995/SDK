import { createContext, useContext, useState, useEffect, type ReactNode } from 'react';
import type { Developer } from '../lib/api';

interface AuthState {
  token: string | null;
  developer: Developer | null;
  login: (token: string, developer: Developer) => void;
  logout: () => void;
}

const AuthContext = createContext<AuthState | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [token, setToken] = useState<string | null>(localStorage.getItem('token'));
  const [developer, setDeveloper] = useState<Developer | null>(() => {
    const stored = localStorage.getItem('developer');
    return stored ? JSON.parse(stored) : null;
  });

  function login(t: string, dev: Developer) {
    localStorage.setItem('token', t);
    localStorage.setItem('developer', JSON.stringify(dev));
    setToken(t);
    setDeveloper(dev);
  }

  function logout() {
    localStorage.removeItem('token');
    localStorage.removeItem('developer');
    setToken(null);
    setDeveloper(null);
  }

  useEffect(() => {
    if (!token) {
      setDeveloper(null);
    }
  }, [token]);

  return (
    <AuthContext.Provider value={{ token, developer, login, logout }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used within AuthProvider');
  return ctx;
}
