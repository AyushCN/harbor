/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  async rewrites() {
    return [
      {
        source: '/api/:path*',
        destination: 'http://localhost:8082/api/:path*',
      },
    ];
  },
  webpack: (config, { isServer, dev }) => {
    config.externals = [...(config.externals || []), { canvas: 'canvas' }];
    
    if (!isServer) {
      config.resolve.fallback = {
        ...config.resolve.fallback,
        fs: false,
        path: false,
        crypto: false,
      };
    }
    
    // Disable terser for monaco-editor to avoid worker parsing issues
    config.optimization.minimizer.forEach((minimizer) => {
      if (minimizer.constructor.name === 'TerserPlugin') {
        const originalTerserOptions = minimizer.options.terserOptions;
        minimizer.options.terserOptions = {
          ...originalTerserOptions,
          module: true,
        };
        minimizer.options.exclude = /monaco-editor/;
      }
    });
    
    // Use null-loader for monaco-editor ESM files that cause issues
    config.module.rules.push({
      test: /monaco-editor[\\/]esm[\\/].*\.js$/,
      use: 'null-loader',
    });
    
    return config;
  },
};

module.exports = nextConfig;