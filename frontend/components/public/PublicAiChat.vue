<script setup lang="ts">
const { t } = useI18n()
const { publicChat, extractFetchError } = useAi()

const message = ref('')
const reply = ref('')
const errorMsg = ref('')
const loading = ref(false)
const sessionId = ref('')
onMounted(() => {
  sessionId.value = globalThis.crypto?.randomUUID?.() ?? String(Date.now())
})

const send = async () => {
  const text = message.value.trim()
  if (!text || loading.value) return
  errorMsg.value = ''
  loading.value = true
  try {
    const res = await publicChat(text, sessionId.value)
    reply.value = res.reply
    message.value = ''
  } catch (err) {
    errorMsg.value = extractFetchError(err)
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <aside class="public-ai-chat" aria-label="Assistant IA Kore">
    <div class="public-ai-chat__header">
      <AppAiBadge variant="assistant" />
      <p class="public-ai-chat__disclosure">{{ $t('ai.chat_disclosure') }}</p>
    </div>
    <p v-if="reply" class="public-ai-chat__reply">{{ reply }}</p>
    <p v-if="errorMsg" class="public-ai-chat__error" role="alert">{{ errorMsg }}</p>
    <form class="public-ai-chat__form" @submit.prevent="send">
      <AppInput
        id="public-ai-message"
        v-model="message"
        :label="$t('ai.chat_placeholder')"
        :disabled="loading"
      />
      <AppButton variant="secondary" size="sm" type="submit" :disabled="loading || !message.trim()">
        {{ $t('ai.chat_send') }}
      </AppButton>
    </form>
  </aside>
</template>

<style scoped>
.public-ai-chat {
  margin-top: var(--kore-space-xl);
  padding: var(--kore-space-lg);
  border-radius: var(--kore-radius-lg);
  border: 1px solid var(--kore-border);
  background: var(--kore-bg-subtle);
  display: grid;
  gap: var(--kore-space-md);
}

.public-ai-chat__header {
  display: grid;
  gap: var(--kore-space-xs);
}

.public-ai-chat__disclosure {
  margin: 0;
  font-size: var(--kore-text-caption);
  color: var(--kore-text-muted);
}

.public-ai-chat__reply {
  margin: 0;
  font-size: var(--kore-text-small);
  line-height: 1.5;
}

.public-ai-chat__error {
  margin: 0;
  color: var(--kore-error);
  font-size: var(--kore-text-caption);
}

.public-ai-chat__form {
  display: grid;
  gap: var(--kore-space-sm);
}

@media (max-width: 768px) {
  .public-ai-chat__form :deep(.app-button) {
    width: 100%;
  }
}
</style>
