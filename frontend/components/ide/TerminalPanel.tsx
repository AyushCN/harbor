'use client';

import { X } from 'lucide-react';

interface TerminalPanelProps { 
  onClose: () => void; 
  envId: string;
}

export function TerminalPanel({ onClose, envId }: any) {
  return (
    <div className="bg-slate-900 border-t border-slate-700 h-48 flex flex-col">
      <div className="flex items-center justify-between px-3 py-2 border-b border-slate-700">
        <span className="text-slate-400 text-sm">Terminal</span>
        <button onClick={onClose} className="p-1 hover:bg-slate-800 rounded text-slate-400 hover:text-white">
          <X className="w-4 h-4" />
        </button>
      </div>
      <div className="flex-1 p-3 font-mono text-sm text-green-400 overflow-auto">
        <div className="flex items-center gap-2">
          <span className="text-green-500">$</span>
          <input className="bg-transparent border-none outline-none text-white flex-1" placeholder="Type a command..." />
        </div>
        <div className="text-slate-500 text-xs mt-2">Terminal would connect to environment container via WebSocket</div>
      </div>
    </div>
  );
}