/**
 * Runs `task` immediately and then every `intervalMs`, skipping a tick while the
 * previous run is still in flight. Returns a stop function (use inside `$effect`).
 */
export function startPolling(task: () => Promise<void> | void, intervalMs: number): () => void {
  let running = false;
  let stopped = false;

  const tick = async (): Promise<void> => {
    if (running || stopped) return;
    running = true;
    try {
      await task();
    } finally {
      running = false;
    }
  };

  void tick();
  const id = setInterval(() => void tick(), intervalMs);
  return () => {
    stopped = true;
    clearInterval(id);
  };
}
