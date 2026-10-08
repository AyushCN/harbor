'use client';

import { useEffect, Suspense } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { useAuthStore } from '@/lib/store';
import { authApi } from '@/lib/api';
import { Button } from '@/components/ui/Button';
import { Card, CardContent } from '@/components/ui/Card';
import { Github, Loader2, Terminal } from 'lucide-react';

function LoginPageContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { setAuth } = useAuthStore();

  const error = searchParams.get('error');

  useEffect(() => {
    const storedToken = localStorage.getItem('harbor_token');
    if (storedToken) {
      validateToken(storedToken);
    }
  }, [router]);

  const validateToken = async (token: string) => {
    try {
      // Use Next.js proxy to avoid CORS issues
      const response = await fetch('/api/v1/auth/me', {
        credentials: 'include',
      });
      if (response.ok) {
        const userData = await response.json();
        useAuthStore.getState().setAuth(userData, token);
        router.push('/dashboard');
      } else {
        localStorage.removeItem('harbor_token');
      }
    } catch {
      localStorage.removeItem('harbor_token');
    }
  };

  const handleLogin = () => {
    authApi.login();
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-slate-50 via-white to-harbor-50 px-4">
      <Card className="w-full max-w-md">
        <CardContent className="pt-6">
          <div className="text-center mb-8">
            <div className="w-16 h-16 bg-harbor-600 rounded-xl flex items-center justify-center mx-auto mb-4">
              <Terminal className="w-8 h-8 text-white" />
            </div>
            <h1 className="text-2xl font-bold text-slate-900">Welcome back</h1>
            <p className="text-slate-600 mt-2">Sign in with GitHub to continue to Harbor</p>
          </div>

          {error && (
            <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-lg text-red-700 text-sm">
              <p>Authentication failed: {error}</p>
            </div>
          )}

          <Button 
            onClick={handleLogin} 
            className="w-full" 
            size="lg"
            disabled={false}
          >
            <Github className="w-5 h-5 mr-2" />
            Continue with GitHub
          </Button>

          <div className="mt-6 text-center text-sm text-slate-500">
            <p>By signing in, you agree to our Terms of Service and Privacy Policy.</p>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

export default function LoginPage() {
  return (
    <Suspense fallback={<div className="min-h-screen flex items-center justify-center"><Loader2 className="w-8 h-8 animate-spin text-harbor-600" /></div>}>
      <LoginPageContent />
    </Suspense>
  );
}