import { useState, type FormEvent } from 'react';
import { api } from '../lib/api';
import { useAuth } from '../context/AuthContext';
import RoleBadge from '../components/RoleBadge';
import { ROLE_META } from '../types';
import type { PortalRole } from '../types';

const ADMIN_ROLES: PortalRole[] = ['advertiser', 'read_only_analyst', 'operator_admin'];

export default function Team() {
  const { user, isAdmin } = useAuth();
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [companyName, setCompanyName] = useState(user?.company_name ?? '');
  const [role, setRole] = useState<PortalRole>('advertiser');
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [loading, setLoading] = useState(false);

  if (!isAdmin) {
    return (
      <div className="card p-8 text-center max-w-lg mx-auto">
        <h1 className="text-lg font-semibold text-gray-900">Access restricted</h1>
        <p className="text-gray-500 mt-2">Only operator admins can manage team accounts.</p>
      </div>
    );
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setError('');
    setSuccess('');
    setLoading(true);
    try {
      const res = await api.createUser({
        name,
        email,
        password,
        company_name: companyName || undefined,
        role,
      });
      setSuccess(`Created account for ${res.user.email} (${ROLE_META[res.user.role].label})`);
      setName('');
      setEmail('');
      setPassword('');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create user');
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="max-w-2xl">
      <h1 className="text-2xl font-bold text-gray-900">Team management</h1>
      <p className="text-gray-500 mt-1">Provision advertiser, analyst, or operator admin accounts</p>

      <div className="card mt-8 p-6 sm:p-8">
        <form onSubmit={handleSubmit} className="space-y-5">
          {error && (
            <div className="rounded-lg border border-red-200 bg-red-50 text-red-700 px-4 py-3 text-sm">{error}</div>
          )}
          {success && (
            <div className="rounded-lg border border-emerald-200 bg-emerald-50 text-emerald-700 px-4 py-3 text-sm">{success}</div>
          )}

          <div className="grid sm:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1.5">Name</label>
              <input required value={name} onChange={e => setName(e.target.value)} className="field-input" />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1.5">Email</label>
              <input type="email" required value={email} onChange={e => setEmail(e.target.value)} className="field-input" />
            </div>
            <div className="sm:col-span-2">
              <label className="block text-sm font-medium text-gray-700 mb-1.5">Password</label>
              <input type="password" required minLength={8} value={password} onChange={e => setPassword(e.target.value)} className="field-input" />
            </div>
            <div className="sm:col-span-2">
              <label className="block text-sm font-medium text-gray-700 mb-1.5">Company name</label>
              <input value={companyName} onChange={e => setCompanyName(e.target.value)} className="field-input" />
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">Role</label>
            <div className="space-y-2">
              {ADMIN_ROLES.map(r => (
                <label
                  key={r}
                  className={`flex items-start gap-3 rounded-xl border p-4 cursor-pointer transition ${
                    role === r ? 'border-brand-500 bg-brand-50' : 'border-gray-200 hover:border-brand-200'
                  }`}
                >
                  <input
                    type="radio"
                    name="role"
                    value={r}
                    checked={role === r}
                    onChange={() => setRole(r)}
                    className="mt-1 accent-brand-600"
                  />
                  <div>
                    <div className="flex items-center gap-2">
                      <span className="font-medium text-sm text-gray-900">{ROLE_META[r].label}</span>
                      <RoleBadge role={r} size="sm" />
                    </div>
                    <p className="text-xs text-gray-500 mt-1">{ROLE_META[r].description}</p>
                  </div>
                </label>
              ))}
            </div>
          </div>

          <button type="submit" disabled={loading} className="btn-primary">
            {loading ? 'Creating…' : 'Create user'}
          </button>
        </form>
      </div>
    </div>
  );
}
