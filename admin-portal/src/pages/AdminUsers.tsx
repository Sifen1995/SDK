import { useState, type FormEvent } from 'react';
import { api } from '../lib/api';

export default function AdminUsers() {
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [role, setRole] = useState<'operator_admin' | 'advertiser' | 'read_only_analyst'>('operator_admin');
  const [companyName, setCompanyName] = useState('');
  
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState<{ type: 'success' | 'error', text: string } | null>(null);

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setMessage(null);
    setLoading(true);

    try {
      await api.createUser({
        name,
        email,
        password,
        role,
        company_name: companyName || undefined,
      });
      setMessage({ type: 'success', text: `User ${email} created successfully.` });
      setName('');
      setEmail('');
      setPassword('');
      setCompanyName('');
    } catch (err) {
      setMessage({ type: 'error', text: err instanceof Error ? err.message : 'Failed to create user' });
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="max-w-2xl">
      <h1 className="text-xl font-bold text-primary mb-1">User Management</h1>
      <p className="text-sm text-muted mb-6">Create new portal users, including operator admins and managed advertisers.</p>

      <form onSubmit={handleSubmit} className="card p-6 sm:p-8 space-y-6 bg-white dark:bg-[#1c1b22]">
        {message && (
          <div className={`p-4 rounded-lg border text-sm ${
            message.type === 'success' 
              ? 'bg-green-50 border-green-200 text-green-700 dark:bg-green-900/20 dark:border-green-800' 
              : 'bg-red-50 border-red-200 text-red-700 dark:bg-red-900/20 dark:border-red-800'
          }`}>
            {message.text}
          </div>
        )}

        <div className="grid sm:grid-cols-2 gap-6">
          <div>
            <label className="block text-sm font-medium text-primary mb-1.5">Full Name</label>
            <input required value={name} onChange={e => setName(e.target.value)} className="field-input" placeholder="Jane Doe" />
          </div>
          <div>
            <label className="block text-sm font-medium text-primary mb-1.5">Email Address</label>
            <input required type="email" value={email} onChange={e => setEmail(e.target.value)} className="field-input" placeholder="jane@skykin.com" />
          </div>
        </div>

        <div>
          <label className="block text-sm font-medium text-primary mb-1.5">Role</label>
          <select value={role} onChange={e => setRole(e.target.value as any)} className="field-input">
            <option value="operator_admin">Operator Admin (Full Access)</option>
            <option value="advertiser">Advertiser</option>
            <option value="read_only_analyst">Read-Only Analyst</option>
          </select>
        </div>

        {role !== 'operator_admin' && (
          <div>
            <label className="block text-sm font-medium text-primary mb-1.5">Company Name</label>
            <input required value={companyName} onChange={e => setCompanyName(e.target.value)} className="field-input" placeholder="Acme Corp" />
          </div>
        )}

        <div>
          <label className="block text-sm font-medium text-primary mb-1.5">Temporary Password</label>
          <input required type="password" value={password} onChange={e => setPassword(e.target.value)} className="field-input" minLength={8} />
          <p className="text-xs text-muted mt-1">Must be at least 8 characters. The user should change this after logging in.</p>
        </div>

        <div className="pt-4 border-t border-[var(--border)]">
          <button type="submit" disabled={loading} className="btn-primary">
            {loading ? 'Creating...' : 'Create User'}
          </button>
        </div>
      </form>
    </div>
  );
}
