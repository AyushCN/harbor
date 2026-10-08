'use client';

import { useEffect, useState, useMemo } from 'react';
import { useRouter } from 'next/navigation';
import { useAuthStore, useEnvironmentStore, useUIStore } from '@/lib/store';
import { environmentsApi, type Environment } from '@/lib/api';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/Button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card';
import { Input } from '@/components/ui/Input';
import { Badge, StatusBadge } from '@/components/ui/Badge';
import { Avatar, AvatarGroup } from '@/components/ui/Avatar';
import { Dropdown } from '@/components/ui/Dropdown';
import { Modal } from '@/components/ui/Modal';
import { useToast } from '@/components/ui/Toast';
import { 
  Plus, 
  Github, 
  Search, 
  Filter, 
  Menu, 
  X, 
  MoreVertical,
  Settings,
  Users,
  Play,
  Square,
  RotateCcw,
  Trash2,
  Share2,
  ExternalLink,
  Loader2,
  ChevronDown,
  Terminal,
  Clock,
  Activity,
  Users as UsersIcon,
  Server,
  Archive,
  Bell,
  ChevronRight,
  Copy,
  Globe,
  Shield,
  Star,
  Bell as BellIcon,
  Moon,
  Sun,
  LogOut,
  User,
  Mail,
  AlertTriangle,
  RefreshCw,
  ArrowUpDown,
  LayoutDashboard,
  FolderGit2,
  Share2 as ShareIcon,
  UserCheck,
  UserX,
  Archive as ArchiveIcon,
  Zap,
  Sparkles,
} from 'lucide-react';

type TabType = 'all' | 'hosted' | 'shared' | 'archived' | 'failed' | 'building';
type SortType = 'newest' | 'active' | 'name';
type RoleFilterType = 'all' | 'OWNER' | 'COLLABORATOR' | 'VIEWER';

export default function DashboardPage() {
  const router = useRouter();
  const { user, token, setAuth, logout } = useAuthStore();
  const { environments, setEnvironments, addEnvironment, updateEnvironment, removeEnvironment, setCurrentEnvironment } = useEnvironmentStore();
  const { sidebarOpen, toggleSidebar } = useUIStore();
  const { success, error } = useToast();
  
  // State
  const [searchQuery, setSearchQuery] = useState('');
  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [roleFilter, setRoleFilter] = useState<RoleFilterType>('all');
  const [sortBy, setSortBy] = useState<SortType>('newest');
  const [activeTab, setActiveTab] = useState<TabType>('all');
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [newEnvForm, setNewEnvForm] = useState({ git_url: '', git_branch: 'main', name: '', visibility: 'private' });
  const [creating, setCreating] = useState(false);
  const [loading, setLoading] = useState(true);
  const [dropdownId, setDropdownId] = useState<string | null>(null);
  const [userDropdownOpen, setUserDropdownOpen] = useState(false);
  const [notificationsOpen, setNotificationsOpen] = useState(false);

  useEffect(() => {
    if (!token) {
      fetchUserFromApi();
      return;
    }
    fetchEnvironments();
  }, [token, router]);

  const fetchUserFromApi = async () => {
    try {
      const response = await fetch('/api/v1/auth/me', { credentials: 'include' });
      if (response.ok) {
        const userData = await response.json();
        setAuth(userData, 'cookie');
        fetchEnvironments();
      } else {
        router.push('/login');
      }
    } catch {
      router.push('/login');
    }
  };

  const fetchEnvironments = async () => {
    try {
      const response = await environmentsApi.list();
      setEnvironments(response.data);
    } catch (err) {
      console.error('Failed to fetch environments:', err);
      error('Failed to load environments');
    } finally {
      setLoading(false);
    }
  };

  const handleCreateEnvironment = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newEnvForm.git_url.trim()) return;
    
    setCreating(true);
    try {
      const response = await environmentsApi.create(newEnvForm);
      success('Environment created', 'Your environment is being built');
      setShowCreateModal(false);
      setNewEnvForm({ git_url: '', git_branch: 'main', name: '', visibility: 'private' });
      fetchEnvironments();
    } catch (err: any) {
      error('Failed to create environment', err.response?.data?.error?.message || 'Unknown error');
    } finally {
      setCreating(false);
    }
  };


  // Stats calculation
  const stats = useMemo(() => {
    const envs = environments || [];
    return {
      total: envs.length,
      running: envs.filter(e => e.status === 'RUNNING').length,
      failed: envs.filter(e => e.status === 'BUILD_FAILED').length,
      building: envs.filter(e => e.status === 'BUILDING' || e.status === 'CREATED').length,
      hosted: envs.filter(e => e.role === 'OWNER').length,
      shared: envs.filter(e => e.role !== 'OWNER').length,
    };
  }, [environments]);

  // Filter and sort environments
  const filteredEnvironments = useMemo(() => {
    const envs = environments || [];
    let result = envs.filter((env) => {
      // Tab filter
      if (activeTab === 'hosted' && env.role !== 'OWNER') return false;
      if (activeTab === 'shared' && env.role === 'OWNER') return false;
      if (activeTab === 'archived' && !['STOPPED', 'SUSPENDED', 'CRASHED', 'BUILD_FAILED'].includes(env.status)) return false;
      
      // Status filter
      if (statusFilter !== 'all' && env.status !== statusFilter) return false;
      if (activeTab === 'failed' && env.status !== 'BUILD_FAILED') return false;
      if (activeTab === 'building' && !['BUILDING', 'CREATED'].includes(env.status)) return false;
      
      // Role filter
      if (roleFilter !== 'all' && env.role !== roleFilter) return false;
      
      // Search
      const matchesSearch = env.name?.toLowerCase().includes(searchQuery.toLowerCase()) ||
        env.workspace?.git_url.toLowerCase().includes(searchQuery.toLowerCase());
      if (!matchesSearch) return false;
      
      return true;
    });

    // Sort
    result.sort((a, b) => {
      switch (sortBy) {
        case 'newest':
          return new Date(b.created_at).getTime() - new Date(a.created_at).getTime();
        case 'active':
          return new Date((b as any).updated_at || b.created_at).getTime() - new Date((a as any).updated_at || a.created_at).getTime();
        case 'name':
          return (a.name || '').localeCompare(b.name || '');
        default:
          return 0;
      }
    });

    return result;
  }, [environments, activeTab, statusFilter, roleFilter, searchQuery, sortBy]);


  const handleAction = async (envId: string, action: 'start' | 'stop' | 'resume' | 'delete' | 'retry') => {
    try {
      if (action === 'start') await environmentsApi.start(envId);
      else if (action === 'stop') await environmentsApi.stop(envId);
      else if (action === 'resume') await environmentsApi.resume(envId);
      else if (action === 'delete') await environmentsApi.delete(envId);
      else if (action === 'retry') await environmentsApi.start(envId); // Retry uses start endpoint
      
      success(`${action.charAt(0).toUpperCase() + action.slice(1)} requested`, `Environment ${action} has been queued`);
      fetchEnvironments();
    } catch (err: any) {
      error(`Failed to ${action} environment`, err.response?.data?.error?.message || 'Unknown error');
    }
  };

  const getActionItems = (env: Environment) => {
    const isOwner = env.role === 'OWNER';
    const canStart = ['CREATED', 'STOPPED', 'SUSPENDED'].includes(env.status) && env.role !== 'VIEWER';
    const canStop = env.status === 'RUNNING' && isOwner;
    const canResume = env.status === 'SUSPENDED' && env.role !== 'VIEWER';
    const canDelete = isOwner;
    const isRunning = env.status === 'RUNNING';
    const isFailed = env.status === 'BUILD_FAILED';
    const isBuilding = ['CREATED', 'BUILDING'].includes(env.status);

    return [
      // Always show "Open Environment" for non-running environments
      ...(!isRunning ? [{
        label: 'Open Environment',
        onClick: () => router.push(`/environments/${env.id}`),
        icon: <ExternalLink className="w-4 h-4" />
      }] : []),
      
      ...(isRunning ? [
        { label: 'Open IDE', onClick: () => router.push(`/ide/${env.id}`), icon: <Terminal className="w-4 h-4" /> },
        { label: 'Open Preview', onClick: () => env.public_url && window.open(env.public_url, '_blank'), icon: <ExternalLink className="w-4 h-4" />, disabled: !env.public_url },
      ] : []),
      ...(canStart ? [{ label: 'Start', onClick: () => handleAction(env.id, 'start'), icon: <Play className="w-4 h-4" /> }] : []),
      ...(canStop ? [{ label: 'Stop', onClick: () => handleAction(env.id, 'stop'), icon: <Square className="w-4 h-4" /> }] : []),
      ...(canResume ? [{ label: 'Resume', onClick: () => handleAction(env.id, 'resume'), icon: <RotateCcw className="w-4 h-4" /> }] : []),
      // Failed build actions
      ...(env.status === 'BUILD_FAILED' ? [
        { label: 'View Logs', onClick: () => router.push(`/environments/${env.id}?tab=logs`), icon: <Activity className="w-4 h-4" /> },
        { label: 'Retry Build', onClick: () => handleAction(env.id, 'retry'), icon: <RotateCcw className="w-4 h-4" /> },
      ] : []),
      { label: 'Members', onClick: () => router.push(`/environments/${env.id}/members`), icon: <Users className="w-4 h-4" /> },
      ...(isRunning ? [{ label: 'Share', onClick: () => router.push(`/environments/${env.id}/share`), icon: <ShareIcon className="w-4 h-4" /> }] : []),
      ...(canDelete ? [{ label: 'Delete', onClick: () => handleAction(env.id, 'delete'), icon: <Trash2 className="w-4 h-4" />, danger: true }] : []),
    ];
  };

  const copyUrl = (url: string) => {
    navigator.clipboard.writeText(url);
    success('Copied', 'Preview URL copied to clipboard');
  };

  const formatRelativeTime = (date: string) => {
    const now = new Date();
    const then = new Date(date);
    const diffMs = now.getTime() - then.getTime();
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMs / 3600000);
    const diffDays = Math.floor(diffMs / 86400000);

    if (diffMins < 1) return 'just now';
    if (diffMins < 60) return `${diffMins}m ago`;
    if (diffHours < 24) return `${diffHours}h ago`;
    if (diffDays < 7) return `${diffDays}d ago`;
    return new Date(date).toLocaleDateString();
  };

  if (!token) {
    return <div className="flex items-center justify-center min-h-screen"><Loader2 className="w-8 h-8 animate-spin text-harbor-600" /></div>;
  }

  return (
    <div className="min-h-screen bg-slate-50">
      {/* Top Navigation Bar */}
      <header className="bg-white border-b border-slate-200 sticky top-0 z-50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-16">
            {/* Left: Logo + Search */}
            <div className="flex items-center gap-4 flex-1">
              <button onClick={toggleSidebar} className="lg:hidden p-2 rounded-lg hover:bg-slate-100">
                <Menu className="w-6 h-6" />
              </button>
              <div className="flex items-center gap-2">
                <div className="w-8 h-8 bg-harbor-600 rounded-lg flex items-center justify-center">
                  <Terminal className="w-5 h-5 text-white" />
                </div>
                <span className="text-xl font-bold text-slate-900">Harbor</span>
              </div>
              
              {/* Search Bar */}
              <div className="hidden md:block relative w-80">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400" />
                <Input
                  type="search"
                  placeholder="Search environments..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="w-full pl-10"
                />
              </div>
            </div>

            {/* Right: Create + Notifications + User */}
            <div className="flex items-center gap-3">
              {/* Notifications */}
              <div className="relative">
                <button 
                  onClick={() => setNotificationsOpen(!notificationsOpen)}
                  className="relative p-2 rounded-lg hover:bg-slate-100 text-slate-600 hover:text-slate-900"
                >
                  <BellIcon className="w-5 h-5" />
                  <span className="absolute top-1 right-1 w-2 h-2 bg-red-500 rounded-full"></span>
                </button>
                {notificationsOpen && (
                  <div className="absolute right-0 mt-2 w-80 bg-white border border-slate-200 rounded-lg shadow-lg py-2 z-50">
                    <div className="px-4 py-3 border-b border-slate-200">
                      <h4 className="font-semibold text-slate-900">Notifications</h4>
                    </div>
                    <div className="max-h-64 overflow-y-auto">
                      <div className="px-4 py-3 text-slate-500 text-sm text-center">No new notifications</div>
                    </div>
                    <div className="px-4 py-2 border-t border-slate-200 text-center">
                      <button className="text-sm text-harbor-600 hover:text-harbor-700">View all</button>
                    </div>
                  </div>
                )}
              </div>

              {/* Create Button */}
              <Button onClick={() => setShowCreateModal(true)} className="hidden sm:flex">
                <Plus className="w-4 h-4 mr-2" />
                Create Environment
              </Button>

              {/* Mobile Create Button */}
              <Button variant="ghost" size="sm" onClick={() => setShowCreateModal(true)} className="md:hidden">
                <Plus className="w-5 h-5" />
              </Button>

              {/* User Avatar Dropdown */}
              <div className="relative">
                <button 
                  onClick={() => setUserDropdownOpen(!userDropdownOpen)}
                  className="flex items-center gap-2 p-1.5 rounded-lg hover:bg-slate-100"
                >
                  <Avatar 
                    src={user?.avatar_url} 
                    name={user?.name || user?.username} 
                    size="sm" 
                    status="online" 
                  />
                  <ChevronDown className="w-4 h-4 text-slate-500" />
                </button>
                {userDropdownOpen && (
                  <div className="absolute right-0 mt-2 w-48 bg-white border border-slate-200 rounded-lg shadow-lg py-1 z-50">
                    <div className="px-4 py-2 border-b border-slate-100">
                      <p className="font-medium text-slate-900 truncate">{user?.name || user?.username}</p>
                      <p className="text-sm text-slate-500 truncate">{user?.username}</p>
                    </div>
                    <a href="/settings" className="flex items-center gap-2 px-4 py-2 text-slate-700 hover:bg-slate-50">
                      <User className="w-4 h-4" />
                      Profile
                    </a>
                    <a href="/settings" className="flex items-center gap-2 px-4 py-2 text-slate-700 hover:bg-slate-50">
                      <Settings className="w-4 h-4" />
                      Settings
                    </a>
                    <hr className="my-1 border-slate-100" />
                    <button 
                      onClick={() => {
                        logout();
                        localStorage.removeItem('harbor_token');
                        router.push('/login');
                      }}
                      className="flex items-center gap-2 w-full px-4 py-2 text-red-600 hover:bg-red-50"
                    >
                      <LogOut className="w-4 h-4" />
                      Sign Out
                    </button>
                  </div>
                )}
              </div>
            </div>
          </div>
        </div>
      </header>

      <div className="flex">
        {/* Mobile Sidebar Overlay */}
        {sidebarOpen && (
          <div 
            className="fixed inset-0 z-40 bg-black/50 lg:hidden" 
            onClick={() => toggleSidebar()}
            aria-hidden="true"
          />
        )}

        {/* Sidebar */}
        <aside className={cn(
          'fixed inset-y-0 left-0 z-50 w-64 bg-white border-r border-slate-200 transform transition-transform duration-300 lg:relative lg:translate-x-0',
          sidebarOpen ? 'translate-x-0' : '-translate-x-full'
        )}>
          <nav className="flex flex-col h-full p-4 space-y-1">
            <button
              onClick={() => router.push('/dashboard')}
              className="sidebar-item sidebar-item-active"
            >
              <LayoutDashboard className="w-5 h-5" />
              <span>Dashboard</span>
            </button>
            <button
              onClick={() => router.push('/environments')}
              className="sidebar-item"
            >
              <FolderGit2 className="w-5 h-5" />
              <span>Environments</span>
            </button>
            <hr className="my-4 border-slate-200" />
            <div className="px-3 py-2 text-xs font-semibold text-slate-500 uppercase tracking-wider">
              Account
            </div>
            <a href="/settings" className="sidebar-item">
              <Settings className="w-5 h-5" />
              <span>Settings</span>
            </a>
            <button
              onClick={() => {
                logout();
                localStorage.removeItem('harbor_token');
                router.push('/login');
              }}
              className="sidebar-item text-red-600 hover:bg-red-50"
            >
              <LogOut className="w-5 h-5" />
              <span>Sign Out</span>
            </button>
          </nav>
        </aside>

        {sidebarOpen && (
          <div 
            className="fixed inset-0 z-40 bg-black/50 lg:hidden" 
            onClick={() => toggleSidebar()}
            aria-hidden="true"
          />
        )}

        {/* Main Content */}
        <main className="flex-1 lg:ml-64 p-6">
          {/* Stats Cards */}
          <div className="grid grid-cols-2 md:grid-cols-5 gap-4 mb-6">
            <StatCard 
              title="Total Environments" 
              value={stats.total} 
              icon={<Server className="w-5 h-5" />}
              color="bg-harbor-500"
              onClick={() => { setActiveTab('all'); setStatusFilter('all'); }}
            />
            <StatCard 
              title="Running" 
              value={stats.running} 
              icon={<Activity className="w-5 h-5" />}
              color="bg-emerald-500"
              onClick={() => { setActiveTab('all'); setStatusFilter('RUNNING'); }}
            />
            <StatCard 
              title="Failed" 
              value={stats.failed} 
              icon={<AlertTriangle className="w-5 h-5" />}
              color="bg-red-500"
              onClick={() => { setActiveTab('all'); setStatusFilter('BUILD_FAILED'); }}
            />
            <StatCard 
              title="Building" 
              value={stats.building} 
              icon={<Zap className="w-5 h-5" />}
              color="bg-amber-500"
              onClick={() => { setActiveTab('all'); setStatusFilter('BUILDING'); }}
            />
            <StatCard 
              title="Hosted by Me" 
              value={stats.hosted} 
              icon={<Shield className="w-5 h-5" />}
              color="bg-purple-500"
              onClick={() => { setActiveTab('hosted'); }}
            />
            <StatCard 
              title="Shared with Me" 
              value={stats.shared} 
              icon={<UsersIcon className="w-5 h-5" />}
              color="bg-amber-500"
              onClick={() => { setActiveTab('shared'); }}
            />
          </div>

          {/* Tabs */}
          <div className="flex gap-1 mb-4 border-b border-slate-200">
            {([
              { id: 'all', label: 'All', icon: <LayoutDashboard className="w-4 h-4" /> },
              { id: 'hosted', label: 'Hosted by Me', icon: <Shield className="w-4 h-4" /> },
              { id: 'shared', label: 'Shared with Me', icon: <UsersIcon className="w-4 h-4" /> },
              { id: 'archived', label: 'Archived', icon: <Archive className="w-4 h-4" /> },
              { id: 'failed', label: 'Failed', icon: <AlertTriangle className="w-4 h-4" /> },
              { id: 'building', label: 'Building', icon: <Zap className="w-4 h-4" /> },
            ] as const).map((tab) => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id as TabType)}
                className={cn(
                  'flex items-center gap-2 px-4 py-2 text-sm font-medium border-b-2 -mb-px transition-colors',
                  activeTab === tab.id
                    ? 'text-harbor-600 border-harbor-600'
                    : 'text-slate-500 hover:text-slate-700 hover:border-slate-300'
                )}
              >
                {tab.icon}
                <span>{tab.label}</span>
              </button>
            ))}
          </div>

          {/* Filters */}
          <div className="flex flex-wrap items-center gap-4 mb-6 p-4 bg-white rounded-lg border border-slate-200">
            <div className="flex-1 min-w-[200px] relative">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-slate-400" />
              <Input
                type="search"
                placeholder="Search environments..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="pl-10 w-full"
              />
            </div>

            <select
              value={statusFilter}
              onChange={(e) => setStatusFilter(e.target.value)}
              className="px-3 py-2 text-sm border border-slate-300 rounded-lg bg-white focus:outline-none focus:ring-2 focus:ring-harbor-500"
            >
              <option value="all">All Status</option>
              <option value="RUNNING">Running</option>
              <option value="STOPPED">Stopped</option>
              <option value="SUSPENDED">Suspended</option>
              <option value="CREATED">Created</option>
              <option value="BUILDING">Building</option>
              <option value="BUILD_FAILED">Build Failed</option>
            </select>

            <select
              value={roleFilter}
              onChange={(e) => setRoleFilter(e.target.value as RoleFilterType)}
              className="px-3 py-2 text-sm border border-slate-300 rounded-lg bg-white focus:outline-none focus:ring-2 focus:ring-harbor-500"
            >
              <option value="all">All Roles</option>
              <option value="OWNER">Owner</option>
              <option value="COLLABORATOR">Collaborator</option>
              <option value="VIEWER">Viewer</option>
            </select>

            <select
              value={sortBy}
              onChange={(e) => setSortBy(e.target.value as SortType)}
              className="px-3 py-2 text-sm border border-slate-300 rounded-lg bg-white focus:outline-none focus:ring-2 focus:ring-harbor-500"
            >
              <option value="newest">Newest First</option>
              <option value="active">Last Active</option>
              <option value="name">Name A-Z</option>
            </select>

            <Button variant="ghost" size="sm" onClick={() => setShowCreateModal(true)}>
              <Plus className="w-4 h-4 mr-2" />
              New Environment
            </Button>
          </div>

          {/* Environments Grid */}
          {loading ? (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              {[1, 2, 3].map((i) => (
                <Card key={i} className="animate-pulse">
                  <CardContent className="pt-6">
                    <div className="h-4 bg-slate-200 rounded w-3/4 mb-4"></div>
                    <div className="h-4 bg-slate-200 rounded w-1/2 mb-2"></div>
                    <div className="h-4 bg-slate-200 rounded w-1/4"></div>
                  </CardContent>
                </Card>
              ))}
            </div>
          ) : filteredEnvironments.length === 0 ? (
            <Card className="text-center py-12">
              <CardContent>
                <Sparkles className="w-16 h-16 text-slate-300 mx-auto mb-4" />
                <h3 className="text-lg font-medium text-slate-900 mb-2">No environments found</h3>
                <p className="text-slate-500 mb-6">{activeTab !== 'all' ? `No ${activeTab} environments` : 'Create your first development environment'}</p>
                <Button onClick={() => setShowCreateModal(true)}>
                  <Plus className="w-4 h-4 mr-2" />
                  Create Environment
                </Button>
              </CardContent>
            </Card>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              {filteredEnvironments.map((env) => (
                <EnvironmentCard 
                  key={env.id} 
                  env={env} 
                  onAction={(action) => handleAction(env.id, action)}
                  dropdownItems={getActionItems(env)}
                  dropdownId={env.id}
                  setDropdownId={setDropdownId}
                  dropdownIdState={dropdownId}
                />
              ))}
            </div>
          )}
        </main>
      </div>

      {/* Create Environment Modal */}
      <CreateEnvironmentModal
        isOpen={showCreateModal}
        onClose={() => setShowCreateModal(false)}
        onSubmit={handleCreateEnvironment}
        formData={newEnvForm}
        setFormData={setNewEnvForm}
        loading={creating}
      />
    </div>
  );
}

// ============ StatCard Component ============
function StatCard({ title, value, icon, color, onClick }: { title: string; value: number; icon: React.ReactNode; color: string; onClick?: () => void }) {
  return (
    <Card className={cn(onClick && 'cursor-pointer hover:shadow-md transition-shadow')} onClick={onClick}>
      <CardContent className="p-6">
        <div className="flex items-center justify-between">
          <div>
            <p className="text-sm text-slate-500 mb-1">{title}</p>
            <p className="text-3xl font-bold text-slate-900">{value}</p>
          </div>
          <div className={cn('p-3 rounded-xl', color)}>
            {icon}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

// ============ EnvironmentCard Component ============
function EnvironmentCard({ 
  env, 
  onAction, 
  dropdownItems, 
  dropdownId, 
  setDropdownId, 
  dropdownIdState 
}: { 
  env: Environment; 
  onAction: (action: 'start' | 'stop' | 'resume' | 'delete') => void;
  dropdownItems: any[];
  dropdownId: string;
  setDropdownId: (id: string | null) => void;
  dropdownIdState: string | null;
}) {
  const isOpen = dropdownIdState === dropdownId;
  const isOwner = env.role === 'OWNER';
  const canStart = ['CREATED', 'STOPPED', 'SUSPENDED'].includes(env.status) && env.role !== 'VIEWER';

  return (
    <Card className="h-full hover:shadow-lg transition-shadow group">
      <CardContent className="pt-6">
        {/* Header */}
        <div className="flex items-start justify-between mb-4">
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-3 mb-2 flex-wrap">
              <h3 className="text-lg font-semibold text-slate-900 truncate">
                {env.name || env.workspace?.name || 'Unnamed'}
              </h3>
              <StatusBadge status={env.status} />
              <Badge 
                variant={env.role === 'OWNER' ? 'info' : env.role === 'COLLABORATOR' ? 'success' : 'neutral'} 
                size="sm"
              >
                {env.role}
              </Badge>
            </div>
            <p className="text-sm text-slate-500 truncate mb-2">
              {env.workspace?.git_url}
            </p>
            <div className="flex items-center gap-2 text-sm text-slate-500 flex-wrap">
              <span>Branch: {env.workspace?.git_branch || 'main'}</span>
              <span className="text-slate-300">•</span>
              <span>Created {formatRelativeTime(env.created_at)}</span>
              <span className="text-slate-300">•</span>
              <span className="flex items-center gap-1">
                <Activity className="w-3 h-3" />
                <span>Updated {formatRelativeTime((env as any).updated_at || env.created_at)}</span>
              </span>
            </div>
          </div>
          <Dropdown
            trigger={<button className="p-2 rounded-lg hover:bg-slate-100 text-slate-400 hover:text-slate-600"><MoreVertical className="w-5 h-5" /></button>}
            items={dropdownItems}
          />
        </div>

        {/* Footer */}
        <div className="flex items-center justify-between pt-4 border-t border-slate-100">
          <div className="flex items-center gap-3">
            <AvatarGroup 
              avatars={[
                { src: env.host.avatar_url, name: env.host.username },
              ]} 
              max={5} 
              size="sm" 
            />
            <div>
              <p className="text-xs text-slate-500">Host</p>
              <p className="text-sm font-medium text-slate-900 truncate max-w-[150px]">{env.host.username}</p>
            </div>
            {env.status === 'RUNNING' && (
              <div className="flex items-center gap-1 px-2 py-1 bg-emerald-50 text-emerald-700 rounded-full text-xs">
                <Activity className="w-3 h-3 animate-pulse" />
                <span>Live</span>
              </div>
            )}
          </div>
          <div className="flex items-center gap-2">
            {env.public_url && (
              <button 
                onClick={() => navigator.clipboard.writeText(env.public_url!)}
                className="text-sm text-harbor-600 hover:text-harbor-700 font-medium flex items-center gap-1 px-2 py-1 rounded-lg hover:bg-harbor-50"
              >
                <Globe className="w-3 h-3" />
                <span>Preview URL</span>
                <Copy className="w-3 h-3" />
              </button>
            )}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

// ============ CreateEnvironmentModal Component ============
function CreateEnvironmentModal({ 
  isOpen, 
  onClose, 
  onSubmit, 
  formData, 
  setFormData, 
  loading 
}: { 
  isOpen: boolean; 
  onClose: () => void; 
  onSubmit: (e: React.FormEvent) => void;
  formData: { git_url: string; git_branch: string; name: string; visibility: string };
  setFormData: React.Dispatch<React.SetStateAction<{ git_url: string; git_branch: string; name: string; visibility: string }>>;
  loading: boolean;
}) {
  const [branches, setBranches] = useState<string[]>([]);
  const [branchesLoading, setBranchesLoading] = useState(false);
  const [defaultBranch, setDefaultBranch] = useState<string>('main');
  const [urlError, setUrlError] = useState<string | null>(null);

  // Fetch branches when git_url changes
  useEffect(() => {
    if (!formData.git_url.trim()) {
      setBranches([]);
      setDefaultBranch('main');
      setUrlError(null);
      return;
    }

    // Validate GitHub URL format
    const githubUrlPattern = /^(https?:\/\/)?(www\.)?github\.com\/[\w-]+\/[\w\-.]+(\.git)?\/?$/;
    if (!githubUrlPattern.test(formData.git_url)) {
      setUrlError('Please enter a valid GitHub repository URL (e.g., https://github.com/owner/repo)');
      setBranches([]);
      return;
    }
    setUrlError(null);

    // Fetch branches
    const fetchBranches = async () => {
      setBranchesLoading(true);
      try {
        const response = await fetch('/api/v1/environments/branches', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          credentials: 'include',
          body: JSON.stringify({ git_url: formData.git_url }),
        });
        if (response.ok) {
          const data = await response.json();
          const branchNames = data.branches.map((b: any) => b.name);
          setBranches(branchNames);
          setDefaultBranch(data.default_branch || 'main');
          
          // Auto-select default branch if current branch is not in the list
          if (!branchNames.includes(formData.git_branch)) {
            setFormData(prev => ({ ...prev, git_branch: data.default_branch || 'main' }));
          }
        } else {
          // Fallback to default branches
          setBranches(['main', 'master', 'develop', 'dev']);
        }
      } catch (err) {
        console.error('Failed to fetch branches:', err);
        setBranches(['main', 'master', 'develop', 'dev']);
      } finally {
        setBranchesLoading(false);
      }
    };

    const debounceTimer = setTimeout(fetchBranches, 500);
    return () => clearTimeout(debounceTimer);
  }, [formData.git_url]);

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title="Create New Environment"
      description="Enter a GitHub repository URL to create a new development environment"
      size="lg"
    >
      <form onSubmit={onSubmit}>
        <div className="space-y-4">
          <Input
            label="GitHub Repository URL"
            placeholder="https://github.com/owner/repo"
            value={formData.git_url}
            onChange={(e) => { setFormData({ ...formData, git_url: e.target.value }); setUrlError(null); }}
            required
          />
          {urlError && <p className="text-sm text-red-600">{urlError}</p>}
          
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-slate-700 mb-2">Branch</label>
              {branchesLoading ? (
                <div className="relative">
                  <select
                    value={formData.git_branch}
                    onChange={(e) => setFormData({ ...formData, git_branch: e.target.value })}
                    disabled
                    className="w-full px-4 py-2.5 border border-slate-300 rounded-lg bg-slate-50 text-slate-500"
                  >
                    <option value="main">Loading branches...</option>
                  </select>
                  <div className="absolute right-3 top-1/2 -translate-y-1/2">
                    <svg className="animate-spin h-5 w-5 text-slate-400" viewBox="0 0 24 24">
                      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none"/>
                      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"/>
                    </svg>
                  </div>
                </div>
              ) : (
                <select
                  value={formData.git_branch}
                  onChange={(e) => setFormData({ ...formData, git_branch: e.target.value })}
                  className="w-full px-4 py-2.5 border border-slate-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-harbor-500 focus:border-transparent transition-all duration-200"
                >
                  {branches.length === 0 ? (
                    <option value="main">main (enter GitHub URL to fetch branches)</option>
                  ) : (
                    branches.map((branch) => (
                      <option key={branch} value={branch}>{branch}</option>
                    ))
                  )}
                </select>
              )}
            </div>
            <Input
              label="Environment Name (optional)"
              placeholder="my-app"
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-slate-700 mb-2">Visibility</label>
            <div className="flex gap-4">
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="radio"
                  name="visibility"
                  value="private"
                  checked={formData.visibility === 'private'}
                  onChange={(e) => setFormData({ ...formData, visibility: e.target.value })}
                  className="w-4 h-4 text-harbor-600 border-slate-300 focus:ring-harbor-500"
                />
                <span className="text-sm text-slate-700">Private</span>
              </label>
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="radio"
                  name="visibility"
                  value="public"
                  checked={formData.visibility === 'public'}
                  onChange={(e) => setFormData({ ...formData, visibility: e.target.value })}
                  className="w-4 h-4 text-harbor-600 border-slate-300 focus:ring-harbor-500"
                />
                <span className="text-sm text-slate-700">Public</span>
              </label>
            </div>
          </div>
        </div>
        <div className="mt-6 flex justify-end gap-3">
          <Button type="button" variant="secondary" onClick={onClose} disabled={loading}>
            Cancel
          </Button>
          <Button type="submit" loading={loading}>
            Create Environment
          </Button>
        </div>
      </form>
    </Modal>
  );
}

function formatRelativeTime(date: string) {
  const now = new Date();
  const then = new Date(date);
  const diffMs = now.getTime() - then.getTime();
  const diffMins = Math.floor(diffMs / 60000);
  const diffHours = Math.floor(diffMs / 3600000);
  const diffDays = Math.floor(diffMs / 86400000);

  if (diffMins < 1) return 'just now';
  if (diffMins < 60) return `${diffMins}m ago`;
  if (diffHours < 24) return `${diffHours}h ago`;
  if (diffDays < 7) return `${diffDays}d ago`;
  return new Date(date).toLocaleDateString();
}