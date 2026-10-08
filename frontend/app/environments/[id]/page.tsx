'use client';

import { useEffect, useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { useAuthStore, useEnvironmentStore } from '@/lib/store';
import { environmentsApi, type Environment, type Member } from '@/lib/api';
import { Button } from '@/components/ui/Button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card';
import { Badge, StatusBadge } from '@/components/ui/Badge';
import { Avatar, AvatarGroup } from '@/components/ui/Avatar';
import { Modal } from '@/components/ui/Modal';
import { useToast } from '@/components/ui/Toast';
import { 
  Github, 
  Users, 
  Play, 
  Square, 
  RotateCcw, 
  Trash2, 
  Share2, 
  ExternalLink,
  Settings,
  Loader2,
  ChevronLeft,
  Terminal,
  Activity,
  Globe,
  Plus,
  MoreVertical,
  AlertTriangle,
  Search
} from 'lucide-react';

export default function EnvironmentDetailPage() {
  const params = useParams();
  const router = useRouter();
  const { user, token } = useAuthStore();
  const { setCurrentEnvironment } = useEnvironmentStore();
  const { success, error } = useToast();
  
  const envId = params.id as string;
  const [env, setEnv] = useState<Environment | null>(null);
  const [members, setMembers] = useState<Member[]>([]);
  const [loading, setLoading] = useState(true);
  const [actionLoading, setActionLoading] = useState<string | null>(null);
  const [showShareModal, setShowShareModal] = useState(false);
  const [shareForm, setShareForm] = useState({ role: 'VIEWER' as 'COLLABORATOR' | 'VIEWER', expires_in_hours: 24, max_uses: 10 });
  const [shareLink, setShareLink] = useState<string | null>(null);

  useEffect(() => {
    if (!token) {
      router.push('/login');
      return;
    }
    fetchData();
  }, [token, router]);

  const fetchData = async () => {
    try {
      const [envRes, membersRes] = await Promise.all([
        environmentsApi.get(envId),
        environmentsApi.listMembers(envId),
      ]);
      setEnv(envRes.data);
      setMembers(membersRes.data);
      setCurrentEnvironment(envRes.data);
    } catch (err) {
      console.error('Failed to fetch environment:', err);
      error('Failed to load environment');
      router.push('/dashboard');
    } finally {
      setLoading(false);
    }
  };

  const handleAction = async (action: 'start' | 'stop' | 'resume' | 'delete') => {
    setActionLoading(action);
    try {
      if (action === 'start') await environmentsApi.start(envId);
      else if (action === 'stop') await environmentsApi.stop(envId);
      else if (action === 'resume') await environmentsApi.resume(envId);
      else if (action === 'delete') await environmentsApi.delete(envId);
      
      success(`${action.charAt(0).toUpperCase() + action.slice(1)} requested`);
      if (action !== 'delete') {
        fetchData();
      } else {
        router.push('/dashboard');
      }
    } catch (err: any) {
      error(`Failed to ${action} environment`, err.response?.data?.error?.message || 'Unknown error');
    } finally {
      setActionLoading(null);
    }
  };

  const handleCreateShareLink = async () => {
    try {
      const response = await environmentsApi.createShareLink(envId, shareForm);
      setShareLink(response.data.url);
      success('Share link created');
    } catch (err: any) {
      error('Failed to create share link', err.response?.data?.error?.message || 'Unknown error');
    }
  };

  if (!token) {
    return <div className="flex items-center justify-center min-h-screen"><Loader2 className="w-8 h-8 animate-spin text-harbor-600" /></div>;
  }

  if (loading || !env) {
    return (
      <div className="min-h-screen bg-slate-50 flex items-center justify-center">
        <Loader2 className="w-8 h-8 animate-spin text-harbor-600" />
      </div>
    );
  }

  const isOwner = env.role === 'OWNER';
  const isRunning = env.status === 'RUNNING';
  const canStart = ['CREATED', 'STOPPED', 'SUSPENDED'].includes(env.status) && env.role !== 'VIEWER';
  const canStop = env.status === 'RUNNING' && isOwner;
  const canResume = env.status === 'SUSPENDED' && env.role !== 'VIEWER';

  return (
    <div className="min-h-screen bg-slate-50">
      <header className="bg-white border-b border-slate-200 sticky top-0 z-40">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-16">
            <div className="flex items-center gap-4">
              <button onClick={() => router.push('/dashboard')} className="p-2 rounded-lg hover:bg-slate-100">
                <ChevronLeft className="w-5 h-5" />
              </button>
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 bg-harbor-600 rounded-lg flex items-center justify-center">
                  <Terminal className="w-6 h-6 text-white" />
                </div>
                <div>
                  <h1 className="text-xl font-bold text-slate-900 truncate max-w-md">
                    {env.name || env.workspace?.name || 'Unnamed Environment'}
                  </h1>
                  <div className="flex items-center gap-2 text-sm text-slate-500">
                    <StatusBadge status={env.status} />
                    {env.role === 'OWNER' && <span className="px-2 py-0.5 text-xs bg-harbor-100 text-harbor-700 rounded-full">Owner</span>}
                  </div>
                </div>
              </div>
            </div>
            <div className="flex items-center gap-3">
              {env.public_url && (
                <Button variant="secondary" onClick={() => window.open(env.public_url!, '_blank')}>
                  <Globe className="w-4 h-4 mr-2" />
                  Open Preview
                </Button>
              )}
              <Button onClick={() => router.push(`/ide/${env.id}`)}>
                <Terminal className="w-4 h-4 mr-2" />
                Open IDE
              </Button>
              {isOwner && (
                <>
                  <Button variant="secondary" onClick={() => router.push(`/environments/${env.id}/members`)}>
                    <Users className="w-4 h-4 mr-2" />
                    Members
                  </Button>
                  <Button variant="secondary" onClick={() => setShowShareModal(true)}>
                    <Plus className="w-4 h-4 mr-2" />
                    Share
                  </Button>
                </>
              )}
              <Button variant="ghost" size="sm" onClick={() => router.push('/dashboard')}>
                <Settings className="w-4 h-4 mr-1" />
                Dashboard
              </Button>
            </div>
          </div>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {loading && <div className="h-4 bg-slate-200 animate-pulse rounded mb-6" />}

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          <div className="lg:col-span-2 space-y-6">
            <Card>
              <CardHeader className="flex flex-row items-center justify-between">
                <CardTitle>Environment Details</CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <DetailRow label="Repository" value={env.workspace?.git_url || env.git_url || 'Unknown'} copyable />
                  <DetailRow label="Branch" value={env.workspace?.git_branch || env.git_branch || 'main'} />
                  <DetailRow label="Created" value={new Date(env.created_at).toLocaleString()} />
                  <DetailRow label="Host" value={env.host?.username || 'Unknown'} avatar={env.host?.avatar_url} />
                </div>
                
                {env.public_url && (
                  <div className="pt-4 border-t border-slate-100">
                    <h4 className="font-medium text-slate-900 mb-3">Preview URL</h4>
                    <div className="flex items-center gap-3">
                      <span className="flex-1 font-mono text-sm bg-slate-100 px-3 py-2 rounded-lg">{env.public_url}</span>
                      <Button variant="secondary" size="sm" onClick={() => window.open(env.public_url!, '_blank')}>
                        <Globe className="w-4 h-4 mr-1" />
                        Open
                      </Button>
                    </div>
                  </div>
                )}

                <div className="pt-4 border-t border-slate-100">
                  <h4 className="font-medium text-slate-900 mb-3">Actions</h4>
                  <div className="flex flex-wrap gap-3">
                    <Button 
                      onClick={() => handleAction('start')} 
                      disabled={!canStart || actionLoading === 'start'}
                      loading={actionLoading === 'start'}
                    >
                      <Play className="w-4 h-4 mr-2" />
                      Start
                    </Button>
                    <Button 
                      variant="secondary"
                      onClick={() => handleAction('stop')} 
                      disabled={!canStop || actionLoading === 'stop'}
                      loading={actionLoading === 'stop'}
                    >
                      <Square className="w-4 h-4 mr-2" />
                      Stop
                    </Button>
                    <Button 
                      variant="secondary"
                      onClick={() => handleAction('resume')} 
                      disabled={!canResume || actionLoading === 'resume'}
                      loading={actionLoading === 'resume'}
                    >
                      <RotateCcw className="w-4 h-4 mr-2" />
                      Resume
                    </Button>
                    <Button 
                      onClick={() => router.push(`/ide/${env.id}`)}
                      disabled={!isRunning}
                      className={!isRunning ? 'opacity-50 cursor-not-allowed' : ''}
                    >
                      <Terminal className="w-4 h-4 mr-2" />
                      Open IDE
                    </Button>
                    {isOwner && (
                      <Button 
                        variant="danger"
                        onClick={() => handleAction('delete')} 
                        disabled={actionLoading === 'delete'}
                        loading={actionLoading === 'delete'}
                      >
                        <Trash2 className="w-4 h-4 mr-2" />
                        Delete
                      </Button>
                    )}
                  </div>
                  {!isRunning && (
                    <p className="mt-3 text-sm text-slate-500">
                      <Activity className="w-4 h-4 inline mr-1" />
                      Environment is {env.status === 'BUILDING' ? 'building' : env.status === 'CREATED' ? 'starting' : env.status.toLowerCase()}. IDE will be available when running.
                    </p>
                  )}
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="flex flex-row items-center justify-between">
                <CardTitle>Members ({members.length})</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  {members.map((member) => (
                    <div key={member.user.id} className="flex items-center justify-between py-3 border-b border-slate-100 last:border-0">
                      <div className="flex items-center gap-3">
                        <Avatar src={member.user.avatar_url} name={member.user.username} size="sm" />
                        <div>
                          <p className="font-medium text-slate-900">{member.user.username}</p>
                          <p className="text-sm text-slate-500">Joined {new Date(member.joined_at).toLocaleDateString()}</p>
                        </div>
                      </div>
                      <Badge variant={member.role === 'OWNER' ? 'info' : member.role === 'COLLABORATOR' ? 'success' : 'neutral'}>
                        {member.role}
                      </Badge>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          </div>

          {/* Build Log Panel */}
          <Card>
            <CardHeader className="flex flex-row items-center justify-between">
              <CardTitle className="flex items-center gap-2">
                <Terminal className="w-5 h-5" />
                Build Log
                {env.status === 'BUILDING' && (
                  <span className="text-amber-600 text-sm font-normal">● Building</span>
                )}
                {env.status === 'RUNNING' && (
                  <span className="text-green-600 text-sm font-normal">● Ready</span>
                )}
                {env.status === 'BUILD_FAILED' && (
                  <span className="text-red-600 text-sm font-normal">● Failed</span>
                )}
              </CardTitle>
            </CardHeader>
            <CardContent>
              <BuildLogPanel environmentId={env.id} status={env.status} />
            </CardContent>
          </Card>

          <div className="space-y-6">
            <Card>
              <CardHeader className="flex flex-row items-center justify-between">
                <CardTitle>Share Environment</CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <p className="text-slate-600">
                  Generate a shareable link to invite collaborators or viewers to this environment.
                </p>
                
                <div className="space-y-4">
                  <div>
                    <label className="block text-sm font-medium text-slate-700 mb-2">Role</label>
                    <select
                      value={shareForm.role}
                      onChange={(e) => setShareForm({ ...shareForm, role: e.target.value as 'COLLABORATOR' | 'VIEWER' })}
                      className="w-full px-4 py-2.5 border border-slate-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-harbor-500"
                    >
                      <option value="VIEWER">Viewer (read-only)</option>
                      <option value="COLLABORATOR">Collaborator (can edit)</option>
                    </select>
                  </div>
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <label className="block text-sm font-medium text-slate-700 mb-2">Expires in (hours)</label>
                      <input
                        type="number"
                        min="1"
                        max="168"
                        value={shareForm.expires_in_hours}
                        onChange={(e) => setShareForm({ ...shareForm, expires_in_hours: parseInt(e.target.value) })}
                        className="w-full px-4 py-2.5 border border-slate-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-harbor-500"
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-slate-700 mb-2">Max Uses</label>
                      <input
                        type="number"
                        min="1"
                        max="100"
                        value={shareForm.max_uses}
                        onChange={(e) => setShareForm({ ...shareForm, max_uses: parseInt(e.target.value) })}
                        className="w-full px-4 py-2.5 border border-slate-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-harbor-500"
                      />
                    </div>
                  </div>
                  
                  <Button onClick={handleCreateShareLink} className="w-full">
                    Create Share Link
                  </Button>
                </div>
                
                {shareLink && (
                  <div className="p-4 bg-harbor-50 border border-harbor-200 rounded-lg">
                    <p className="text-sm font-medium text-harbor-900 mb-2">Share Link Created</p>
                    <div className="flex gap-2">
                      <input
                        type="text"
                        value={shareLink}
                        readOnly
                        className="flex-1 px-3 py-2 bg-white border border-harbor-200 rounded-lg text-sm font-mono"
                      />
                      <Button 
                        variant="secondary" 
                        size="sm"
                        onClick={() => {
                          navigator.clipboard.writeText(shareLink);
                          success('Copied to clipboard');
                        }}
                      >
                        Copy
                      </Button>
                    </div>
                  </div>
                )}
              </CardContent>
            </Card>
          </div>

          <div className="lg:col-span-3">
            <Card>
              <CardHeader className="flex flex-row items-center justify-between">
                <CardTitle className="flex items-center gap-2">
                  <Terminal className="w-5 h-5" />
                  Build Log
                  {env?.status === 'BUILDING' && (
                    <span className="text-amber-600 text-sm font-normal">● Building</span>
                  )}
                  {env?.status === 'RUNNING' && (
                    <span className="text-green-600 text-sm font-normal">● Ready</span>
                  )}
                  {env?.status === 'BUILD_FAILED' && (
                    <span className="text-red-600 text-sm font-normal">● Failed</span>
                  )}
                </CardTitle>
              </CardHeader>
              <CardContent>
                <BuildLogPanel environmentId={env.id} status={env?.status} />
              </CardContent>
            </Card>
          </div>
}
}

// ============ BuildLogPanel Component ============
function BuildLogPanel({ environmentId, status }: { environmentId: string; status?: string }) {
  const [logs, setLogs] = useState<Array<{ id: string; sequence: number; message: string; level: string; created_at: string }>>([]);
  const [loading, setLoading] = useState(true);
  const wsRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    // Fetch historical logs
    fetch(`/api/v1/environments/${environmentId}/logs?limit=200`)
      .then(res => res.json())
      .then(data => {
        if (data.logs) {
          setLogs(data.logs);
        }
      })
      .catch(console.error)
      .finally(() => setLoading(false));

    // Connect to WebSocket for real-time logs
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const ws = new WebSocket(`${protocol}//${window.location.host}/api/v1/environments/${environmentId}/build-logs/ws`);
    wsRef.current = ws;

    ws.onmessage = (event) => {
      try {
        const log = JSON.parse(event.data);
        if (log.type === 'log' && log.message) {
          setLogs(prev => [...prev, log]);
        }
      } catch (err) {
        console.error('Failed to parse log message:', err);
      }
    };

    ws.onerror = () => {
      console.error('WebSocket error');
    };

    return () => {
      ws.close();
    };
  }, [environmentId]);

  const getLevelColor = (level: string) => {
    switch (level) {
      case 'success': return 'text-green-600';
      case 'error': return 'text-red-600';
      case 'warn': return 'text-amber-600';
      default: return 'text-slate-600';
    }
  };

  const getStatusIndicator = () => {
    if (status === 'BUILDING') return <span className="text-amber-600">● Building...</span>;
    if (status === 'RUNNING') return <span className="text-green-600">● Ready</span>;
    if (status === 'BUILD_FAILED') return <span className="text-red-600">● Failed</span>;
    return <span className="text-slate-400">● Waiting</span>;
  };

  if (loading) {
    return <div className="h-64 flex items-center justify-center text-slate-400">Loading build logs...</div>;
  }

  return (
    <div className="h-64 overflow-y-auto bg-slate-900 rounded-lg p-4 font-mono text-sm text-slate-100">
      <div className="flex items-center gap-2 mb-3 pb-2 border-b border-slate-700">
        <span className="text-xs text-slate-400">Status:</span>
        {getStatusIndicator()}
        <span className="ml-auto text-xs text-slate-500">{logs.length} lines</span>
      </div>
      <div className="space-y-1">
        {logs.length === 0 ? (
          <div className="flex items-center justify-center h-full text-slate-500">
            No build logs yet
          </div>
        ) : (
          logs.map((log) => (
            <div
              key={log.id}
              className={`flex gap-2 ${getLevelColor(log.level)}`}
            >
              <span className="text-slate-500 w-20 shrink-0">
                {new Date(log.created_at).toLocaleTimeString()}
              </span>
              <span className="text-xs font-mono w-6 shrink-0">
                {log.sequence.toString().padStart(4, '0')}
              </span>
              <span className="flex-1 break-all whitespace-pre-wrap">{log.message}</span>
            </div>
          ))}
      </div>
    </div>
  );
}

function DetailRow({ label, value, copyable, avatar }: { label: string; value: string; copyable?: boolean; avatar?: string }) {
  return (
    <div>
      <label className="block text-sm font-medium text-slate-500 mb-1">{label}</label>
      <div className="flex items-center gap-2">
        {avatar && <Avatar src={avatar} name={value} size="sm" />}
        <span className="font-mono text-sm text-slate-700 truncate flex-1">{value}</span>
        {copyable && (
          <button className="p-1.5 text-slate-400 hover:text-slate-600 rounded transition-colors" onClick={() => navigator.clipboard.writeText(value)}>
            <span className="sr-only">Copy</span>
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
            </svg>
          </button>
        )}
      </div>
    </div>
  );
}}
