import { useState, type FormEvent } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { api } from '../lib/api';
import { ROLE_META } from '../types';
import type { PortalRole } from '../types';
import ThemeToggle from '../components/ThemeToggle';

type RegisterRole = 'advertiser' | 'read_only_analyst';

export default function Register() {
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [companyName, setCompanyName] = useState('');
  const [role, setRole] = useState<RegisterRole>('advertiser');
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setError('');
    setSuccess('');
    setLoading(true);
    try {
      await api.register({
        name,
        email,
        password,
        company_name: companyName,
        role,
      });
      setSuccess('Account created successfully. Redirecting to sign in…');
      setTimeout(() => navigate('/login'), 1800);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Registration failed');
    } finally {
      setLoading(false);
    }
  }

  const selectableRoles: RegisterRole[] = ['advertiser', 'read_only_analyst'];

  return (
    <div className="auth-page">
      <div className="hidden lg:flex lg:w-[48%] auth-hero flex-col justify-between p-12">
        <div className="auth-hero-content">
          <div className="auth-logo-mark mb-8">Sk</div>
          <h1 className="text-4xl font-bold leading-tight mb-4 text-white">Start advertising on Skykin</h1>
          <p className="text-lg text-white/85 max-w-md leading-relaxed">
            Create your advertiser account, pick a subscription plan, and launch campaigns that match real user intent signals.
          </p>
        </div>
        <div className="auth-hero-content space-y-3">
          <div className="auth-step">
            <span className="auth-step-num">1</span>
            <span>Register &amp; subscribe to a plan</span>
          </div>
          <div className="auth-step">
            <span className="auth-step-num">2</span>
            <span>Build campaigns with channels &amp; segments</span>
          </div>
          <div className="auth-step">
            <span className="auth-step-num">3</span>
            <span>Go live after operator approval</span>
          </div>
        </div>
      </div>

      <div className="flex-1 flex flex-col min-h-screen">
        <div className="flex justify-end p-6">
          <ThemeToggle variant="header" />
        </div>

        <div className="flex-1 flex items-center justify-center px-6 py-8">
          <div className="w-full max-w-lg">
            <div className="lg:hidden text-center mb-6">
              <div className="auth-logo-mark mx-auto mb-4">Sk</div>
              <h1 className="text-2xl font-bold text-primary">Create account</h1>
            </div>

            <div className="mb-6 hidden lg:block">
              <h2 className="text-2xl font-bold text-primary">Create advertiser account</h2>
              <p className="text-muted mt-2">Join the Skykin ad network in minutes.</p>
            </div>

            <form onSubmit={handleSubmit} className="auth-card space-y-5">
              {error && <div className="alert-error">{error}</div>}
              {success && <div className="alert-success">{success}</div>}

              <div className="grid sm:grid-cols-2 gap-4">
                <div className="sm:col-span-2">
                  <label className="block text-sm font-medium text-primary mb-1.5">Full name</label>
                  <input required value={name} onChange={e => setName(e.target.value)} className="field-input" placeholder="Jane Doe" />
                </div>
                <div className="sm:col-span-2">
                  <label className="block text-sm font-medium text-primary mb-1.5">Company name</label>
                  <input required value={companyName} onChange={e => setCompanyName(e.target.value)} className="field-input" placeholder="Acme Inc" />
                </div>
                <div className="sm:col-span-2">
                  <label className="block text-sm font-medium text-primary mb-1.5">Work email</label>
                  <input type="email" required value={email} onChange={e => setEmail(e.target.value)} className="field-input" placeholder="jane@acme.com" />
                </div>
                <div className="sm:col-span-2">
                  <label className="block text-sm font-medium text-primary mb-1.5">Password</label>
                  <input type="password" required minLength={8} value={password} onChange={e => setPassword(e.target.value)} className="field-input" placeholder="Min. 8 characters" />
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium text-primary mb-2">Account role</label>
                <div className="grid sm:grid-cols-2 gap-3">
                  {selectableRoles.map(r => {
                    const meta = ROLE_META[r as PortalRole];
                    const selected = role === r;
                    return (
                      <button
                        key={r}
                        type="button"
                        onClick={() => setRole(r)}
                        className={`role-card text-left ${selected ? 'role-card-selected' : ''}`}
                      >
                        <p className="font-semibold text-sm text-primary">{meta.label}</p>
                        <p className="text-xs text-muted mt-1 leading-relaxed">{meta.description}</p>
                      </button>
                    );
                  })}
                </div>
                <p className="text-xs text-faint mt-2">
                  Operator admin accounts are provisioned by Skykin — not available for self-registration.
                </p>
              </div>

              <button type="submit" disabled={loading} className="btn-primary w-full py-3">
                {loading ? 'Creating account…' : 'Create account'}
              </button>

              <p className="text-center text-sm text-muted">
                Already registered?{' '}
                <Link to="/login" className="text-brand-600 dark:text-brand-400 font-semibold hover:underline">
                  Sign in
                </Link>
              </p>
            </form>
          </div>
        </div>
      </div>
    </div>
  );
}
