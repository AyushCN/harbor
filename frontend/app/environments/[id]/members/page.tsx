'use client';

import { useEffect, useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { useAuthStore } from '@/lib/store';
import { environmentsApi, type Member } from '@/lib/api';
import { Button } from '@/components/ui/Button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card';
import { Input } from '@/components/ui/Input';
import { Badge } from '@/components/ui/Badge';
import { Avatar } from '@/components/ui/Avatar';
import { Modal } from '@/components/ui/Modal';
import { useToast } from '@/components/ui/Toast';
import { 
  ChevronLeft, 
  Users, 
  UserPlus, 
  UserMinus, 
  Settings,
  MoreVertical,
  Mail,
  Trash2
} from 'lucide-react';

export default function EnvironmentMembersPage() {
  const params = useParams();
  const router = useRouter();
  const { user, token } = useAuthStore();
  const { success, error } = useToast();
  
  const envId = params.id as string;
  const [env, setEnv] = useState<any>(null);
  const [members, setMembers] = useState<Member[]>([]);
  const [loading, setLoading] = useState(true);
  const [showInviteModal, setShowInviteModal] = useState(false);
  const [inviteForm, setInviteForm] = useState({ username: '', role: 'COLLABORATOR' as 'COLLABORATOR' | 'VIEWER' });
  const [inviting, setInviting] = useState(false);
  const [removing, setRemoving] = useState<string | null>(null);

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
    } catch (err) {
      console.error('Failed to fetch environment:', err);
      error('Failed to load environment');
      router.push('/dashboard');
    } finally {
      setLoading(false);
    }
  };

  const handleInvite = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!inviteForm.username.trim()) return;
    
    setInviting(true);
    try {
      await environmentsApi.inviteMember(envId, inviteForm);
      success('Invitation sent', `${inviteForm.username} has been invited as ${inviteForm.role}`);
      setShowInviteModal(false);
      setInviteForm({ username: '', role: 'COLLABORATOR' });
      fetchData();
    } catch (err: any) {
      error('Failed to invite member', err.response?.data?.error?.message || 'Unknown error');
    } finally {
      setInviting(false);
    }
  };

  const handleRemoveMember = async (userId: string) => {
    if (!confirm('Are you sure you want to remove this member?')) return;
    
    setRemoving(userId);
    try {
      await environmentsApi.removeMember(envId, userId);
      success('Member removed');
      fetchData();
    } catch (err: any) {
      error('Failed to remove member', err.response?.data?.error?.message || 'Unknown error');
    } finally {
      setRemoving(null);
    }
  };

  const handleRoleChange = async (userId: string, newRole: 'COLLABORATOR' | 'VIEWER') => {
    try {
      await environmentsApi.updateMemberRole(envId, userId, newRole);
      success('Role updated');
      fetchData();
    } catch (err: any) {
      error('Failed to update role', err.response?.data?.error?.message || 'Unknown error');
    }
  };

  if (!token) {
    return <div className="flex items-center justify-center min-h-screen">Loading...</div>;
  }

  if (loading) {
    return <div className="flex items-center justify-center min-h-screen">Loading...</div>;
  }

  const isOwner = env?.role === 'OWNER';

  return (
    <div className="min-h-screen bg-slate-50">
      <header className="bg-white border-b border-slate-200 sticky top-0 z-40">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-16">
            <div className="flex items-center gap-4">
              <button onClick={() => router.push(`/environments/${envId}`)} className="p-2 rounded-lg hover:bg-slate-100">
                <ChevronLeft className="w-5 h-5" />
              </button>
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 bg-harbor-600 rounded-lg flex items-center justify-center">
                  <Users className="w-6 h-6 text-white" />
                </div>
                <div>
                  <h1 className="text-xl font-bold text-slate-900">Members</h1>
                  <p className="text-sm text-slate-500">{env?.name || 'Environment'}</p>
                </div>
              </div>
            </div>
            <div className="flex items-center gap-3">
              {isOwner && (
                <Button onClick={() => setShowInviteModal(true)}>
                  <UserPlus className="w-4 h-4 mr-2" />
                  Invite Member
                </Button>
              )}
            </div>
          </div>
        </div>
      </header>

      <main className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {loading && <div className="h-4 bg-slate-200 animate-pulse rounded mb-6" />}

        <Card>
          <CardHeader className="flex flex-row items-center justify-between">
            <CardTitle>Environment Members ({members.length})</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {members.map((member) => (
                <div key={member.user.id} className="flex items-center justify-between py-4 border-b border-slate-100 last:border-0">
                  <div className="flex items-center gap-4">
                    <Avatar src={member.user.avatar_url} name={member.user.username} size="md" />
                    <div>
                      <p className="font-medium text-slate-900">{member.user.username}</p>
                      <p className="text-sm text-slate-500">Joined {new Date(member.joined_at).toLocaleDateString()}</p>
                    </div>
                  </div>
                  <div className="flex items-center gap-3">
                    <Badge variant={member.role === 'OWNER' ? 'info' : member.role === 'COLLABORATOR' ? 'success' : 'neutral'}>
                      {member.role}
                    </Badge>
                    {isOwner && member.user.id !== user?.id && (
                      <select
                        value={member.role}
                        onChange={(e) => handleRoleChange(member.user.id, e.target.value as 'COLLABORATOR' | 'VIEWER')}
                        className="px-3 py-1.5 text-sm border border-slate-300 rounded-lg bg-white focus:outline-none focus:ring-2 focus:ring-harbor-500"
                      >
                        <option value="COLLABORATOR">Collaborator</option>
                        <option value="VIEWER">Viewer</option>
                      </select>
                    )}
                    {isOwner && member.user.id !== user?.id && (
                      <button
                        onClick={() => handleRemoveMember(member.user.id)}
                        disabled={removing === member.user.id}
                        className="p-2 text-red-600 hover:bg-red-50 rounded-lg transition-colors disabled:opacity-50"
                      >
                        {removing === member.user.id ? (
                          <span className="animate-spin">⟳</span>
                        ) : (
                          <Trash2 className="w-5 h-5" />
                        )}
                      </button>
                    )}
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>

        {isOwner && (
          <Card className="mt-6">
            <CardHeader className="flex flex-row items-center justify-between">
              <CardTitle>Invite New Member</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-slate-600 mb-4">
                Invite a GitHub user to collaborate on this environment. They will receive a notification to join.
              </p>
              <Button onClick={() => setShowInviteModal(true)}>
                <UserPlus className="w-4 h-4 mr-2" />
                Invite Member
              </Button>
            </CardContent>
          </Card>
        )}
      </main>

      <InviteModal
        isOpen={showInviteModal}
        onClose={() => setShowInviteModal(false)}
        onSubmit={handleInvite}
        formData={inviteForm}
        setFormData={setInviteForm}
        loading={inviting}
      />
    </div>
  );
}

function InviteModal({ 
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
  formData: { username: string; role: 'COLLABORATOR' | 'VIEWER' };
  setFormData: React.Dispatch<React.SetStateAction<{ username: string; role: 'COLLABORATOR' | 'VIEWER' }>>;
  loading: boolean;
}) {
  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title="Invite Member"
      description="Enter the GitHub username of the person you want to invite"
      size="md"
    >
      <form onSubmit={onSubmit}>
        <div className="space-y-4">
          <Input
            label="GitHub Username"
            placeholder="github_username"
            value={formData.username}
            onChange={(e) => setFormData({ ...formData, username: e.target.value })}
            required
          />
          <div>
            <label className="block text-sm font-medium text-slate-700 mb-2">Role</label>
            <select
              value={formData.role}
              onChange={(e) => setFormData({ ...formData, role: e.target.value as 'COLLABORATOR' | 'VIEWER' })}
              className="w-full px-4 py-2.5 border border-slate-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-harbor-500"
            >
              <option value="COLLABORATOR">Collaborator (can edit)</option>
              <option value="VIEWER">Viewer (read-only)</option>
            </select>
          </div>
        </div>
        <div className="mt-6 flex justify-end gap-3">
          <Button type="button" variant="secondary" onClick={onClose} disabled={loading}>
            Cancel
          </Button>
          <Button type="submit" loading={loading}>
            Send Invitation
          </Button>
        </div>
      </form>
    </Modal>
  );
}