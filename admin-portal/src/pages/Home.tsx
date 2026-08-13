import { Link } from 'react-router-dom';
import { Button, PortalHome, ThemeToggle } from '@skykin/ui';
import {
  ShieldCheck, MapPin, BarChart3, Layers, CreditCard, KeyRound,
} from 'lucide-react';
import { useTheme } from '../context/ThemeContext';

// Operator-facing. Note the actions block below has no "create account" —
// this portal has no Register page; operator accounts are provisioned from
// within the portal itself via POST /ad-portal/admin/users.
const features = [
  {
    icon: ShieldCheck,
    title: 'Campaign moderation',
    description:
      'Review pending campaigns, approve or reject with notes, and take approved campaigns live.',
  },
  {
    icon: MapPin,
    title: 'Geofence activation',
    description:
      'Advertiser store zones arrive inactive. Approving a campaign activates its linked zones, or activate them individually.',
  },
  {
    icon: BarChart3,
    title: 'Platform analytics',
    description:
      'Revenue and MRR, delivery volume and funnel, advertiser spend, and per-campaign performance.',
  },
  {
    icon: Layers,
    title: 'Segment candidates',
    description:
      'Review intent segments the platform discovers from SDK signals, then approve them into sellable AudienceMart segments.',
  },
  {
    icon: CreditCard,
    title: 'Plans and billing rates',
    description:
      'Create and suspend subscription plans, and set the ETB rate for each billing model.',
  },
  {
    icon: KeyRound,
    title: 'Roles and access',
    description:
      'Provision portal users, define roles, and assign the permissions each role carries.',
  },
];

export default function Home() {
  const { theme, toggleTheme } = useTheme();

  return (
    <PortalHome
      eyebrow="Operator portal"
      headline="Run the Skykin ad platform"
      subhead="Moderate what goes live, activate store geofences, and keep an eye on revenue, delivery and advertiser health."
      utility={<ThemeToggle isDark={theme === 'dark'} onToggle={toggleTheme} />}
      actions={
        <>
          <Button asChild size="lg" variant="secondary">
            <Link to="/login">Operator sign in</Link>
          </Button>
          <span className="self-center text-sm text-white/70">
            Access is provisioned by Skykin — there is no self-registration.
          </span>
        </>
      }
      features={features}
    />
  );
}
