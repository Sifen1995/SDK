import { useState, type FormEvent } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { api } from '../lib/api';
import { useAuth } from '../context/AuthContext';

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
    <div className="min-h-[75vh] flex items-center justify-center">
      <div className="w-full max-w-md">
        <div className="text-center mb-8">
          <div className="logo-mark mx-auto mb-4 h-14 w-14 rounded-2xl flex items-center justify-center text-xl font-bold shadow-lg">
            Sk
          </div>
          <h1 className="text-3xl font-bold text-primary">Welcome back</h1>
          <p className="text-muted mt-2">Sign in to manage your ad campaigns</p>
        </div>

        <form onSubmit={handleSubmit} className="card p-8 space-y-5">
          {error && <div className="alert-error">{error}</div>}

          <div>
            <label className="block text-sm font-medium text-primary mb-1.5">Email</label>
            <input
              type="email"
              required
              value={email}
              onChange={e => setEmail(e.target.value)}
              className="field-input"
              placeholder="you@company.com"
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
            />
          </div>

          <button type="submit" disabled={loading} className="btn-primary w-full py-3">
            {loading ? 'Signing in…' : 'Sign in'}
          </button>

          <p className="text-center text-sm text-muted">
            New advertiser?{' '}
            <Link to="/register" className="text-primary font-medium hover:underline">
              Create account
            </Link>
          </p>
        </form>
      </div>
    </div>
  );
}
