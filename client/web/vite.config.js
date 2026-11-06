import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  // base: '/Hoodwink-Web/',
  // host: true escuta em 0.0.0.0, para os amigos entrarem pelo IP do Radmin
  server: { host: true, port: 5173, strictPort: true },
});