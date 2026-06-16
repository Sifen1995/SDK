import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import type { ReactNode } from 'react';
import { AuthProvider, useAuth } from './context/AuthContext';
import { SubscriptionProvider, useSubscription } from './context/SubscriptionContext';
import { ThemeProvider } from './context/ThemeContext';
import Layout from './components/Layout';
import Login from './pages/Login';
import Register from './pages/Register';
import Campaigns from './pages/Campaigns';
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
  return token ? <Navigate to="/" replace /> : <>{children}</>;
}

function WriteRoute({ children }: { children: ReactNode }) {
  const { token, canWrite } = useAuth();
  if (!token) return <Navigate to="/login" replace />;
  if (!canWrite) return <Navigate to="/" replace />;
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
  if (!isAdmin) return <Navigate to="/" replace />;
  return <>{children}</>;
}

export default function App() {
  return (
    <ThemeProvider>
      <BrowserRouter>
        <AuthProvider>
          <SubscriptionProvider>
            <Routes>
              <Route path="/login" element={<GuestRoute><Login /></GuestRoute>} />
              <Route path="/register" element={<GuestRoute><Register /></GuestRoute>} />
              <Route element={<Layout />}>
                <Route path="/" element={<ProtectedRoute><Campaigns /></ProtectedRoute>} />
                <Route path="/subscription" element={<ProtectedRoute><Subscription /></ProtectedRoute>} />
                <Route path="/profile" element={<ProtectedRoute><Profile /></ProtectedRoute>} />
                <Route path="/campaigns/new" element={<SubscribedWriteRoute><CampaignNew /></SubscribedWriteRoute>} />
                <Route path="/campaigns/:id" element={<ProtectedRoute><CampaignDetail /></ProtectedRoute>} />
                <Route path="/team" element={<AdminRoute><Team /></AdminRoute>} />
              </Route>
            </Routes>
          </SubscriptionProvider>
        </AuthProvider>
      </BrowserRouter>
    </ThemeProvider>
  );
}
