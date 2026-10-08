'use client';

import { useEffect, Suspense } from 'react';
import { useParams, useRouter, useSearchParams } from 'next/navigation';
import { useAuthStore } from '@/lib/store';
import { authApi } from '@/lib/api';
import { Button } from '@/components/ui/Button';
import { Card, CardContent } from '@/components/ui/Card';
import { Github, Loader2, Terminal, AlertCircle, CheckCircle } from 'lucide-react';

function SharePageContent() {
  const params = useParams();
  const router = useRouter();
  const searchParams = useSearchParams();
  const { setAuth } = useAuthStore();

  const token = params.token as string;
  const error = searchParams.get('error');
  const joined = searchParams.get('joined');


  const joinEnvironment = async (shareToken: string) => {
    try {
      const response = await fetch(`/api/v1/share/${shareToken}/join`, {
        method: 'POST',
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
        },
      });
      if (response.ok) {
        const data = await response.json();
        router.push(`/environments/${data.environment_id}?joined=true`);
      } else {
        const errorData = await response.json();
        console.error('Join failed:', errorData);
      }
    } catch (err) {
      console.error('Join error:', err);
    }
  };

  // Handle OAuth callback - user comes back here after GitHub auth
  useEffect(() => {
    const storedToken = localStorage.getItem('share_token');
    const currentToken = token || storedToken;

    if (currentToken && !error && !joined) {
      // Call join API
      joinEnvironment(currentToken);
    }
  }, [token, error, joined]);

  const handleLogin = () => {
    if (token) {
      localStorage.setItem('share_token', token);
    }
    authApi.login();
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-slate-50 via-white to-harbor-50 px-4">
      <div className="w-full max-w-md">
        <div className="bg-white rounded-xl shadow-lg p-8">
          <div className="text-center mb-8">
            <div className="w-16 h-16 bg-harbor-600 rounded-xl flex items-center justify-center mx-auto mb-4">
              <Terminal className="w-8 h-8 text-white" />
            </div>
            <h1 className="text-2xl font-bold text-slate-900 mb-2">Join Environment</h1>
            <p className="text-slate-600">Sign in with GitHub to join this shared environment</p>
          </div>

          {error && (
            <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-lg text-red-700 text-sm">
              <p>Authentication failed: {error}</p>
            </div>
          )}

          {joined === 'true' && (
            <div className="mb-6 p-4 bg-green-50 border border-green-200 rounded-lg text-green-700 text-sm flex items-center gap-2">
              <CheckCircle className="w-5 h-5" />
              <p>Successfully joined the environment!</p>
            </div>
          )}

          <Button 
            onClick={handleLogin} 
            className="w-full" 
            size="lg"
            disabled={loading}
          >
            <Github className="w-5 h-5 mr-2" />
            Continue with GitHub
          </Button>

          <div className="mt-6 text-center text-sm text-slate-500">
            <p>By signing in, you agree to our Terms of Service and Privacy Policy.</p>
          </div>
        </div>
      </div>
    </div>
  );
}

function LoadingFallback() {
  return (
    <div className="min-h-screen flex items-center justify-center">
      <div className="animate-spin rounded-full h-12 w-12 border-4 border-harbor-600 border-t-transparent"></div>
    </div>
  );
}

export default function SharePage() {
  return (
    <Suspense fallback={<LoadingFallback />}>
      <SharePageContent />
    </Suspense>
  );
}