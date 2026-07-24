import { useState, type FormEvent } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { Button, Card, CardContent, Input, Label, InlineError, ThemeToggle, SkykinMark, cn } from '@skykin/ui';
import { api } from '../lib/api';
import { ROLE_META, type PortalRole } from '../types';
import { useTheme } from '../context/ThemeContext';

type RegisterRole = 'advertiser' | 'read_only_analyst';
const selectableRoles: RegisterRole[] = ['advertiser', 'read_only_analyst'];
const steps = ['Register & subscribe to a plan', 'Build campaigns with channels & segments', 'Go live after operator approval'];

export default function Register() {
  const [form, setForm] = useState({ name: '', email: '', password: '', companyName: '' });
  const [role, setRole] = useState<RegisterRole>('advertiser');
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');
  const [loading, setLoading] = useState(false);
  const { theme, toggleTheme } = useTheme();
  const navigate = useNavigate();

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setError('');
    setSuccess('');
    setLoading(true);
    try {
      await api.register({ name: form.name, email: form.email, password: form.password, company_name: form.companyName, role });
      setSuccess('Account created. Redirecting to sign in…');
      setTimeout(() => navigate('/login'), 1600);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Registration failed');
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="flex min-h-screen">
      <div className="relative hidden w-[48%] flex-col justify-between overflow-hidden bg-primary p-12 text-primary-foreground lg:flex">
        <div className="pointer-events-none absolute inset-0 opacity-90" style={{ background: 'radial-gradient(circle at 15% 85%, rgb(255 255 255 / 0.12), transparent 48%)' }} />
        <div className="relative">
          <span className="mb-8 flex size-12 items-center justify-center rounded-xl bg-white/15"><SkykinMark className="size-7 text-white" /></span>
          <h1 className="font-display text-4xl font-bold leading-tight">Start advertising on Skykin</h1>
          <p className="mt-4 max-w-md text-lg leading-relaxed text-white/85">
            Create your advertiser account, pick a subscription plan, and launch campaigns that match real user intent signals.
          </p>
        </div>
        <div className="relative space-y-3">
          {steps.map((s, i) => (
            <div key={s} className="flex items-center gap-3 text-sm text-white/90">
              <span className="flex size-7 shrink-0 items-center justify-center rounded-full border border-white/30 bg-white/15 text-xs font-bold">{i + 1}</span>
              {s}
            </div>
          ))}
        </div>
      </div>

      <div className="flex flex-1 flex-col">
        <div className="flex justify-end p-6"><ThemeToggle isDark={theme === 'dark'} onToggle={toggleTheme} /></div>
        <div className="flex flex-1 items-center justify-center px-6 py-8">
          <div className="w-full max-w-lg">
            <div className="mb-6 text-center lg:hidden">
              <span className="mx-auto mb-4 flex size-14 items-center justify-center rounded-2xl bg-identity/12 text-identity"><SkykinMark className="size-8" /></span>
              <h1 className="font-display text-2xl font-bold">Create account</h1>
            </div>
            <div className="mb-6 hidden lg:block">
              <h2 className="font-display text-2xl font-bold">Create advertiser account</h2>
              <p className="mt-1 text-sm text-muted-foreground">Join the Skykin ad network in minutes.</p>
            </div>

            <Card>
              <CardContent className="p-6">
                <form onSubmit={handleSubmit} className="space-y-5">
                  {error && <InlineError message={error} />}
                  {success && (
                    <div className="rounded-md border border-success/30 bg-success-surface px-3 py-2 text-sm text-success">{success}</div>
                  )}

                  <div className="space-y-4">
                    <div className="space-y-1.5"><Label htmlFor="name">Full name</Label><Input id="name" required value={form.name} onChange={e => setForm(f => ({ ...f, name: e.target.value }))} placeholder="Jane Doe" /></div>
                    <div className="space-y-1.5"><Label htmlFor="company">Company name</Label><Input id="company" required value={form.companyName} onChange={e => setForm(f => ({ ...f, companyName: e.target.value }))} placeholder="Acme Inc" /></div>
                    <div className="space-y-1.5"><Label htmlFor="email">Work email</Label><Input id="email" type="email" required value={form.email} onChange={e => setForm(f => ({ ...f, email: e.target.value }))} placeholder="jane@acme.com" /></div>
                    <div className="space-y-1.5"><Label htmlFor="password">Password</Label><Input id="password" type="password" required minLength={8} value={form.password} onChange={e => setForm(f => ({ ...f, password: e.target.value }))} placeholder="Min. 8 characters" /></div>
                  </div>

                  <div>
                    <Label className="mb-2 block">Account role</Label>
                    <div className="grid gap-3 sm:grid-cols-2">
                      {selectableRoles.map(r => {
                        const meta = ROLE_META[r as PortalRole];
                        const selected = role === r;
                        return (
                          <button
                            key={r}
                            type="button"
                            onClick={() => setRole(r)}
                            className={cn(
                              'rounded-lg border p-3 text-left transition-colors',
                              selected ? 'border-identity bg-identity/5 ring-1 ring-identity' : 'border-border hover:border-border',
                            )}
                          >
                            <p className="text-sm font-semibold">{meta.label}</p>
                            <p className="mt-1 text-xs leading-relaxed text-muted-foreground">{meta.description}</p>
                          </button>
                        );
                      })}
                    </div>
                    <p className="mt-2 text-xs text-muted-foreground">Operator admin accounts are provisioned by Skykin — not available for self-registration.</p>
                  </div>

                  <Button type="submit" disabled={loading} className="w-full">{loading ? 'Creating account…' : 'Create account'}</Button>
                  <p className="text-center text-sm text-muted-foreground">
                    Already registered? <Link to="/login" className="font-semibold text-identity hover:underline">Sign in</Link>
                  </p>
                </form>
              </CardContent>
            </Card>
          </div>
        </div>
      </div>
    </div>
  );
}
