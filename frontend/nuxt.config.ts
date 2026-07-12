export default defineNuxtConfig({
  devtools: { enabled: true },
  modules: ['@pinia/nuxt', '@nuxt/fonts'],
  fonts: {
    families: [
      { name: 'Material Symbols Outlined', provider: 'google' }
    ]
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
