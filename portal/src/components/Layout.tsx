import { Outlet, Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';

export default function Layout() {
  const { developer, logout } = useAuth();
  const navigate = useNavigate();

  function handleLogout() {
    logout();
    navigate('/login');
  }

  return (
    <div className="min-h-screen bg-gray-950 text-gray-100">
      <nav className="border-b border-gray-800 bg-gray-950/80 backdrop-blur-sm sticky top-0 z-50">
        <div className="max-w-6xl mx-auto px-6 h-16 flex items-center justify-between">
          <Link to="/" className="text-xl font-bold tracking-tight text-white">
            Skykin<span className="text-indigo-400">Portal</span>
          </Link>

          {developer && (
            <div className="flex items-center gap-6">
              <span className="text-sm text-gray-400">{developer.email}</span>
              <button
                onClick={handleLogout}
                className="text-sm text-gray-400 hover:text-white transition-colors cursor-pointer"
              >
                Sign out
              </button>
            </div>
          )}
        </div>
      </nav>

      <main className="max-w-6xl mx-auto px-6 py-10">
        <Outlet />
      </main>
    </div>
  );
}
