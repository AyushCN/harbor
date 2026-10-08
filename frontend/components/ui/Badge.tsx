'use client';

import { cn } from '@/lib/utils';

interface BadgeProps {
  children: React.ReactNode;
  variant?: 'default' | 'success' | 'warning' | 'danger' | 'info' | 'neutral';
  size?: 'sm' | 'md';
  className?: string;
}

export function Badge({ children, variant = 'default', size = 'md', className }: BadgeProps) {
  const variants = {
    default: 'bg-slate-100 text-slate-700',
    success: 'bg-emerald-100 text-emerald-700',
    warning: 'bg-amber-100 text-amber-700',
    danger: 'bg-red-100 text-red-700',
    info: 'bg-harbor-100 text-harbor-700',
    neutral: 'bg-slate-100 text-slate-600',
  };
  
  const sizes = {
    sm: 'px-2 py-0.5 text-xs',
    md: 'px-2.5 py-1 text-sm',
  };

  return (
    <span
      className={cn(
        'inline-flex items-center font-medium rounded-full',
        variants[variant],
        sizes[size],
        className
      )}
    >
      {children}
    </span>
  );
}

export function StatusBadge({ status }: { status: string }) {
  const statusMap: Record<string, { label: string; variant: BadgeProps['variant'] }> = {
    CREATED: { label: 'Created', variant: 'neutral' },
    BUILDING: { label: 'Building', variant: 'info' },
    RUNNING: { label: 'Running', variant: 'success' },
    STOPPED: { label: 'Stopped', variant: 'neutral' },
    SUSPENDED: { label: 'Suspended', variant: 'warning' },
    CRASHED: { label: 'Crashed', variant: 'danger' },
    BUILD_FAILED: { label: 'Build Failed', variant: 'danger' },
    DELETING: { label: 'Deleting', variant: 'warning' },
  };

  const { label, variant } = statusMap[status] || { label: status, variant: 'default' };
  
  return <Badge variant={variant}>{label}</Badge>;
}