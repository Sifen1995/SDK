import { useState, type FormEvent } from 'react';
import { useNavigate } from 'react-router-dom';
import { Button, Card, CardContent, Input, Label, InlineError, SkykinMark } from '@skykin/ui';
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
      setError(err instanceof Error ? err.message : 'Login failed. Check your credentials and try again.');
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="flex min-h-screen">
      <div className="relative hidden w-[46%] flex-col justify-between overflow-hidden bg-primary p-12 text-primary-foreground lg:flex">
        <div
          className="pointer-events-none absolute inset-0 opacity-90"
          style={{ background: 'radial-gradient(circle at 20% 80%, rgb(255 255 255 / 0.12), transparent 45%), radial-gradient(circle at 85% 15%, rgb(255 255 255 / 0.08), transparent 40%)' }}
        />
        <div className="relative">
          <span className="mb-8 flex size-12 items-center justify-center rounded-xl bg-white/15">
            <SkykinMark className="size-7 text-white" />
          </span>
          <h1 className="font-display text-3xl font-bold leading-tight">Skykin Admin</h1>
          <p className="mt-3 max-w-md leading-relaxed text-white/80">
            Monitor platform health, approve campaigns, and analyze delivery performance across the ad network.
          </p>
        </div>
        <div className="relative flex flex-wrap gap-2 text-xs">
          {['Real-time analytics', 'Campaign moderation', 'Operator access'].map(t => (
            <span key={t} className="rounded-full border border-white/25 bg-white/12 px-3 py-1 font-medium">{t}</span>
          ))}
        </div>
      </div>

      <div className="flex flex-1 items-center justify-center px-6 py-12">
        <div className="w-full max-w-md">
          <div className="mb-8 text-center lg:hidden">
            <span className="mx-auto mb-4 flex size-14 items-center justify-center rounded-2xl bg-identity/12 text-identity">
              <SkykinMark className="size-8" />
            </span>
            <h1 className="font-display text-2xl font-bold">Skykin Admin</h1>
          </div>
          <div className="mb-6 hidden lg:block">
            <h2 className="font-display text-xl font-bold">Welcome back</h2>
            <p className="mt-1 text-sm text-muted-foreground">Sign in with your operator admin credentials.</p>
          </div>

          <Card>
            <CardContent className="p-6">
              <form onSubmit={handleSubmit} className="space-y-4">
                {error && <InlineError message={error} />}
                <div className="space-y-1.5">
                  <Label htmlFor="email">Email</Label>
                  <Input id="email" type="email" required value={email} onChange={e => setEmail(e.target.value)} placeholder="admin@skykin.io" autoComplete="email" />
                </div>
                <div className="space-y-1.5">
                  <Label htmlFor="password">Password</Label>
                  <Input id="password" type="password" required value={password} onChange={e => setPassword(e.target.value)} placeholder="••••••••" autoComplete="current-password" />
                </div>
                <Button type="submit" disabled={loading} className="w-full">
                  {loading ? 'Signing in…' : 'Sign in to Admin'}
                </Button>
              </form>
            </CardContent>
          </Card>
          <p className="mt-6 text-center text-xs text-muted-foreground">Restricted to operator administrators only.</p>
        </div>
      </div>
    </div>
  );
}
