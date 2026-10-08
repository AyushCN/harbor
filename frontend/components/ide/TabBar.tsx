'use client';

import { FileCode, X, Plus } from 'lucide-react';
import { Button } from '@/components/ui/Button';

const FILE_ICONS: Record<string, string> = {
  '.js': 'text-yellow-500',
  '.ts': 'text-blue-500',
  '.jsx': 'text-yellow-500',
  '.tsx': 'text-blue-500',
  '.py': 'text-green-500',
  '.go': 'text-cyan-500',
  '.rs': 'text-orange-500',
  '.java': 'text-red-500',
  '.html': 'text-orange-500',
  '.css': 'text-sky-500',
  '.json': 'text-yellow-400',
  '.md': 'text-gray-500',
  '.yml': 'text-red-500',
  '.yaml': 'text-red-500',
  '.sh': 'text-green-500',
  '.dockerfile': 'text-blue-500',
  '.gitignore': 'text-gray-500',
};

function getFileIcon(filename: string) {
  const ext = '.' + filename.split('.').pop()?.toLowerCase();
  return FILE_ICONS[ext] || 'text-slate-500';
}

interface TabButtonProps { 
  path: string; 
  active: boolean; 
  onClick: () => void; 
  onClose: () => void; 
  modified: boolean 
}

export function TabButton({ path, active, onClick, onClose, modified }: any) {
  const filename = path.split('/').pop() || path;
  const ext = '.' + filename.split('.').pop()?.toLowerCase();
  
  return (
    <button
      onClick={onClick}
      className={`flex items-center gap-1.5 px-2 py-1.5 rounded-l text-sm transition-colors max-w-[180px] min-w-[120px] ${active ? 'bg-slate-700 text-white' : 'bg-slate-800 text-slate-300 hover:bg-slate-700 hover:text-white'}`}
    >
      <span className={`text-xs ${getFileIcon(filename)}`}>
        📄
      </span>
      <span className="truncate flex-1">{filename}</span>
      {modified && <span className="w-1.5 h-1.5 bg-harbor-500 rounded-full" />}
      <button 
        onClick={(e) => { e.stopPropagation(); onClose(); }}
        className="ml-1 p-0.5 rounded hover:bg-slate-600 text-slate-400 hover:text-white"
      >
        <X className="w-3 h-3" />
      </button>
    </button>
  );
}

