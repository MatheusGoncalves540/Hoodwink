import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

// O Vite serve a pasta public/ cru. O .wasm.br do replay é brotli puro: sem o
// header Content-Encoding o navegador não descomprime e o sha256 do manifest
// não bate. Em produção o nginx.conf faz o mesmo (e aplica a CSP). O
// wasm_exec.js tem nome fixo, então não pode ser cacheado como imutável.
function replayAssetHeaders(req, res, next) {
  const pathname = (req.url || '').split('?')[0];
  if (pathname.endsWith('.wasm.br')) {
    res.setHeader('Content-Type', 'application/octet-stream');
    res.setHeader('Content-Encoding', 'br');
    res.setHeader('Cache-Control', 'public, max-age=31536000, immutable');
    res.setHeader('Vary', 'Accept-Encoding');
  } else if (pathname === '/replay/wasm_exec.js') {
    res.setHeader('Cache-Control', 'no-cache');
  }
  next();
}

const replayAssets = () => ({
  name: 'hoodwink-replay-assets',
  configureServer(server) {
    server.middlewares.use(replayAssetHeaders);
  },
  configurePreviewServer(server) {
    server.middlewares.use(replayAssetHeaders);
  },
});

export default defineConfig({
  plugins: [react(), replayAssets()],
  // base: '/Hoodwink-Web/',
  // host: true escuta em 0.0.0.0, para os amigos entrarem pelo IP do Radmin
  server: { host: true, port: 5173, strictPort: true },
});
