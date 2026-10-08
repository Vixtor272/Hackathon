<script lang="ts">
  import { get } from 'svelte/store';
  import { app } from '../../container';
  import { errorMessage, type Conversation, type SampleMedia } from '../../domain';
  import ErrorBanner from '../shared/ErrorBanner.svelte';
  import { startPolling } from '../shared/poll';
  import { activePhone } from '../session';
  import ChatPanel from './ChatPanel.svelte';
  import { DEMO_CONTACTS } from './contacts';
  import MediaPicker from './MediaPicker.svelte';
  import PhonePanel from './PhonePanel.svelte';

  const { chatSession } = app;
  const POLL_MS = 2000;

  let phone = $state(get(activePhone) ?? DEMO_CONTACTS[0]?.phone ?? '+593991111111');
  let conversation = $state<Conversation | null>(null);
  let connected = $state(false);
  let busy = $state(false);
  let error = $state<string | null>(null);

  let pickerOpen = $state(false);
  let media = $state<SampleMedia[]>([]);
  let mediaLoading = $state(false);
  let mediaError = $state<string | null>(null);

  $effect(() => {
    const current = phone;
    activePhone.set(current);
    conversation = null;
    return startPolling(() => refresh(current), POLL_MS);
  });

  async function refresh(target: string): Promise<void> {
    try {
      const transcript = await chatSession.loadTranscript(target);
      if (target !== phone) return;
      conversation = transcript;
      connected = true;
    } catch (cause) {
      connected = false;
      error = errorMessage(cause);
    }
  }

  async function run(action: () => Promise<unknown>): Promise<void> {
    busy = true;
    error = null;
    try {
      await action();
      await refresh(phone);
    } catch (cause) {
      error = errorMessage(cause);
    } finally {
      busy = false;
    }
  }

  function sendText(text: string): void {
    void run(() => chatSession.sendText(phone, text));
  }

  function sendImage(mediaId: string): void {
    pickerOpen = false;
    void run(() => chatSession.sendImage(phone, mediaId));
  }

  function reset(): void {
    void run(() => chatSession.reset(phone));
  }

  function selectPhone(next: string): void {
    if (next !== phone) {
      error = null;
      phone = next;
    }
  }

  async function openPicker(): Promise<void> {
    pickerOpen = true;
    if (media.length > 0) return;
    mediaLoading = true;
    mediaError = null;
    try {
      media = await chatSession.listSampleMedia();
    } catch (cause) {
      mediaError = errorMessage(cause);
    } finally {
      mediaLoading = false;
    }
  }
</script>

<div class="layout">
  <PhonePanel
    {phone}
    conversationState={conversation?.state ?? null}
    orderId={conversation?.orderId ?? null}
    {busy}
    onSelect={selectPhone}
    onReset={reset}
  />
  <div class="main">
    <ErrorBanner message={error} onclose={() => (error = null)} />
    <ChatPanel
      {phone}
      messages={conversation?.messages ?? []}
      conversationState={conversation?.state ?? null}
      {connected}
      {busy}
      onSendText={sendText}
      onAttach={() => void openPicker()}
    />
  </div>
</div>

{#if pickerOpen}
  <MediaPicker {media} loading={mediaLoading} error={mediaError} onPick={sendImage} onClose={() => (pickerOpen = false)} />
{/if}

<style>
  .layout {
    display: grid;
    grid-template-columns: 300px minmax(0, 1fr);
    gap: 16px;
    align-items: start;
  }
  .main {
    min-width: 0;
  }
  @media (max-width: 860px) {
    .layout {
      grid-template-columns: 1fr;
    }
  }
</style>
