'use client';

import { useEffect, useRef, useState, useMemo } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { useAuthStore, useEnvironmentStore } from '@/lib/store';
import { environmentsApi, type PresenceUser } from '@/lib/api';
import { Button } from '@/components/ui/Button';
import { Avatar, AvatarGroup } from '@/components/ui/Avatar';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card';
import { Modal } from '@/components/ui/Modal';
import { useToast } from '@/components/ui/Toast';
import { Input } from '@/components/ui/Input';
import { 
  ChevronLeft, 
  Terminal, 
  Users, 
  Search, 
  FileCode, 
  FolderOpen, 
  Save, 
  X, 
  MoreVertical,
  Settings,
  Menu,
  Maximize2,
  Minimize2,
  Play,
  Square,
  Wifi,
  WifiOff,
  Plus
} from 'lucide-react';

import { Header } from '@/components/ide/Header';
import { Sidebar } from '@/components/ide/Sidebar';
import { TabButton } from '@/components/ide/TabBar';
import { Editor } from '@/components/ide/Editor';
import { TerminalPanel } from '@/components/ide/TerminalPanel';
import { CollaborationPanel } from '@/components/ide/CollaborationPanel';
import { StatusBadge } from '@/components/ide/StatusBadge';

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


function getLanguage(filePath: string): string {
  const ext = '.' + filePath.split('.').pop()?.toLowerCase();
  const languages: Record<string, string> = {
    '.js': 'javascript',
    '.ts': 'typescript',
    '.jsx': 'javascriptreact',
    '.tsx': 'typescriptreact',
    '.py': 'python',
    '.go': 'go',
    '.rs': 'rust',
    '.java': 'java',
    '.html': 'html',
    '.css': 'css',
    '.json': 'json',
    '.md': 'markdown',
    '.yml': 'yaml',
    '.yaml': 'yaml',
    '.sh': 'shell',
    'dockerfile': 'dockerfile',
  };
  return languages[ext] || 'plaintext';
}

export default function IDEPage() {
  const params = useParams();
  const router = useRouter();
  const { user, token } = useAuthStore();
  const { setCurrentEnvironment } = useEnvironmentStore();
  const { success, error } = useToast();
  
  const envId = params.id as string;
  const [env, setEnv] = useState<any>(null);
  const [files, setFiles] = useState<any[]>([]);
  const [openFiles, setOpenFiles] = useState<string[]>([]);
  const [activeFile, setActiveFile] = useState<string | null>(null);
  const [fileContent, setFileContent] = useState<string>('');
  const [fileContentOriginal, setFileContentOriginal] = useState<string>('');
  const [loading, setLoading] = useState(true);
  const [presenceUsers, setPresenceUsers] = useState<PresenceUser[]>([]);
  const [terminalOpen, setTerminalOpen] = useState(false);
  const [sidebarOpen, setSidebarOpen] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');
  const [fileTreeOpen, setFileTreeOpen] = useState(true);
  
  const editorRef = useRef<any>(null);
  const monacoRef = useRef<any>(null);

  useEffect(() => {
    if (!token) {
      router.push('/login');
      return;
    }
    fetchEnvironment();
    loadMonaco();
  }, [token, router]);

  const fetchEnvironment = async () => {
    try {
      const response = await environmentsApi.get(envId);
      setEnv(response.data);
    } catch (err) {
      console.error('Failed to fetch environment:', err);
      router.push('/dashboard');
    } finally {
      setLoading(false);
    }
  };

  const loadMonaco = async () => {
    if (typeof window !== 'undefined' && !monacoRef.current) {
      try {
        const monaco = await import('monaco-editor');
        monacoRef.current = monaco;
        
        // Configure TypeScript/JavaScript if available
        const ts = monaco.languages.typescript as any;
        if (ts?.javascriptDefaults) {
          ts.javascriptDefaults.setCompilerOptions({
            target: ts.ScriptTarget.ES2020,
            allowNonTsExtensions: true,
            moduleResolution: ts.ModuleResolutionKind.NodeJs,
            module: ts.ModuleKind.CommonJS,
            esModuleInterop: true,
            allowSyntheticDefaultImports: true,
            strictNullChecks: true,
          });
        }
      } catch (err) {
        console.error('Failed to load Monaco:', err);
      }
    }
  };

  const fetchContent = async () => {
    try {
      const content = `# ${activeFile}\n\n// File content would be loaded from the environment`;
      setFileContent(content);
      setFileContentOriginal(content);
    } catch (err) {
      console.error('Failed to load file:', err);
    }
  };

  useEffect(() => {
    if (!activeFile) return;
    fetchContent();
  }, [activeFile]);

  const handleFileClick = (file: any) => {
    if (file.type === 'file') {
      if (openFiles.includes(file.path)) {
        setActiveFile(file.path);
      } else {
        setOpenFiles([...openFiles, file.path]);
        setActiveFile(file.path);
      }
    } else {
      setFileTreeOpen(!fileTreeOpen);
    }
  };

  const handleFileChange = (value: string) => {
    setFileContent(value);
  };

  const saveFile = async () => {
    if (!activeFile) return;
    try {
      setFileContentOriginal(fileContent);
      success('File saved');
    } catch (err) {
      error('Failed to save file');
    }
  };

  const closeFile = (path: string) => {
    setOpenFiles(openFiles.filter(f => f !== path));
    if (activeFile === path) {
      setActiveFile(openFiles[openFiles.indexOf(path) - 1] || openFiles[openFiles.indexOf(path) + 1] || null);
    }
  };

  const filteredFiles = useMemo(() => 
    files.filter(f => 
      f.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      f.path.toLowerCase().includes(searchQuery.toLowerCase())
    ), [files, searchQuery]);

  if (loading) {
    return (
      <div className="h-screen flex items-center justify-center bg-slate-50">
        <div className="animate-spin rounded-full h-12 w-12 border-4 border-harbor-600 border-t-transparent"></div>
      </div>
    );
  }

  return (
    <div className="h-screen bg-slate-900 flex flex-col">
      <Header
        env={env}
        envId={envId}
        onBack={() => router.push(`/environments/${envId}`)}
        onTerminalToggle={() => setTerminalOpen(!terminalOpen)}
        terminalOpen={terminalOpen}
        presenceUsers={presenceUsers}
      />
      <div className="flex-1 flex overflow-hidden">
        <Sidebar
          sidebarOpen={sidebarOpen}
          setSidebarOpen={setSidebarOpen}
          filteredFiles={filteredFiles}
          openFiles={openFiles}
          activeFile={activeFile}
          onFileClick={handleFileClick}
        />
        <div className="flex-1 flex flex-col overflow-hidden">
          <div className="bg-slate-800 border-b border-slate-700 h-8 flex items-center px-3 gap-2">
            <div className="flex-1 flex gap-1 overflow-x-auto pb-1">
              {openFiles.map((filePath) => (
                <TabButton
                  key={filePath}
                  path={filePath}
                  active={activeFile === filePath}
                  onClick={() => setActiveFile(filePath)}
                  onClose={() => closeFile(filePath)}
                  modified={fileContent !== fileContentOriginal}
                />
              ))}
              <button className="p-1.5 hover:bg-slate-700 rounded text-slate-400 hover:text-white">
                <Plus className="w-4 h-4" />
              </button>
            </div>

            <div className="flex-1 relative overflow-hidden">
              {activeFile ? (
                <Editor 
                  filePath={activeFile}
                  content={fileContent}
                  onChange={handleFileChange}
                  monaco={monacoRef.current}
                  editorRef={editorRef}
                />
              ) : (
                <div className="h-full flex items-center justify-center text-slate-500">
                  <p>No file open. Select a file from the explorer.</p>
                </div>
              )}
            </div>

            {terminalOpen && (
              <TerminalPanel 
                onClose={() => setTerminalOpen(false)} 
                envId={envId}
              />
            )}
          </div>

          <CollaborationPanel presenceUsers={presenceUsers} />
        </div>
      </div>
    </div>
  );
}


