import { describe, expect, it } from 'vitest';
import { InMemoryChatGateway } from '../../adapters/memory';
import { isApiError } from '../../domain';
import { ChatSession, normalizePhone } from './ChatSession';

describe('ChatSession', () => {
  it('rejects empty messages before reaching the channel', async () => {
    const session = new ChatSession(new InMemoryChatGateway());
    await expect(session.sendText('+593991111111', '   ')).rejects.toSatisfy(
      (error: unknown) => isApiError(error) && error.code === 'VALIDATION',
    );
  });

  it('records inbound and outbound messages in the transcript and advances the state', async () => {
    const session = new ChatSession(new InMemoryChatGateway());
    const phone = '+593 99 111 1111';

    const greeting = await session.sendText(phone, 'hola');
    expect(greeting.state).toBe('ASK_ID');
    expect(greeting.replies[0]?.text).toContain('cédula');

    const afterId = await session.sendText(phone, '1712345678');
    expect(afterId.state).toBe('ASK_PRESCRIPTION');

    const afterImage = await session.sendImage(phone, 'receta-001');
    expect(afterImage.state).toBe('ASK_ZONE');

    const transcript = await session.loadTranscript(phone);
    expect(transcript.phone).toBe('+593991111111');
    expect(transcript.messages.map((message) => message.direction)).toEqual(['in', 'out', 'in', 'out', 'in', 'out']);
    expect(transcript.messages[4]?.type).toBe('image');
    expect(transcript.messages[4]?.mediaUrl).toBe('/recetas/receta-001.svg');
  });

  it('resets a conversation back to the start', async () => {
    const session = new ChatSession(new InMemoryChatGateway());
    await session.sendText('+593992222222', '0912345678');
    await session.reset('+593992222222');
    const transcript = await session.loadTranscript('+593992222222');
    expect(transcript.state).toBe('ASK_ID');
    expect(transcript.messages).toHaveLength(0);
  });

  it('normalizes phone numbers', () => {
    expect(normalizePhone('+593 (99) 111-1111')).toBe('+593991111111');
    expect(normalizePhone('0991111111')).toBe('0991111111');
  });
});
