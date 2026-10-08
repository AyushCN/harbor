'use client';

import { cn } from '@/lib/utils';

interface StatusBadgeProps {
  status: string;
}

export function StatusBadge({ status }: StatusBadgeProps) {
  const statusMap: Record<string, { label: string; variant: 'success' | 'warning' | 'danger' | 'info' | 'neutral' }> = {
    CREATED: { label: 'Created', variant: 'neutral' },
    BUILDING: { label: 'Building', variant: 'info' },
    RUNNING: { label: 'Running', variant: 'success' },
    STOPPED: { label: 'Stopped', variant: 'neutral' },
    SUSPENDED: { label: 'Suspended', variant: 'warning' },
    CRASHED: { label: 'Crashed', variant: 'danger' },
    BUILD_FAILED: { label: 'Build Failed', variant: 'danger' },
    DELETING: { label: 'Deleting', variant: 'warning' },
  };

  const { label, variant } = statusMap[status] || { label: status, variant: 'neutral' };
  
  const colors = {
    success: 'bg-emerald-500/20 text-emerald-400 border-emerald-500/30',
    warning: 'bg-amber-500/20 text-amber-400 border-amber-500/30',
    danger: 'bg-red-500/20 text-red-400 border-red-500/30',
    info: 'bg-harbor-500/20 text-harbor-400 border-harbor-500/30',
    neutral: 'bg-slate-700 text-slate-300 border-slate-600',
  };

  return (
    <span className={cn(
      'px-2 py-0.5 text-xs font-medium rounded border',
      colors[variant]
    )}>
      {label}
    </span>
  );
}