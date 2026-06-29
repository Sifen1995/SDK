import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider, useAuth } from './context/AuthContext';
import { ThemeProvider } from './context/ThemeContext';
import Layout from './components/Layout';
import Login from './pages/Login';
import AdminDashboard from './pages/AdminDashboard';
import AdminPendingCampaigns from './pages/AdminPendingCampaigns';
import AdminUsers from './pages/AdminUsers';
import AdminRevenueAnalytics from './pages/AdminRevenueAnalytics';
import AdminDeliveryAnalytics from './pages/AdminDeliveryAnalytics';
import AdminAdvertisersAnalytics from './pages/AdminAdvertisersAnalytics';
import AdminCampaignsAnalytics from './pages/AdminCampaignsAnalytics';
import AdminCampaignDetailAnalytics from './pages/AdminCampaignDetailAnalytics';
import AdminSegmentCandidates from './pages/AdminSegmentCandidates';
import AdminSegments from './pages/AdminSegments';
import AdminPlans from './pages/AdminPlans';
import type { ReactNode } from 'react';

function GuestRoute({ children }: { children: ReactNode }) {
  const { token } = useAuth();
  return token ? <Navigate to="/" replace /> : <>{children}</>;
}

function AdminRoute({ children }: { children: ReactNode }) {
  const { token, isAdmin } = useAuth();
  if (!token) return <Navigate to="/login" replace />;
  if (!isAdmin) return <div className="p-8 text-center text-red-500 text-sm">Access Denied. Operator Admin only.</div>;
  return <>{children}</>;
}

export default function App() {
  return (
    <ThemeProvider>
      <BrowserRouter>
        <AuthProvider>
          <Routes>
            <Route path="/login" element={<GuestRoute><Login /></GuestRoute>} />
            <Route path="/" element={<AdminRoute><Layout /></AdminRoute>}>
              <Route index element={<AdminDashboard />} />
              <Route path="revenue" element={<AdminRevenueAnalytics />} />
              <Route path="delivery" element={<AdminDeliveryAnalytics />} />
              <Route path="advertisers" element={<AdminAdvertisersAnalytics />} />
              <Route path="campaigns/pending" element={<AdminPendingCampaigns />} />
              <Route path="campaigns/:id" element={<AdminCampaignDetailAnalytics />} />
              <Route path="campaigns" element={<AdminCampaignsAnalytics />} />
              <Route path="segment-candidates" element={<AdminSegmentCandidates />} />
              <Route path="segments" element={<AdminSegments />} />
              <Route path="plans" element={<AdminPlans />} />
              <Route path="users" element={<AdminUsers />} />
            </Route>
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </AuthProvider>
      </BrowserRouter>
    </ThemeProvider>
  );
}
