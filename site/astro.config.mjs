import { defineConfig } from 'astro/config';

export default defineConfig({
  site: 'https://ai-daily-brief.github.io',
  base: '/',
  output: 'static',
  build: {
    assets: '_assets'
  },
  publicDir: 'public'
});
