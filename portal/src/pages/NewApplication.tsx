import { useState, type FormEvent } from 'react';
import { useNavigate } from 'react-router-dom';
import { api, type Credentials } from '../lib/api';

const platforms = ['flutter', 'android', 'ios', 'web'];

export default function NewApplication() {
  const [appName, setAppName] = useState('');
  const [platform, setPlatform] = useState('flutter');
  const [bundleId, setBundleId] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const [credentials, setCredentials] = useState<Credentials | null>(null);
  const [copied, setCopied] = useState<string | null>(null);
  const navigate = useNavigate();

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setError('');
    setLoading(true);
    try {
      const res = await api.createApplication(appName, platform, bundleId);
      setCredentials(res.data.credentials);
    } catch (err: any) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }

  function copyToClipboard(text: string, label: string) {
    navigator.clipboard.writeText(text);
    setCopied(label);
    setTimeout(() => setCopied(null), 2000);
  }

  if (credentials) {
    return (
      <div className="max-w-2xl mx-auto">
        <div className="bg-gray-900 border border-gray-800 rounded-2xl p-8">
          <div className="text-center mb-6">
            <div className="text-4xl mb-3">🔐</div>
            <h2 className="text-2xl font-bold text-white">Application Created</h2>
            <p className="text-amber-400 text-sm mt-2 font-medium">
              Save these credentials now — the secret key will never be shown again.
            </p>
          </div>

          <div className="space-y-4">
            <CredentialField
              label="Publishable Key"
              value={credentials.publishable_key}
              description="Use this in your mobile app as the X-API-Key header"
              onCopy={() => copyToClipboard(credentials.publishable_key, 'pk')}
              copied={copied === 'pk'}
            />
            <CredentialField
              label="Secret Key"
              value={credentials.secret_key}
              description="Use this to compute HMAC signatures for request payloads"
              onCopy={() => copyToClipboard(credentials.secret_key, 'sk')}
              copied={copied === 'sk'}
              secret
            />
          </div>

          <div className="mt-8 flex justify-end">
            <button
              onClick={() => navigate('/')}
              className="px-6 py-2.5 bg-indigo-600 hover:bg-indigo-500 text-white font-medium rounded-lg transition text-sm cursor-pointer"
            >
              Go to Dashboard
            </button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-lg mx-auto">
      <div className="mb-8">
        <h1 className="text-2xl font-bold text-white">New Application</h1>
        <p className="text-gray-400 mt-1">Register an app to get SDK credentials</p>
      </div>

      <form onSubmit={handleSubmit} className="bg-gray-900 border border-gray-800 rounded-2xl p-8 space-y-5">
        {error && (
          <div className="bg-red-500/10 border border-red-500/20 text-red-400 px-4 py-3 rounded-lg text-sm">
            {error}
          </div>
        )}

        <div>
          <label className="block text-sm font-medium text-gray-300 mb-1.5">App Name</label>
          <input
            type="text"
            required
            minLength={2}
            value={appName}
            onChange={e => setAppName(e.target.value)}
            className="w-full px-4 py-2.5 bg-gray-800 border border-gray-700 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent transition"
            placeholder="My Awesome App"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-300 mb-1.5">Platform</label>
          <div className="grid grid-cols-4 gap-2">
            {platforms.map(p => (
              <button
                type="button"
                key={p}
                onClick={() => setPlatform(p)}
                className={`py-2 rounded-lg text-sm font-medium transition cursor-pointer border ${
                  platform === p
                    ? 'bg-indigo-600 border-indigo-500 text-white'
                    : 'bg-gray-800 border-gray-700 text-gray-400 hover:border-gray-600'
                }`}
              >
                {p}
              </button>
            ))}
          </div>
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-300 mb-1.5">Bundle ID</label>
          <input
            type="text"
            required
            value={bundleId}
            onChange={e => setBundleId(e.target.value)}
            className="w-full px-4 py-2.5 bg-gray-800 border border-gray-700 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent transition font-mono text-sm"
            placeholder="com.company.app"
          />
        </div>

        <div className="flex gap-3 pt-2">
          <button
            type="button"
            onClick={() => navigate('/')}
            className="flex-1 py-2.5 bg-gray-800 hover:bg-gray-700 text-gray-300 font-medium rounded-lg transition text-sm cursor-pointer border border-gray-700"
          >
            Cancel
          </button>
          <button
            type="submit"
            disabled={loading}
            className="flex-1 py-2.5 bg-indigo-600 hover:bg-indigo-500 disabled:opacity-50 disabled:cursor-not-allowed text-white font-medium rounded-lg transition text-sm cursor-pointer"
          >
            {loading ? 'Creating...' : 'Create Application'}
          </button>
        </div>
      </form>
    </div>
  );
}

function CredentialField({
  label, value, description, onCopy, copied, secret,
}: {
  label: string; value: string; description: string;
  onCopy: () => void; copied: boolean; secret?: boolean;
}) {
  return (
    <div className="bg-gray-800 border border-gray-700 rounded-lg p-4">
      <div className="flex items-center justify-between mb-1">
        <span className="text-sm font-medium text-gray-300">{label}</span>
        <button
          onClick={onCopy}
          className="text-xs text-indigo-400 hover:text-indigo-300 transition cursor-pointer"
        >
          {copied ? 'Copied!' : 'Copy'}
        </button>
      </div>
      <code className={`block text-sm font-mono break-all ${secret ? 'text-amber-400' : 'text-green-400'}`}>
        {value}
      </code>
      <p className="text-xs text-gray-500 mt-2">{description}</p>
    </div>
  );
}
