'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useAuthStore } from '@/lib/store';
import { authApi } from '@/lib/api';
import { Button } from '@/components/ui/Button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card';
import { Input } from '@/components/ui/Input';
import { Loader2, Github, Terminal, Users, Globe, Zap } from 'lucide-react';

export default function HomePage() {
  const router = useRouter();
  const { user, token, setAuth } = useAuthStore();

  useEffect(() => {
    // Check if we're returning from OAuth
    const urlParams = new URLSearchParams(window.location.search);
    const error = urlParams.get('error');
    
    if (error) {
      // Handle OAuth error
      console.error('OAuth error:', error);
      urlParams.delete('error');
      window.history.replaceState({}, '', '/');
    }
    
    // Check if we have a valid token
    const storedToken = localStorage.getItem('harbor_token');
    if (storedToken && !token) {
      // Token exists but not in store, validate it
      validateToken(storedToken);
    } else if (!storedToken && !token) {
      // No token, user needs to login
    } else if (token) {
      // Already authenticated
      router.push('/dashboard');
    }
  }, [token, router]);

  const validateToken = async (token: string) => {
    try {
      const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8087'}/api/v1/auth/me`, {
        headers: { Authorization: `Bearer ${token}` },
      });
      if (response.ok) {
        const userData = await response.json();
        setAuth(userData, token);
        router.push('/dashboard');
      } else {
        localStorage.removeItem('harbor_token');
      }
    } catch (error) {
      localStorage.removeItem('harbor_token');
    }
  };

  const handleLogin = () => {
    authApi.login();
  };

  if (token) {
    return <div className="flex items-center justify-center min-h-screen"><Loader2 className="w-8 h-8 animate-spin text-harbor-600" /></div>;
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 via-white to-harbor-50">
      <nav className="border-b border-slate-200 bg-white/80 backdrop-blur-sm sticky top-0 z-40">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-16">
            <div className="flex items-center gap-2">
              <div className="w-8 h-8 bg-harbor-600 rounded-lg flex items-center justify-center">
                <Terminal className="w-5 h-5 text-white" />
              </div>
              <span className="text-xl font-bold text-slate-900">Harbor</span>
            </div>
            <div className="hidden md:flex items-center gap-4">
              <Button variant="ghost" onClick={handleLogin}>
                <Github className="w-4 h-4 mr-2" />
                Sign in with GitHub
              </Button>
            </div>
          </div>
        </div>
      </nav>

      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-20">
        <div className="text-center max-w-3xl mx-auto">
          <h1 className="text-5xl md:text-6xl font-bold text-slate-900 tracking-tight mb-6">
            Collaborative Development
            <br />
            <span className="text-harbor-600">Environments</span>
          </h1>
          <p className="text-xl text-slate-600 mb-10 max-w-2xl mx-auto">
            Spin up isolated development environments from any GitHub repository in seconds.
            Real-time collaboration, live previews, and browser-based IDE — all with clear ownership.
          </p>
          
          <div className="flex flex-col sm:flex-row items-center justify-center gap-4 mb-16">
            <Button size="lg" onClick={handleLogin} className="w-full sm:w-auto">
              <Github className="w-5 h-5 mr-2" />
              Start with GitHub
            </Button>
          </div>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mt-16">
          <FeatureCard
            icon={<Zap className="w-6 h-6" />}
            title="Zero-Config Environments"
            description="Paste a GitHub URL — we detect the language, build the container, and give you a live URL."
          />
          <FeatureCard
            icon={<Users className="w-6 h-6" />}
            title="Real-Time Collaboration"
            description="Invite team members, see who's online, and work together in the same environment with presence awareness."
          />
          <FeatureCard
            icon={<Globe className="w-6 h-6" />}
            title="Live Preview URLs"
            description="Every environment gets a public HTTPS URL. Share your work instantly with stakeholders."
          />
        </div>

        <div className="mt-20 text-center">
          <Card className="border-harbor-200 bg-harbor-50">
            <CardContent className="pt-6">
              <h3 className="text-xl font-semibold text-slate-900 mb-2">How it works</h3>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-6 text-left mt-6">
                <StepCard step={1} title="Connect GitHub" description="Sign in with GitHub OAuth and authorize Harbor to access your repositories." />
                <StepCard step={2} title="Create Environment" description="Paste any GitHub repository URL. We'll detect the stack and build your container." />
                <StepCard step={3} title="Collaborate" description="Invite teammates, share preview URLs, and code together in real-time." />
              </div>
            </CardContent>
          </Card>
        </div>
      </main>

      <footer className="border-t border-slate-200 bg-white mt-20">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <p className="text-center text-slate-500 text-sm">
            Harbor — Collaborative development environments with clear ownership
          </p>
        </div>
      </footer>
    </div>
  );
}

function FeatureCard({ icon, title, description }: { icon: React.ReactNode; title: string; description: string }) {
  return (
    <Card className="h-full hover:shadow-lg transition-shadow">
      <CardContent className="pt-6">
        <div className="w-12 h-12 bg-harbor-100 text-harbor-600 rounded-xl flex items-center justify-center mb-4">
          {icon}
        </div>
        <h3 className="text-lg font-semibold text-slate-900 mb-2">{title}</h3>
        <p className="text-slate-600">{description}</p>
      </CardContent>
    </Card>
  );
}

function StepCard({ step, title, description }: { step: number; title: string; description: string }) {
  return (
    <div className="relative pl-8 pb-8 border-l-2 border-harbor-200 last:border-0">
      <div className="absolute left-0 top-0 w-6 h-6 bg-harbor-600 text-white rounded-full flex items-center justify-center text-sm font-bold">
        {step}
      </div>
      <h4 className="font-semibold text-slate-900 mb-1">{title}</h4>
      <p className="text-slate-600 text-sm">{description}</p>
    </div>
  );
}