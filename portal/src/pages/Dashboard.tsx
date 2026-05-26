import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { api, type Application } from '../lib/api';
import { useAuth } from '../context/AuthContext';

const platformColors: Record<string, string> = {
  flutter: 'bg-sky-500/10 text-sky-400 border-sky-500/20',
  android: 'bg-green-500/10 text-green-400 border-green-500/20',
  ios: 'bg-gray-500/10 text-gray-300 border-gray-500/20',
  web: 'bg-purple-500/10 text-purple-400 border-purple-500/20',
};

export default function Dashboard() {
  const { developer } = useAuth();
  const [apps, setApps] = useState<Application[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    api.getApplications()
      .then(res => setApps(res.data.applications || []))
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  return (
    <div>
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-2xl font-bold text-white">
            Welcome, {developer?.name}
          </h1>
          <p className="text-gray-400 mt-1">Manage your applications and SDK credentials</p>
        </div>
        <Link
          to="/applications/new"
          className="px-5 py-2.5 bg-indigo-600 hover:bg-indigo-500 text-white font-medium rounded-lg transition text-sm"
        >
          + New Application
        </Link>
      </div>

      {loading ? (
        <div className="flex items-center justify-center py-20">
          <div className="h-8 w-8 border-2 border-indigo-500 border-t-transparent rounded-full animate-spin" />
        </div>
      ) : apps.length === 0 ? (
        <div className="bg-gray-900 border border-gray-800 rounded-2xl p-12 text-center">
          <div className="text-5xl mb-4">📦</div>
          <h2 className="text-xl font-semibold text-white mb-2">No applications yet</h2>
          <p className="text-gray-400 mb-6">Create your first application to get SDK credentials</p>
          <Link
            to="/applications/new"
            className="inline-block px-6 py-2.5 bg-indigo-600 hover:bg-indigo-500 text-white font-medium rounded-lg transition text-sm"
          >
            Create Application
          </Link>
        </div>
      ) : (
        <div className="grid gap-4">
          {apps.map(app => (
            <div
              key={app.id}
              className="bg-gray-900 border border-gray-800 rounded-xl p-6 hover:border-gray-700 transition"
            >
              <div className="flex items-start justify-between">
                <div>
                  <h3 className="text-lg font-semibold text-white">{app.app_name}</h3>
                  <p className="text-sm text-gray-400 mt-1 font-mono">{app.bundle_id}</p>
                </div>
                <div className="flex items-center gap-3">
                  <span
                    className={`text-xs font-medium px-2.5 py-1 rounded-full border ${
                      platformColors[app.platform] || 'bg-gray-500/10 text-gray-400 border-gray-500/20'
                    }`}
                  >
                    {app.platform}
                  </span>
                  <span
                    className={`text-xs font-medium px-2.5 py-1 rounded-full border ${
                      app.status === 'active'
                        ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20'
                        : 'bg-amber-500/10 text-amber-400 border-amber-500/20'
                    }`}
                  >
                    {app.status}
                  </span>
                </div>
              </div>
              <div className="mt-4 text-xs text-gray-500">
                Created {new Date(app.created_at).toLocaleDateString('en-US', {
                  year: 'numeric', month: 'long', day: 'numeric',
                })}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
