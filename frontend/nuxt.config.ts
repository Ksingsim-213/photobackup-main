export default defineNuxtConfig({
  ssr: false,
  devtools: { enabled: false },
  telemetry: false,
  compatibilityDate: '2025-01-01',
  devServer: { host: '127.0.0.1', port: 3000 },

  alias: {
    '~/wailsjs': './wailsjs'
  },

  experimental: {
    appManifest: false
  },

  css: [
    '@fontsource/ibm-plex-sans-thai/400.css',
    '@fontsource/ibm-plex-sans-thai/500.css',
    '@fontsource/ibm-plex-sans-thai/600.css',
    '~/assets/main.css',
  ],
})