'use client';

import { useMemo } from 'react';
import { FileCode, FolderOpen } from 'lucide-react';

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

interface SidebarProps {
  sidebarOpen: boolean;
  setSidebarOpen: (open: boolean) => void;
  filteredFiles: any[];
  openFiles: string[];
  activeFile: string | null;
  onFileClick: (file: any) => void;
}

export function Sidebar({ 
  sidebarOpen, 
  setSidebarOpen, 
  filteredFiles, 
  openFiles, 
  activeFile, 
  onFileClick 
}: SidebarProps) {
  const buildTree = (files: any[]) => {
    const tree: any = {};
    files.forEach((file: any) => {
      const parts = file.path.split('/');
      let current = tree;
      parts.forEach((part: string, i: number) => {
        if (!current[part]) {
          current[part] = { name: part, type: i === parts.length - 1 ? file.type : 'folder', path: file.path, children: {} };
        }
        current = current[part].children;
      });
    });
    return tree;
  };

  const renderTree = (tree: any, depth = 0) => (
    <div>
      {Object.entries(tree).map(([name, node]: [string, any]) => (
        <div key={name}>
          <div 
            className={`flex items-center gap-1 px-2 py-1.5 rounded text-sm transition-colors ${depth > 0 ? 'pl-6' : ''} ${node.type === 'file' && activeFile === node.path ? 'bg-slate-700 text-white' : 'text-slate-300 hover:bg-slate-700 hover:text-white'}`}
            onClick={() => onFileClick(node)}
          >
            {node.type === 'folder' ? (
              <span className="w-4 h-4 text-slate-400">📁</span>
            ) : (
              <span className={`w-4 h-4 ${getFileIcon(node.name)}`}>📄</span>
            )}
            <span className="truncate">{name}</span>
          </div>
          {node.type === 'folder' && node.children && Object.keys(node.children).length > 0 && (
            <div>{renderTree(node.children, depth + 1)}</div>
          )}
        </div>
      ))}
    </div>
  );

  const tree = useMemo(() => buildTree(filteredFiles), [filteredFiles]);
  return <div>{renderTree(tree)}</div>;
}