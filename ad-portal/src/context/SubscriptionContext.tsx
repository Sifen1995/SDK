import { createContext, useCallback, useContext, useEffect, useState, type ReactNode } from 'react';
import { api } from '../lib/api';
import { useAuth } from './AuthContext';
import type { SubscriptionStatus } from '../types';

interface SubscriptionContextValue extends SubscriptionStatus {
  loading: boolean;
  refresh: () => Promise<void>;
}

const defaultValue: SubscriptionContextValue = {
  subscribed: false,
  subscription: null,
  loading: true,
  refresh: async () => {},
};

const SubscriptionContext = createContext<SubscriptionContextValue>(defaultValue);

export function SubscriptionProvider({ children }: { children: ReactNode }) {
  const { token } = useAuth();
  const [status, setStatus] = useState<SubscriptionStatus>({ subscribed: false, subscription: null });
  const [loading, setLoading] = useState(true);

  const refresh = useCallback(async () => {
    if (!token) {
      setStatus({ subscribed: false, subscription: null });
      setLoading(false);
      return;
    }
    setLoading(true);
    try {
      const res = await api.getSubscription();
      setStatus(res);
    } catch {
      setStatus({ subscribed: false, subscription: null });
    } finally {
      setLoading(false);
    }
  }, [token]);

  useEffect(() => {
    refresh();
  }, [refresh]);

  return (
    <SubscriptionContext.Provider value={{ ...status, loading, refresh }}>
      {children}
    </SubscriptionContext.Provider>
  );
}

export function useSubscription() {
  return useContext(SubscriptionContext);
}
