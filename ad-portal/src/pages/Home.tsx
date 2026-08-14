import { Link } from 'react-router-dom';
import { Button, PortalHome, ThemeToggle } from '@skykin/ui';
import { Target, Radio, MapPin, Wallet, Eye, CreditCard } from 'lucide-react';
import { useTheme } from '../context/ThemeContext';

// Only features that exist today. Deliberately absent: campaign editing, pausing
// and per-campaign analytics — none of which the backend or this portal supports.
const features = [
  {
    icon: Target,
    title: 'Target real intent',
    description:
      'Reach users by intent predicted on their own device, or attach a purchased AudienceMart segment when your plan includes it.',
  },
  {
    icon: Radio,
    title: 'Four delivery channels',
    description:
      'In-app banner, push, SMS+ and native feed. Each channel enforces its own creative rules as you build.',
  },
  {
    icon: MapPin,
    title: 'Store geofences',
    description:
      'Draw a radius around your locations and link them to a campaign. Customers who walk into the zone get the offer.',
  },
  {
    icon: Wallet,
    title: 'Budgets and frequency caps',
    description:
      'Set daily and total spend limits plus a per-user daily frequency cap. Spend is tracked against your plan.',
  },
  {
    icon: Eye,
    title: 'Preview before you submit',
    description:
      'See exactly how your creative renders in the chosen channel before it goes to review.',
  },
  {
    icon: CreditCard,
    title: 'Plan-based subscription',
    description:
      'Choose a plan for your campaign slots, monthly impressions and daily budget ceiling. All campaigns bill per click.',
  },
];

const steps = [
  'Create an advertiser account and subscribe to a plan.',
  'Build a campaign: audience, channel, creative, budget — and link store zones if you have them.',
  'An operator reviews it. On approval the campaign goes live and linked zones activate.',
];

export default function Home() {
  const { theme, toggleTheme } = useTheme();

  return (
    <PortalHome
      eyebrow="Advertiser portal"
      headline="Reach customers at the moment of intent"
      subhead="Build campaigns that match what people are actually interested in — predicted on their device, never from data you have to collect yourself."
      utility={<ThemeToggle isDark={theme === 'dark'} onToggle={toggleTheme} />}
      actions={
        <>
          <Button asChild size="lg" variant="secondary">
            <Link to="/register">Create advertiser account</Link>
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
