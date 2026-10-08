import { ApiError } from '../../domain';

export type QueryParams = Record<string, string | undefined>;

/** Minimal JSON HTTP client abstraction so gateways never touch `fetch` directly. */
export interface HttpClient {
  get<T>(path: string, query?: QueryParams): Promise<T>;
  post<T>(path: string, body?: unknown): Promise<T>;
  patch<T>(path: string, body: unknown): Promise<T>;
  delete(path: string): Promise<void>;
}

interface ErrorEnvelope {
  error?: { code?: string; message?: string };
}

type FetchLike = (input: string, init?: RequestInit) => Promise<Response>;

export class FetchHttpClient implements HttpClient {
  private readonly fetchFn: FetchLike;

  constructor(
    private readonly baseUrl: string = '/api/v1',
    fetchFn?: FetchLike,
  ) {
    this.fetchFn = fetchFn ?? ((input, init) => fetch(input, init));
  }

  get<T>(path: string, query?: QueryParams): Promise<T> {
    return this.request<T>('GET', withQuery(path, query));
  }

  post<T>(path: string, body?: unknown): Promise<T> {
    return this.request<T>('POST', path, body);
  }

  patch<T>(path: string, body: unknown): Promise<T> {
    return this.request<T>('PATCH', path, body);
  }

  async delete(path: string): Promise<void> {
    await this.request<unknown>('DELETE', path);
  }

  private async request<T>(method: string, path: string, body?: unknown): Promise<T> {
    let response: Response;
    try {
      response = await this.fetchFn(`${this.baseUrl}${path}`, {
        method,
        headers: body === undefined ? { Accept: 'application/json' } : { Accept: 'application/json', 'Content-Type': 'application/json' },
        body: body === undefined ? undefined : JSON.stringify(body),
      });
    } catch {
      throw new ApiError('NETWORK', 'No se pudo conectar con el servidor de Farmi', 0);
    }

    if (response.status === 204) {
      return undefined as T;
    }

    const text = await response.text();
    const payload = text.length > 0 ? safeParse(text) : undefined;

    if (!response.ok) {
      const envelope = (payload ?? {}) as ErrorEnvelope;
      throw new ApiError(
        envelope.error?.code ?? 'INTERNAL',
        envelope.error?.message ?? `Error ${response.status} del servidor`,
        response.status,
      );
    }
    return payload as T;
  }
}

function safeParse(text: string): unknown {
  try {
    return JSON.parse(text) as unknown;
  } catch {
    throw new ApiError('INTERNAL', 'El servidor devolvió una respuesta inválida');
  }
}

function withQuery(path: string, query?: QueryParams): string {
  if (!query) return path;
  const params = new URLSearchParams();
  for (const [key, value] of Object.entries(query)) {
    if (value !== undefined && value !== '') params.set(key, value);
  }
  const encoded = params.toString();
  return encoded.length > 0 ? `${path}?${encoded}` : path;
}
