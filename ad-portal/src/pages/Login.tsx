import { useState, type FormEvent } from 'react';
import { Link, useNavigate } from 'react-router-dom';
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
    <div className="auth-page">
      <div className="hidden lg:flex lg:w-[48%] auth-hero flex-col justify-between p-12">
        <div className="auth-hero-content">
          <div className="auth-logo-mark mb-8">Sk</div>
          <h1 className="text-3xl font-bold leading-tight mb-3 text-white">Skykin Ad Portal</h1>
          <p className="text-base text-white/85 max-w-md leading-relaxed">
            Launch intent-driven campaigns, reach the right audience segments, and track performance across the Skykin network.
          </p>
        </div>
        <div className="auth-hero-content flex flex-wrap gap-3">
          <span className="auth-pill">Intent targeting</span>
          <span className="auth-pill">Audiencemart segments</span>
          <span className="auth-pill">Multi-channel delivery</span>
        </div>
      </div>

      <div className="flex-1 flex flex-col min-h-screen">
        <div className="flex justify-end p-6">
          <ThemeToggle variant="header" />
        </div>

        <div className="flex-1 flex items-center justify-center px-6 pb-12">
          <div className="w-full max-w-md">
            <div className="lg:hidden text-center mb-8">
              <div className="auth-logo-mark mx-auto mb-4">Sk</div>
              <h1 className="text-3xl font-bold text-primary">Skykin Ad Portal</h1>
            </div>

            <div className="mb-8 hidden lg:block">
              <h2 className="text-xl font-bold text-primary">Welcome back</h2>
              <p className="text-muted mt-2">Sign in to manage your campaigns and subscription.</p>
            </div>

            <form onSubmit={handleSubmit} className="auth-card space-y-5">
              {error && <div className="alert-error">{error}</div>}

              <div>
                <label className="block text-sm font-medium text-primary mb-1.5">Work email</label>
                <input
                  type="email"
                  required
                  value={email}
                  onChange={e => setEmail(e.target.value)}
                  className="field-input"
                  placeholder="you@company.com"
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
                {loading ? 'Signing in…' : 'Sign in'}
              </button>

              <p className="text-center text-sm text-muted pt-1">
                New advertiser?{' '}
                <Link to="/register" className="text-brand-600 dark:text-brand-400 font-semibold hover:underline">
                  Create account
                </Link>
              </p>
            </form>
          </div>
        </div>
      </div>
    </div>
  );
}
