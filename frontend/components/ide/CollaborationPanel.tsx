'use client';

import { Avatar } from '@/components/ui/Avatar';

interface CollaborationPanelProps { 
  presenceUsers: any[];
}

export function CollaborationPanel({ presenceUsers }: CollaborationPanelProps) {
  return (
    <aside className="bg-slate-800 border-l border-slate-700 flex flex-col w-80">
      <div className="p-3 border-b border-slate-700 flex items-center justify-between">
        <h3 className="text-white font-medium">Collaboration</h3>
      </div>
      <div className="flex-1 overflow-y-auto p-3">
        <h4 className="text-xs font-semibold text-slate-400 uppercase tracking-wider mb-3">Online ({presenceUsers.length})</h4>
        <div className="space-y-2">
          {presenceUsers.map((u) => (
            <div key={u.id} className="flex items-center gap-3 p-2 rounded-lg hover:bg-slate-700">
              <span className="w-8 h-8 rounded-full bg-harbor-100 text-harbor-700 flex items-center justify-center font-medium text-sm">
                {u.username.charAt(0).toUpperCase()}
              </span>
              <div className="flex-1 min-w-0">
                <p className="text-white text-sm truncate">{u.username}</p>
                <p className="text-slate-500 text-xs capitalize">{u.role.toLowerCase()}</p>
              </div>
            </div>
          ))}
        </div>
      </div>
    </aside>
  );
}