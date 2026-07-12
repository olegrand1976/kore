export default defineNuxtConfig({
  devtools: { enabled: true },
  modules: ['@pinia/nuxt', '@nuxt/fonts', '@nuxtjs/i18n'],
  fonts: {
    families: [
      { name: 'DM Sans', provider: 'google', weights: [400, 500, 600, 700] }
    ]
  },
  i18n: {
    restructureDir: false,
    locales: [
      { code: 'fr', language: 'fr-FR', file: 'fr.json', name: 'Français' },
      { code: 'en', language: 'en-US', file: 'en.json', name: 'English' }
    ],
    defaultLocale: 'fr',
    lazy: true,
    langDir: 'locales',
    strategy: 'no_prefix',
    detectBrowserLanguage: {
      useCookie: true,
      cookieKey: 'kore-locale',
      fallbackLocale: 'fr'
    }
  },
  app: {
    head: {
      htmlAttrs: { 'data-theme': 'dark' },
      titleTemplate: '%s | Kore',
      title: 'Kore — PSA/ESN modulaire',
      meta: [
        {
          name: 'description',
          content: 'CRA, TMA, congés, budget UO et conformité européenne — suite PSA/ESN modulaire pour ESN et DSI.'
        },
        { name: 'theme-color', content: '#1a1f2e' },
        { name: 'viewport', content: 'width=device-width, initial-scale=1, viewport-fit=cover' }
      ],
      link: [
        { rel: 'icon', type: 'image/svg+xml', href: '/favicon.svg' },
        { rel: 'icon', type: 'image/svg+xml', href: '/brand/kore-emblem.svg' },
        { rel: 'apple-touch-icon', href: '/brand/kore-emblem.svg' },
        {
          rel: 'stylesheet',
          href: 'https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined:opsz,wght,FILL,GRAD@24,400,0,0&display=swap'
        }
      ]
    }
  },
  runtimeConfig: {
    apiBase: process.env.NUXT_API_BASE || 'http://localhost:8080',
    public: {
      apiBase: process.env.NUXT_PUBLIC_API_BASE || 'http://localhost:8080',
      stripePublishableKey: process.env.NUXT_PUBLIC_STRIPE_PUBLISHABLE_KEY || 'pk_test_mock'
    }
  },
  css: ['~/assets/css/main.css'],
  compatibilityDate: '2024-11-01'
})
