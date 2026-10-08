'use client';

import { useEffect, useRef } from 'react';

interface EditorProps { 
  filePath: string; 
  content: string; 
  onChange: (value: string) => void; 
  monaco: any;
  editorRef: React.RefObject<any>;
}

export function Editor({ filePath, content, onChange, monaco, editorRef }: any) {
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!monaco || !containerRef.current || editorRef.current) return;

    const editor = monaco.editor.create(containerRef.current, {
      value: content,
      language: getLanguage(filePath),
      theme: 'vs-dark',
      automaticLayout: true,
      minimap: { enabled: false },
      fontSize: 13,
      lineNumbers: 'on',
      wordWrap: 'on',
      tabSize: 2,
      insertSpaces: true,
      scrollBeyondLastLine: false,
    });

    editorRef.current = editor;

    editor.onDidChangeModelContent(() => {
      onChange(editor.getValue());
    });

    return () => {
      editor.dispose();
      editorRef.current = null;
    };
  }, [monaco, filePath, content, onChange]);

  useEffect(() => {
    if (editorRef.current && editorRef.current.getValue() !== content) {
      editorRef.current.setValue(content);
    }
  }, [content]);

  return <div ref={containerRef} className="h-full w-full" />;
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