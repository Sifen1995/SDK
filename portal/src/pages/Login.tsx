import { useState, type FormEvent } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { Button, Card, CardContent, Input, Label, InlineError, SkykinMark } from '@skykin/ui';
import { api } from '../lib/api';
import { useAuth } from '../context/AuthContext';
import { AuthHero, SdkSnippet } from '../components/AuthHero';

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
      login(res.data.token, res.data.developer);
      navigate('/');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Sign in failed. Check your credentials and try again.');
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="flex min-h-screen">
      <AuthHero
        eyebrow="Skykin for Developers"
        title={<>Ship intent-aware ads in an afternoon.</>}
        blurb="Register your app, drop in the Flutter SDK, and start serving intent-matched ads. Manage API keys, environments, and delivery — all from one console."
        chips={['Flutter SDK', 'API keys in minutes', 'Real-time delivery']}
      >
        <SdkSnippet />
      </AuthHero>

      <div className="flex flex-1 items-center justify-center px-6 py-12">
        <div className="w-full max-w-md">
          <div className="mb-8 text-center lg:hidden">
            <span className="mx-auto mb-4 flex size-14 items-center justify-center brand-chip rounded-2xl [&_svg]:!text-white"><SkykinMark className="size-8" /></span>
            <h1 className="font-display text-2xl font-bold">Skykin Developer Portal</h1>
          </div>
          <div className="mb-6 hidden lg:block">
            <h2 className="font-display text-xl font-bold">Welcome back</h2>
            <p className="mt-1 text-sm text-muted-foreground">Sign in to your Skykin developer account.</p>
          </div>
        <Card>
          <CardContent className="p-6">
            <form onSubmit={handleSubmit} className="space-y-4">
              {error && <InlineError message={error} />}
              <div className="space-y-1.5">
                <Label htmlFor="email">Email</Label>
                <Input id="email" type="email" required value={email} onChange={e => setEmail(e.target.value)} placeholder="you@company.com" autoComplete="email" />
              </div>
              <div className="space-y-1.5">
                <Label htmlFor="password">Password</Label>
                <Input id="password" type="password" required value={password} onChange={e => setPassword(e.target.value)} placeholder="••••••••" autoComplete="current-password" />
              </div>
              <Button type="submit" disabled={loading} className="w-full">{loading ? 'Signing in…' : 'Sign in'}</Button>
              <p className="text-center text-sm text-muted-foreground">
                No account? <Link to="/register" className="font-semibold text-identity hover:underline">Create one</Link>
              </p>
            </form>
          </CardContent>
        </Card>
        </div>
      </div>
    </div>
  );
}
