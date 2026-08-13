import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { AppProviders } from '@skykin/ui';
import { AuthProvider, useAuth } from './context/AuthContext';
import Layout from './components/Layout';
import Login from './pages/Login';
import Register from './pages/Register';
import Home from './pages/Home';
import Dashboard from './pages/Dashboard';
import NewApplication from './pages/NewApplication';
import type { ReactNode } from 'react';
import { DASHBOARD_PATH } from './routes';

function ProtectedRoute({ children }: { children: ReactNode }) {
  const { token } = useAuth();
  return token ? <>{children}</> : <Navigate to="/login" replace />;
}

function GuestRoute({ children }: { children: ReactNode }) {
  const { token } = useAuth();
  return token ? <Navigate to={DASHBOARD_PATH} replace /> : <>{children}</>;
}

/** Home is public, but a signed-in developer has no use for it. */
function HomeRoute() {
  const { token } = useAuth();
  return token ? <Navigate to={DASHBOARD_PATH} replace /> : <Home />;
}

import { ThemeProvider } from './context/ThemeContext';

export default function App() {
  return (
    <ThemeProvider>
      <BrowserRouter>
        <AuthProvider>
          <AppProviders>
          <Routes>
            {/* Home sits outside Layout: Layout renders bare when there is no
                developer, which would strip the landing page of its chrome. */}
            <Route path="/" element={<HomeRoute />} />
            <Route element={<Layout />}>
              <Route path="/login" element={<GuestRoute><Login /></GuestRoute>} />
              <Route path="/register" element={<GuestRoute><Register /></GuestRoute>} />
              <Route path={DASHBOARD_PATH} element={<ProtectedRoute><Dashboard /></ProtectedRoute>} />
              <Route path="/applications/new" element={<ProtectedRoute><NewApplication /></ProtectedRoute>} />
            </Route>
          </Routes>
          </AppProviders>
        </AuthProvider>
      </BrowserRouter>
    </ThemeProvider>
  );
}
