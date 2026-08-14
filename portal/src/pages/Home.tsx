import { Link } from 'react-router-dom';
import { Button, PortalHome, ThemeToggle } from '@skykin/ui';
import { AppWindow, KeyRound, Smartphone, ShieldCheck } from 'lucide-react';
import { useTheme } from '../context/ThemeContext';

// Copy is limited to what this portal actually does today: register applications
// and issue one key pair each. There is no key rotation, no per-app analytics and
// no key re-fetch — the secret is shown once at creation and never again.
const features = [
  {
    icon: AppWindow,
    title: 'Register applications',
    description:
      'Create an application record with a name and bundle ID for each app you ship the SDK in.',
  },
  {
    icon: KeyRound,
    title: 'Get your key pair',
    description:
      'Each application issues a publishable key and a secret key. The secret is shown once at creation — copy it then.',
  },
  {
    icon: Smartphone,
    title: 'Pick your platform',
    description:
      'Flutter, Android, iOS or web. The Flutter SDK is the reference integration.',
  },
  {
    icon: ShieldCheck,
    title: 'Signed requests',
    description:
      'The publishable key identifies your app; the secret signs request bodies with HMAC-SHA256 so payloads cannot be tampered with.',
  },
];

const steps = [
  'Create a developer account and sign in.',
  'Register an application and copy its publishable and secret keys.',
  'Pass both to SkykinSDK.init in your app and start sending intent signals.',
];

export default function Home() {
  const { theme, toggleTheme } = useTheme();

  return (
    <PortalHome
      eyebrow="Developer portal"
      headline="Ship the Skykin SDK in your app"
      subhead="Register an application, get an API key pair, and start delivering intent-matched ads from on-device signals."
      utility={<ThemeToggle isDark={theme === 'dark'} onToggle={toggleTheme} />}
      actions={
        <>
          <Button asChild size="lg" variant="secondary">
            <Link to="/register">Create developer account</Link>
          </Button>
          <Button
            asChild
            size="lg"
            variant="outline"
            className="border-white/40 bg-transparent text-white hover:bg-white/10 hover:text-white"
          >
            <Link to="/login">Sign in</Link>
          </Button>
        </>
      }
      features={features}
      steps={steps}
    />
  );
}
