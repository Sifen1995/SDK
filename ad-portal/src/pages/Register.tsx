import { useState, type FormEvent } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { api } from '../lib/api';
import { ROLE_META } from '../types';
import type { PortalRole } from '../types';

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
      setSuccess('Account created. You can sign in now.');
      setTimeout(() => navigate('/login'), 1500);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Registration failed');
    } finally {
      setLoading(false);
    }
  }

  const selectableRoles: RegisterRole[] = ['advertiser', 'read_only_analyst'];

  return (
    <div className="min-h-[75vh] flex items-center justify-center py-8">
      <div className="w-full max-w-lg">
        <div className="text-center mb-8">
          <h1 className="text-3xl font-bold text-gray-900">Create advertiser account</h1>
          <p className="text-gray-500 mt-2">Launch intent-driven campaigns on Skykin</p>
        </div>

        <form onSubmit={handleSubmit} className="card p-8 space-y-5">
          {error && (
            <div className="rounded-lg border border-red-200 bg-red-50 text-red-700 px-4 py-3 text-sm">
              {error}
            </div>
          )}
          {success && (
            <div className="rounded-lg border border-emerald-200 bg-emerald-50 text-emerald-700 px-4 py-3 text-sm">
              {success}
            </div>
          )}

          <div className="grid sm:grid-cols-2 gap-4">
            <div className="sm:col-span-2">
              <label className="block text-sm font-medium text-gray-700 mb-1.5">Full name</label>
              <input required value={name} onChange={e => setName(e.target.value)} className="field-input" placeholder="Jane Doe" />
            </div>
            <div className="sm:col-span-2">
              <label className="block text-sm font-medium text-gray-700 mb-1.5">Company name</label>
              <input required value={companyName} onChange={e => setCompanyName(e.target.value)} className="field-input" placeholder="Acme Inc" />
            </div>
            <div className="sm:col-span-2">
              <label className="block text-sm font-medium text-gray-700 mb-1.5">Work email</label>
              <input type="email" required value={email} onChange={e => setEmail(e.target.value)} className="field-input" placeholder="jane@acme.com" />
            </div>
            <div className="sm:col-span-2">
              <label className="block text-sm font-medium text-gray-700 mb-1.5">Password</label>
              <input type="password" required minLength={8} value={password} onChange={e => setPassword(e.target.value)} className="field-input" placeholder="Min. 8 characters" />
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">Account role</label>
            <div className="grid sm:grid-cols-2 gap-3">
              {selectableRoles.map(r => {
                const meta = ROLE_META[r as PortalRole];
                const selected = role === r;
                return (
                  <button
                    key={r}
                    type="button"
                    onClick={() => setRole(r)}
                    className={`text-left rounded-xl border p-4 transition cursor-pointer ${
                      selected
                        ? 'border-brand-500 bg-brand-50 ring-2 ring-brand-200'
                        : 'border-gray-200 hover:border-brand-200 bg-white'
                    }`}
                  >
                    <p className="font-semibold text-gray-900 text-sm">{meta.label}</p>
                    <p className="text-xs text-gray-500 mt-1 leading-relaxed">{meta.description}</p>
                  </button>
                );
              })}
            </div>
            <p className="text-xs text-gray-400 mt-2">
              Operator admin accounts are provisioned by Skykin — not available for self-registration.
            </p>
          </div>

          <button type="submit" disabled={loading} className="btn-primary w-full py-3">
            {loading ? 'Creating account…' : 'Create account'}
          </button>

          <p className="text-center text-sm text-gray-500">
            Already registered?{' '}
            <Link to="/login" className="text-brand-600 font-medium hover:text-brand-700">
              Sign in
            </Link>
          </p>
        </form>
      </div>
    </div>
  );
}
