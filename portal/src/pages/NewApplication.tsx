import { useState, type FormEvent } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Card, CardHeader, CardTitle, CardDescription, CardContent, Button, Input, Label, Badge, InlineError, cn,
} from '@skykin/ui';
import { KeyRound, Copy, Check, ShieldAlert } from 'lucide-react';
import type { Credentials } from '../lib/api';
import { useCreateApplication } from '../lib/queries';
import { DASHBOARD_PATH } from '../routes';

const platforms = ['flutter', 'android', 'ios', 'web'];

export default function NewApplication() {
  const navigate = useNavigate();
  const create = useCreateApplication();
  const [appName, setAppName] = useState('');
  const [platform, setPlatform] = useState('flutter');
  const [bundleId, setBundleId] = useState('');
  const [credentials, setCredentials] = useState<Credentials | null>(null);

  function handleSubmit(e: FormEvent) {
    e.preventDefault();
    create.mutate({ appName, platform, bundleId }, { onSuccess: d => setCredentials(d.credentials) });
  }

  if (credentials) {
    return (
      <div className="mx-auto max-w-2xl">
        <Card>
          <CardContent className="p-6">
            <div className="mb-6 text-center">
              <span className="mx-auto mb-3 flex size-12 items-center justify-center rounded-full bg-identity/12 text-identity"><KeyRound className="size-6" /></span>
              <h2 className="font-display text-xl font-bold">Application created</h2>
              <p className="mt-2 flex items-center justify-center gap-1.5 text-sm text-warning">
                <ShieldAlert className="size-4" /> Save these now — the secret key is shown only once.
              </p>
            </div>
            <div className="space-y-4">
              <CredentialField label="Publishable key" value={credentials.publishable_key} description="Use as the X-API-Key header in your app." />
              <CredentialField label="Secret key" value={credentials.secret_key} description="Use to compute HMAC signatures for request payloads." secret />
            </div>
            <div className="mt-6 flex justify-end">
              <Button onClick={() => navigate(DASHBOARD_PATH)}>Go to dashboard</Button>
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-lg">
      <Card>
        <CardHeader>
          <CardTitle>New application</CardTitle>
          <CardDescription>Register an app to get SDK credentials.</CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-5">
            {create.isError && <InlineError message={(create.error as Error).message} />}
            <div className="space-y-1.5">
              <Label htmlFor="appName">App name</Label>
              <Input id="appName" required minLength={2} value={appName} onChange={e => setAppName(e.target.value)} placeholder="My Awesome App" />
            </div>
            <div>
              <Label className="mb-1.5 block">Platform</Label>
              <div className="grid grid-cols-4 gap-2">
                {platforms.map(p => (
                  <button key={p} type="button" onClick={() => setPlatform(p)}
                    className={cn('rounded-md border py-2 text-sm font-medium capitalize transition-colors', platform === p ? 'border-identity bg-identity/5 text-identity' : 'border-border text-muted-foreground hover:bg-muted')}>
                    {p}
                  </button>
                ))}
              </div>
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="bundleId">Bundle ID</Label>
              <Input id="bundleId" required value={bundleId} onChange={e => setBundleId(e.target.value)} placeholder="com.company.app" className="font-mono" />
            </div>
            <div className="flex gap-3 pt-2">
              <Button type="button" variant="secondary" className="flex-1" onClick={() => navigate(DASHBOARD_PATH)}>Cancel</Button>
              <Button type="submit" className="flex-1" disabled={create.isPending}>{create.isPending ? 'Creating…' : 'Create application'}</Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}

function CredentialField({ label, value, description, secret }: { label: string; value: string; description: string; secret?: boolean }) {
  const [copied, setCopied] = useState(false);
  function copy() {
    navigator.clipboard.writeText(value);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  }
  return (
    <div className="rounded-lg border border-border bg-muted/40 p-4">
      <div className="mb-1.5 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <span className="text-sm font-medium">{label}</span>
          {secret && <Badge variant="warning">shown once</Badge>}
        </div>
        <Button variant="ghost" size="sm" onClick={copy}>
          {copied ? <><Check className="size-3.5" /> Copied</> : <><Copy className="size-3.5" /> Copy</>}
        </Button>
      </div>
      <code className="block break-all font-mono text-sm text-foreground">{value}</code>
      <p className="mt-2 text-xs text-muted-foreground">{description}</p>
    </div>
  );
}
