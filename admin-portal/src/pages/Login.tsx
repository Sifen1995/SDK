import { useState, type FormEvent } from 'react';
import { useNavigate } from 'react-router-dom';
import { api } from '../lib/api';
import { useAuth } from '../context/AuthContext';
import ThemeToggle from '../components/ThemeToggle';

export default function Login() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const { login } = useAuth();
  const navigate = useNavigate();

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setError('');
    setLoading(true);
    try {
      const res = await api.login(email, password);
      login(res.token, res.user);
      navigate('/');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Login failed');
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="login-page">
      <div className="hidden lg:flex lg:w-[46%] login-hero flex-col justify-between p-12">
        <div className="login-hero-content">
          <div className="h-12 w-12 rounded-2xl bg-white/15 flex items-center justify-center text-xl font-bold shadow-lg mb-8">
            ◆
          </div>
          <h1 className="text-4xl font-bold leading-tight mb-4">Skykin Admin</h1>
          <p className="text-lg text-white/80 max-w-md leading-relaxed">
            Monitor platform health, approve campaigns, and analyze delivery performance across the ad network.
          </p>
        </div>
        <div className="login-hero-content flex flex-wrap gap-3">
          <span className="stat-pill">📊 Real-time analytics</span>
          <span className="stat-pill">⚡ Campaign moderation</span>
          <span className="stat-pill">🛡️ Operator access</span>
        </div>
      </div>

      <div className="flex-1 flex flex-col min-h-screen">
        <div className="flex justify-end p-6">
          <ThemeToggle />
        </div>

        <div className="flex-1 flex items-center justify-center px-6 pb-12">
          <div className="w-full max-w-md">
            <div className="lg:hidden text-center mb-8">
              <div className="logo-mark mx-auto mb-4 h-14 w-14 rounded-2xl flex items-center justify-center text-xl font-bold shadow-lg">
                ◆
              </div>
              <h1 className="text-3xl font-bold text-primary">Skykin Admin</h1>
            </div>

            <div className="mb-8 hidden lg:block">
              <h2 className="text-2xl font-bold text-primary">Welcome back</h2>
              <p className="text-muted mt-2">Sign in with your operator admin credentials.</p>
            </div>

            <form onSubmit={handleSubmit} className="card-static p-8 space-y-5">
              {error && <div className="alert-error">{error}</div>}

              <div>
                <label className="block text-sm font-medium text-primary mb-1.5">Email</label>
                <input
                  type="email"
                  required
                  value={email}
                  onChange={e => setEmail(e.target.value)}
                  className="field-input"
                  placeholder="admin@skykin.io"
                  autoComplete="email"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-primary mb-1.5">Password</label>
                <input
                  type="password"
                  required
                  value={password}
                  onChange={e => setPassword(e.target.value)}
                  className="field-input"
                  placeholder="••••••••"
                  autoComplete="current-password"
                />
              </div>

              <button type="submit" disabled={loading} className="btn-primary w-full py-3">
                {loading ? 'Signing in…' : 'Sign in to Admin'}
              </button>
            </form>

            <p className="text-center text-xs text-faint mt-6">
              Restricted to operator administrators only.
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}
