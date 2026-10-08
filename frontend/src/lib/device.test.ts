import { describe, expect, it } from 'vitest';
import { classifyDevice } from './device';

const MAC = 'Mozilla/5.0 (Macintosh; Intel Mac OS X 14_6) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.0 Safari/605.1.15';
const WINDOWS = 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0 Safari/537.36';
const IPHONE = 'Mozilla/5.0 (iPhone; CPU iPhone OS 18_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148';
const ANDROID = 'Mozilla/5.0 (Linux; Android 15; Pixel 9) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0 Mobile Safari/537.36';

describe('classifyDevice', () => {
  it('recognises computers', () => {
    expect(classifyDevice({ userAgent: MAC, maxTouchPoints: 0, finePointer: true })).toBe('desktop');
    expect(classifyDevice({ userAgent: WINDOWS, maxTouchPoints: 10, finePointer: true })).toBe('desktop');
  });

  it('recognises phones and tablets', () => {
    expect(classifyDevice({ userAgent: IPHONE, maxTouchPoints: 5, finePointer: false })).toBe('mobile');
    expect(classifyDevice({ userAgent: ANDROID, maxTouchPoints: 5, finePointer: false })).toBe('mobile');
    expect(classifyDevice({ userAgent: MAC, maxTouchPoints: 5, finePointer: false })).toBe('mobile'); // iPadOS
    expect(classifyDevice({ userAgent: WINDOWS, uaMobile: true, maxTouchPoints: 5, finePointer: true })).toBe('mobile');
  });
});
