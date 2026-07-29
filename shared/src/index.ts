// @skykin/ui — shared design system for the Skykin portals.

// utils
export { cn } from './lib/utils';

// primitives
export { Button, buttonVariants, type ButtonProps } from './components/ui/button';
export { Card, CardHeader, CardTitle, CardDescription, CardAction, CardContent, CardFooter } from './components/ui/card';
export { Badge, badgeVariants, type BadgeProps } from './components/ui/badge';
export { Input } from './components/ui/input';
export { Label } from './components/ui/label';
export { Skeleton } from './components/ui/skeleton';
export { Separator } from './components/ui/separator';
export {
  Dialog, DialogTrigger, DialogClose, DialogContent, DialogHeader, DialogFooter,
  DialogTitle, DialogDescription, DialogOverlay, DialogPortal,
} from './components/ui/dialog';
export { Table, TableHeader, TableBody, TableRow, TableHead, TableCell, TableCaption } from './components/ui/table';
export { Tabs, TabsList, TabsTrigger, TabsContent } from './components/ui/tabs';
export { Avatar, AvatarImage, AvatarFallback } from './components/ui/avatar';
export {
  Select, SelectGroup, SelectValue, SelectTrigger, SelectContent, SelectItem,
} from './components/ui/select';
export {
  DropdownMenu, DropdownMenuTrigger, DropdownMenuContent, DropdownMenuItem,
  DropdownMenuLabel, DropdownMenuSeparator, DropdownMenuGroup,
} from './components/ui/dropdown-menu';
export { TooltipProvider, Tooltip, TooltipTrigger, TooltipContent } from './components/ui/tooltip';

// composite components
export { AppShell, type NavItem, type NavGroup, type BrandLock } from './components/app-shell';
export { SkykinMark } from './components/brand-mark';
export { KpiCard, type Trend } from './components/kpi-card';
export { StatusPill, statusTone, type StatusTone } from './components/status-pill';
export { ThemeToggle } from './components/theme-toggle';
export {
  Spinner, LoadingState, EmptyState, ErrorState, InlineError, TableSkeleton,
} from './components/state';

// data layer
export { AppProviders } from './providers';
export { createQueryClient } from './query';
export * from '@tanstack/react-query';
export { DataTable, type ColumnDef, type SortingState } from './table/data-table';
export { useListUrlState, exportToCsv } from './table/url-state';
export { useForm, useStore, type AnyFieldApi } from './form';

// charts
export {
  CHART_COLORS, chartColor, chartAxis, chartGrid, chartTooltip, chartLegend,
} from './charts/chart-theme';
