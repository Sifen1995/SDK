import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider, useAuth } from './context/AuthContext';
import Layout from './components/Layout';
import Login from './pages/Login';
import Register from './pages/Register';
import Dashboard from './pages/Dashboard';
import NewApplication from './pages/NewApplication';
import type { ReactNode } from 'react';

function ProtectedRoute({ children }: { children: ReactNode }) {
  const { token } = useAuth();
  return token ? <>{children}</> : <Navigate to="/login" replace />;
}

function GuestRoute({ children }: { children: ReactNode }) {
  const { token } = useAuth();
  return token ? <Navigate to="/" replace /> : <>{children}</>;
}

export default function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <Routes>
          <Route element={<Layout />}>
            <Route path="/login" element={<GuestRoute><Login /></GuestRoute>} />
            <Route path="/register" element={<GuestRoute><Register /></GuestRoute>} />
            <Route path="/" element={<ProtectedRoute><Dashboard /></ProtectedRoute>} />
            <Route path="/applications/new" element={<ProtectedRoute><NewApplication /></ProtectedRoute>} />
          </Route>
        </Routes>
      </AuthProvider>
    </BrowserRouter>
  );
}
