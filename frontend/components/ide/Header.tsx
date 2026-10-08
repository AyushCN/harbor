'use client';

import { ChevronLeft, Terminal, Users, Globe } from 'lucide-react';

interface HeaderProps {
  env: any;
  envId: string;
  onBack: () => void;
  onTerminalToggle: () => void;
  terminalOpen: boolean;
  presenceUsers: any[];
}

export function Header({ env, envId, onBack, onTerminalToggle, terminalOpen, presenceUsers }: HeaderProps) {
  return (
    <div data-header className="bg-slate-800 border-b border-slate-700 h-10 flex items-center justify-between px-4">
      <div className="flex items-center gap-3">
        <button onClick={onBack} className="p-1.5 rounded hover:bg-slate-700">
          <ChevronLeft className="w-5 h-5 text-slate-400" />
        </button>
        <div className="flex items-center gap-2">
          <Terminal className="w-5 h-5 text-harbor-500" />
          <span className="text-white font-medium truncate max-w-[200px]">
            {env?.name || env?.workspace?.name || 'Environment'}
          </span>
          <span className="px-2 py-0.5 text-xs font-medium rounded border bg-slate-700 text-slate-300 border-slate-600">
            {env?.status || 'UNKNOWN'}
          </span>
        </div>
        <div className="flex items-center gap-2 ml-auto">
          <button className="p-1.5 hover:bg-slate-700 rounded text-slate-400 hover:text-white" onClick={onTerminalToggle}>
            <Terminal className="w-4 h-4 mr-1" />
            <span>Terminal</span>
          </button>
          <div className="flex items-center gap-1">
            {presenceUsers.slice(0, 3).map((u: any, i: number) => (
              <div key={i} className="w-8 h-8 rounded-full bg-harbor-100 text-harbor-700 flex items-center justify-center font-medium text-sm ring-2 ring-white -ml-1">
                {u.username?.charAt(0)?.toUpperCase() || '?'}
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}
