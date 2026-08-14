import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { AppProviders } from '@skykin/ui';
import { AuthProvider, useAuth } from './context/AuthContext';
import { ThemeProvider } from './context/ThemeContext';
import Layout from './components/Layout';
import { OVERVIEW_PATH } from './routes';
import Login from './pages/Login';
import Home from './pages/Home';
import AdminDashboard from './pages/AdminDashboard';
import AdminPendingZones from './pages/AdminPendingZones';
import AdminPendingCampaigns from './pages/AdminPendingCampaigns';
import AdminUsers from './pages/AdminUsers';
import AdminRevenueAnalytics from './pages/AdminRevenueAnalytics';
import AdminDeliveryAnalytics from './pages/AdminDeliveryAnalytics';
import AdminAdvertisersAnalytics from './pages/AdminAdvertisersAnalytics';
import AdminCampaignsAnalytics from './pages/AdminCampaignsAnalytics';
import AdminCampaignDetailAnalytics from './pages/AdminCampaignDetailAnalytics';
import AdminSegmentCandidates from './pages/AdminSegmentCandidates';
import AdminSdkUsers from './pages/AdminSdkUsers';
import AdminPlans from './pages/AdminPlans';
import AdminRoles from './pages/AdminRoles';
import type { ReactNode } from 'react';

function GuestRoute({ children }: { children: ReactNode }) {
  const { token } = useAuth();
  return token ? <Navigate to={OVERVIEW_PATH} replace /> : <>{children}</>;
}

/** Home is public, but a signed-in operator has no use for it. */
function HomeRoute() {
  const { token } = useAuth();
  return token ? <Navigate to={OVERVIEW_PATH} replace /> : <Home />;
}

function AdminRoute({ children }: { children: ReactNode }) {
  const { token } = useAuth();
  if (!token) return <Navigate to="/login" replace />;
  // The isAdmin gate is deliberately absent — every admin route is already
  // RequirePortalRoles("operator_admin") server-side, and re-checking here on a
  // client-decoded role would only add a second, weaker copy of the rule.
  return <>{children}</>;
}

export default function App() {
  return (
    <ThemeProvider>
      <BrowserRouter>
        <AuthProvider>
          <AppProviders>
          <Routes>
            <Route path="/" element={<HomeRoute />} />
            <Route path="/login" element={<GuestRoute><Login /></GuestRoute>} />
            <Route element={<AdminRoute><Layout /></AdminRoute>}>
              <Route path={OVERVIEW_PATH} element={<AdminDashboard />} />
              <Route path="/revenue" element={<AdminRevenueAnalytics />} />
              <Route path="/delivery" element={<AdminDeliveryAnalytics />} />
              <Route path="/advertisers" element={<AdminAdvertisersAnalytics />} />
              <Route path="/campaigns/pending" element={<AdminPendingCampaigns />} />
              <Route path="/campaigns/:id" element={<AdminCampaignDetailAnalytics />} />
              <Route path="/campaigns" element={<AdminCampaignsAnalytics />} />
              <Route path="/geofences/pending" element={<AdminPendingZones />} />
              <Route path="/segment-candidates" element={<AdminSegmentCandidates />} />
              <Route path="/sdk-users" element={<AdminSdkUsers />} />
              <Route path="/segments" element={<Navigate to="/plans" replace />} />
              <Route path="/plans" element={<AdminPlans />} />
              <Route path="/users" element={<AdminUsers />} />
              <Route path="/roles" element={<AdminRoles />} />
            </Route>
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
          </AppProviders>
        </AuthProvider>
      </BrowserRouter>
    </ThemeProvider>
  );
}
