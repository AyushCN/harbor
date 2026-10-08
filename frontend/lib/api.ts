import axios, { AxiosError, InternalAxiosRequestConfig } from 'axios';

// Use relative URLs to go through Next.js proxy
export const api = axios.create({
  baseURL: '/api/v1',
  withCredentials: true,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Add auth token from localStorage if available (fallback for non-cookie auth)
api.interceptors.request.use((config: InternalAxiosRequestConfig) => {
  if (typeof window !== 'undefined') {
    const token = localStorage.getItem('harbor_token');
    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`;
    }
  }
  return config;
});

// Handle auth errors
api.interceptors.response.use(
  (response) => response,
  (error: AxiosError) => {
    if (error.response?.status === 401) {
      if (typeof window !== 'undefined') {
        localStorage.removeItem('harbor_token');
        window.location.href = '/login';
      }
    }
    return Promise.reject(error);
  }
);

// Auth
export const authApi = {
  login: () => window.location.href = '/api/v1/auth/github',
  logout: async () => {
    try {
      await api.post('/auth/logout');
    } catch (err) {
      console.error('Logout error:', err);
    }
    // Clear all auth state
    if (typeof window !== 'undefined') {
      localStorage.removeItem('harbor_token');
      localStorage.removeItem('harbor-auth');
      sessionStorage.clear();
    }
    // Hard redirect to login
    if (typeof window !== 'undefined') {
      window.location.href = '/login';
    }
  },
  me: () => api.get('/auth/me'),
};

// Environments
export const environmentsApi = {
  list: () => api.get('/environments'),
  create: (data: { git_url: string; git_branch?: string; name?: string }) => 
    api.post('/environments', data),
  get: (id: string) => api.get(`/environments/${id}`),
  start: (id: string) => api.post(`/environments/${id}/start`),
  stop: (id: string) => api.post(`/environments/${id}/stop`),
  resume: (id: string) => api.post(`/environments/${id}/resume`),
  delete: (id: string) => api.delete(`/environments/${id}`),
  
  // Members
  listMembers: (id: string) => api.get(`/environments/${id}/members`),
  inviteMember: (id: string, data: { username: string; role: 'COLLABORATOR' | 'VIEWER' }) => 
    api.post(`/environments/${id}/members`, data),
  updateMemberRole: (id: string, userId: string, role: 'COLLABORATOR' | 'VIEWER') => 
    api.patch(`/environments/${id}/members/${userId}`, { role }),
  removeMember: (id: string, userId: string) => 
    api.delete(`/environments/${id}/members/${userId}`),
  
  // Share links
  createShareLink: (id: string, data: { role: 'COLLABORATOR' | 'VIEWER'; expires_in_hours?: number; max_uses?: number }) =>
    api.post(`/environments/${id}/share-links`, data),
  
  // Presence
  getPresence: (id: string) => api.get(`/environments/${id}/presence`),
};

// Share links
export const shareApi = {
  join: (token: string) => api.post(`/share/${token}/join`),
};

export type Environment = {
  id: string;
  name: string;
  status: 'CREATED' | 'BUILDING' | 'RUNNING' | 'STOPPED' | 'SUSPENDED' | 'CRASHED' | 'BUILD_FAILED';
  public_url?: string;
  host: {
    id: string;
    username: string;
    avatar_url?: string;
  };
  role: 'OWNER' | 'COLLABORATOR' | 'VIEWER';
  created_at: string;
  workspace?: {
    id: string;
    name?: string;
    git_url: string;
    git_branch: string;
  };
};

export type Member = {
  user: {
    id: string;
    username: string;
    avatar_url?: string;
  };
  role: 'OWNER' | 'COLLABORATOR' | 'VIEWER';
  joined_at: string;
};

export type ShareLink = {
  token: string;
  url: string;
  role: 'COLLABORATOR' | 'VIEWER';
  expires_at?: string;
  max_uses?: number;
};

export type PresenceUser = {
  id: string;
  username: string;
  avatar_url?: string;
  role: string;
  joined_at: string;
};

export type PresenceResponse = {
  environment_id: string;
  current_user: { id: string; role: string };
  users: PresenceUser[];
  server_time: string;
};