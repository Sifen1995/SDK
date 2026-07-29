import { useState, type FormEvent } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { Check } from 'lucide-react';
import { Button, Card, CardContent, Input, Label, InlineError, SkykinMark } from '@skykin/ui';
import { api } from '../lib/api';
import { AuthHero } from '../components/AuthHero';

export default function Register() {
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setError('');
    setLoading(true);
    try {
      await api.register(name, email, password);
      navigate('/login');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Registration failed');
    } finally {
      setLoading(false);
    }
  }

  const perks = ['Sandbox API keys instantly', 'Full Flutter SDK access', 'Delivery analytics per app'];
  return (
    <div className="flex min-h-screen">
      <AuthHero
        eyebrow="Skykin for Developers"
        title={<>Start building with the Skykin SDK.</>}
        blurb="Create a developer account to register apps, generate API keys, and integrate intent-aware ads into your Flutter product."
        chips={['No credit card', 'Free sandbox', 'Live in minutes']}
      >
        <ul className="mt-8 max-w-md space-y-3">
          {perks.map(p => (
            <li key={p} className="flex items-center gap-3 text-white/90">
              <span className="flex size-5 shrink-0 items-center justify-center rounded-full bg-white/15 ring-1 ring-white/25">
                <Check className="size-3" strokeWidth={3} />
              </span>
              {p}
            </li>
          ))}
        </ul>
      </AuthHero>

      <div className="flex flex-1 items-center justify-center px-6 py-12">
        <div className="w-full max-w-md">
          <div className="mb-8 text-center lg:hidden">
            <span className="mx-auto mb-4 flex size-14 items-center justify-center brand-chip rounded-2xl [&_svg]:!text-white"><SkykinMark className="size-8" /></span>
            <h1 className="font-display text-2xl font-bold">Create your account</h1>
          </div>
          <div className="mb-6 hidden lg:block">
            <h2 className="font-display text-xl font-bold">Create your account</h2>
            <p className="mt-1 text-sm text-muted-foreground">Start building with the Skykin SDK.</p>
          </div>
        <Card>
          <CardContent className="p-6">
            <form onSubmit={handleSubmit} className="space-y-4">
              {error && <InlineError message={error} />}
              <div className="space-y-1.5">
                <Label htmlFor="name">Name</Label>
                <Input id="name" required minLength={2} value={name} onChange={e => setName(e.target.value)} placeholder="Jane Developer" />
              </div>
              <div className="space-y-1.5">
                <Label htmlFor="email">Email</Label>
                <Input id="email" type="email" required value={email} onChange={e => setEmail(e.target.value)} placeholder="you@company.com" autoComplete="email" />
              </div>
              <div className="space-y-1.5">
                <Label htmlFor="password">Password</Label>
                <Input id="password" type="password" required minLength={8} value={password} onChange={e => setPassword(e.target.value)} placeholder="Min. 8 characters" autoComplete="new-password" />
              </div>
              <Button type="submit" disabled={loading} className="w-full">{loading ? 'Creating account…' : 'Create account'}</Button>
              <p className="text-center text-sm text-muted-foreground">
                Already have an account? <Link to="/login" className="font-semibold text-identity hover:underline">Sign in</Link>
              </p>
            </form>
          </CardContent>
        </Card>
        </div>
      </div>
    </div>
  );
}
