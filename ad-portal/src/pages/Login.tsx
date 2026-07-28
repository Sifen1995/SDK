import { useState, type FormEvent } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { Button, Card, CardContent, Input, Label, InlineError, ThemeToggle, SkykinMark } from '@skykin/ui';
import { api } from '../lib/api';
import { useAuth } from '../context/AuthContext';
import { useTheme } from '../context/ThemeContext';

export default function Login() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const { login } = useAuth();
  const { theme, toggleTheme } = useTheme();
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
      <div className="brand-hero hidden w-[48%] flex-col justify-between overflow-hidden p-12 text-white lg:flex">
        <div className="pointer-events-none absolute inset-0 opacity-90" style={{ background: 'radial-gradient(circle at 15% 85%, rgb(255 255 255 / 0.12), transparent 48%), radial-gradient(circle at 85% 15%, rgb(255 255 255 / 0.08), transparent 40%)' }} />
        <div className="relative">
          <span className="mb-8 flex size-12 items-center justify-center rounded-xl bg-white/15"><SkykinMark className="size-7 text-white" /></span>
          <h1 className="font-display text-3xl font-bold leading-tight">Skykin Ad Portal</h1>
          <p className="mt-3 max-w-md leading-relaxed text-white/85">
            Launch intent-driven campaigns, reach the right audience segments, and track performance across the Skykin network.
          </p>
        </div>
        <div className="relative flex flex-wrap gap-2 text-xs">
          {['Intent targeting', 'AudienceMart segments', 'Multi-channel delivery'].map(t => (
            <span key={t} className="rounded-full border border-white/25 bg-white/12 px-3 py-1 font-medium">{t}</span>
          ))}
        </div>
      </div>

      <div className="flex flex-1 flex-col">
        <div className="flex justify-end p-6"><ThemeToggle isDark={theme === 'dark'} onToggle={toggleTheme} /></div>
        <div className="flex flex-1 items-center justify-center px-6 pb-12">
          <div className="w-full max-w-md">
            <div className="mb-8 text-center lg:hidden">
              <span className="mx-auto mb-4 flex size-14 items-center justify-center brand-chip rounded-2xl"><SkykinMark className="size-8" /></span>
              <h1 className="font-display text-2xl font-bold">Skykin Ad Portal</h1>
            </div>
            <div className="mb-6 hidden lg:block">
              <h2 className="font-display text-xl font-bold">Welcome back</h2>
              <p className="mt-1 text-sm text-muted-foreground">Sign in to manage your campaigns and subscription.</p>
            </div>

            <Card>
              <CardContent className="p-6">
                <form onSubmit={handleSubmit} className="space-y-4">
                  {error && <InlineError message={error} />}
                  <div className="space-y-1.5">
                    <Label htmlFor="email">Work email</Label>
                    <Input id="email" type="email" required value={email} onChange={e => setEmail(e.target.value)} placeholder="you@company.com" autoComplete="email" />
                  </div>
                  <div className="space-y-1.5">
                    <Label htmlFor="password">Password</Label>
                    <Input id="password" type="password" required value={password} onChange={e => setPassword(e.target.value)} placeholder="••••••••" autoComplete="current-password" />
                  </div>
                  <Button type="submit" disabled={loading} className="w-full">{loading ? 'Signing in…' : 'Sign in'}</Button>
                  <p className="pt-1 text-center text-sm text-muted-foreground">
                    New advertiser? <Link to="/register" className="font-semibold text-identity hover:underline">Create account</Link>
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
