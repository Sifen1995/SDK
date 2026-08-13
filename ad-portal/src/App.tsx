import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import type { ReactNode } from 'react';
import { AppProviders } from '@skykin/ui';
import { AuthProvider, useAuth } from './context/AuthContext';
import { SubscriptionProvider, useSubscription } from './context/SubscriptionContext';
import { ThemeProvider } from './context/ThemeContext';
import Layout from './components/Layout';
import { CAMPAIGNS_PATH } from './routes';
import Login from './pages/Login';
import Register from './pages/Register';
import Home from './pages/Home';
import Campaigns from './pages/Campaigns';
import Zones from './pages/Zones';
import CampaignNew from './pages/CampaignNew';
import CampaignDetail from './pages/CampaignDetail';
import Profile from './pages/Profile';
import Team from './pages/Team';
import Subscription from './pages/Subscription';

function ProtectedRoute({ children }: { children: ReactNode }) {
  const { token } = useAuth();
  return token ? <>{children}</> : <Navigate to="/login" replace />;
}

function GuestRoute({ children }: { children: ReactNode }) {
  const { token } = useAuth();
  return token ? <Navigate to={CAMPAIGNS_PATH} replace /> : <>{children}</>;
}

/** Home is public, but a signed-in advertiser has no use for it. */
function HomeRoute() {
  const { token } = useAuth();
  return token ? <Navigate to={CAMPAIGNS_PATH} replace /> : <Home />;
}

function WriteRoute({ children }: { children: ReactNode }) {
  const { token, canWrite } = useAuth();
  if (!token) return <Navigate to="/login" replace />;
  if (!canWrite) return <Navigate to={CAMPAIGNS_PATH} replace />;
  return <>{children}</>;
}

function SubscribedWriteRoute({ children }: { children: ReactNode }) {
  const { subscribed, loading } = useSubscription();
  if (loading) return <p className="text-muted p-8">Checking subscription…</p>;
  if (!subscribed) return <Navigate to="/subscription" replace />;
  return <WriteRoute>{children}</WriteRoute>;
}

function AdminRoute({ children }: { children: ReactNode }) {
  const { token, isAdmin } = useAuth();
  if (!token) return <Navigate to="/login" replace />;
  if (!isAdmin) return <Navigate to={CAMPAIGNS_PATH} replace />;
  return <>{children}</>;
}

export default function App() {
  return (
    <ThemeProvider>
      <BrowserRouter>
        <AuthProvider>
          <SubscriptionProvider>
            <AppProviders>
            <Routes>
              <Route path="/" element={<HomeRoute />} />
              <Route path="/login" element={<GuestRoute><Login /></GuestRoute>} />
              <Route path="/register" element={<GuestRoute><Register /></GuestRoute>} />
              <Route element={<Layout />}>
                <Route path={CAMPAIGNS_PATH} element={<ProtectedRoute><Campaigns /></ProtectedRoute>} />
                <Route path="/subscription" element={<ProtectedRoute><Subscription /></ProtectedRoute>} />
                <Route path="/profile" element={<ProtectedRoute><Profile /></ProtectedRoute>} />
                <Route path="/zones" element={<ProtectedRoute><Zones /></ProtectedRoute>} />
                <Route path="/campaigns/new" element={<SubscribedWriteRoute><CampaignNew /></SubscribedWriteRoute>} />
                <Route path="/campaigns/:id" element={<ProtectedRoute><CampaignDetail /></ProtectedRoute>} />
                <Route path="/team" element={<AdminRoute><Team /></AdminRoute>} />
              </Route>
            </Routes>
            </AppProviders>
          </SubscriptionProvider>
        </AuthProvider>
      </BrowserRouter>
    </ThemeProvider>
  );
}
