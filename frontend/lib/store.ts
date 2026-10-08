import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { Environment, Member, PresenceUser } from '@/lib/api';

interface User {
  id: string;
  username: string;
  name?: string;
  avatar_url?: string;
}

interface AuthState {
  user: User | null;
  token: string | null;
  setAuth: (user: User, token: string) => void;
  logout: () => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      user: null,
      token: null,
      setAuth: (user, token) => set({ user, token }),
      logout: () => {
        // Clear all auth-related storage
        if (typeof window !== 'undefined') {
          localStorage.removeItem('harbor-auth');
          localStorage.removeItem('harbor_token');
          localStorage.removeItem('share_token');
          sessionStorage.clear();
        }
        set({ user: null, token: null });
      },
    }),
    { name: 'harbor-auth' }
  )
);

interface EnvironmentState {
  environments: Environment[];
  currentEnvironment: Environment | null;
  members: Member[];
  presenceUsers: PresenceUser[];
  setEnvironments: (envs: Environment[]) => void;
  addEnvironment: (env: Environment) => void;
  updateEnvironment: (id: string, updates: Partial<Environment>) => void;
  removeEnvironment: (id: string) => void;
  setCurrentEnvironment: (env: Environment | null) => void;
  setMembers: (members: Member[]) => void;
  addMember: (member: Member) => void;
  updateMemberRole: (userId: string, role: Member['role']) => void;
  removeMember: (userId: string) => void;
  setPresenceUsers: (users: PresenceUser[]) => void;
  addPresenceUser: (user: PresenceUser) => void;
  removePresenceUser: (userId: string) => void;
  updatePresenceUser: (userId: string, updates: Partial<PresenceUser>) => void;
}

export const useEnvironmentStore = create<EnvironmentState>((set) => ({
  environments: [],
  currentEnvironment: null,
  members: [],
  presenceUsers: [],
  
  setEnvironments: (environments) => set({ environments }),
  addEnvironment: (environment) => set((state) => ({ 
    environments: [environment, ...state.environments] 
  })),
  updateEnvironment: (id, updates) => set((state) => ({
    environments: state.environments.map((e) => 
      e.id === id ? { ...e, ...updates } : e
    ),
    currentEnvironment: state.currentEnvironment?.id === id 
      ? { ...state.currentEnvironment, ...updates } 
      : state.currentEnvironment,
  })),
  removeEnvironment: (id) => set((state) => ({
    environments: state.environments.filter((e) => e.id !== id),
    currentEnvironment: state.currentEnvironment?.id === id ? null : state.currentEnvironment,
  })),
  setCurrentEnvironment: (currentEnvironment) => set({ currentEnvironment }),
  
  setMembers: (members) => set({ members }),
  addMember: (member) => set((state) => ({ 
    members: [...state.members, member] 
  })),
  updateMemberRole: (userId, role) => set((state) => ({
    members: state.members.map((m) => 
      m.user.id === userId ? { ...m, role } : m
    ),
  })),
  removeMember: (userId) => set((state) => ({
    members: state.members.filter((m) => m.user.id !== userId),
  })),
  
  setPresenceUsers: (presenceUsers) => set({ presenceUsers }),
  addPresenceUser: (user) => set((state) => ({ 
    presenceUsers: [...state.presenceUsers.filter(u => u.id !== user.id), user] 
  })),
  removePresenceUser: (userId) => set((state) => ({
    presenceUsers: state.presenceUsers.filter((u) => u.id !== userId),
  })),
  updatePresenceUser: (userId, updates) => set((state) => ({
    presenceUsers: state.presenceUsers.map((u) => 
      u.id === userId ? { ...u, ...updates } : u
    ),
  })),
}));

interface UIState {
  sidebarOpen: boolean;
  toggleSidebar: () => void;
  setSidebarOpen: (open: boolean) => void;
}

export const useUIStore = create<UIState>((set) => ({
  sidebarOpen: true,
  toggleSidebar: () => set((state) => ({ sidebarOpen: !state.sidebarOpen })),
  setSidebarOpen: (sidebarOpen) => set({ sidebarOpen }),
}));