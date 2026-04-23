type RequestConfig = RequestInit & { skipAuth?: boolean };

export class ApiClient {
  private baseUrl: string;
  private getAccessToken: () => string | null;
  private onTokenExpired: () => Promise<string | null>;
  private isRefreshing = false;
  private refreshQueue: Array<{ resolve: () => void; reject: (err: unknown) => void }> = [];

  constructor(config: {
    baseUrl: string;
    getAccessToken: () => string | null;
    onTokenExpired: () => Promise<string | null>;
  }) {
    this.baseUrl = config.baseUrl;
    this.getAccessToken = config.getAccessToken;
    this.onTokenExpired = config.onTokenExpired;
  }

  async request<T>(path: string, config: RequestConfig = {}): Promise<T> {
    const { skipAuth, headers, ...rest } = config;
    const url = `${this.baseUrl}${path}`;

    const defaultHeaders: Record<string, string> = {
      "Content-Type": "application/json",
      ...((headers as Record<string, string>) || {}),
    };

    if (!skipAuth) {
      const token = this.getAccessToken();
      if (token) {
        defaultHeaders.Authorization = `Bearer ${token}`;
      }
    }

    const response = await fetch(url, { ...rest, headers: defaultHeaders });

    if (response.status === 401 && !skipAuth) {
      return this.handle401<T>(path, config);
    }

    if (!response.ok) {
      const error = await response.json().catch(() => ({}));
      throw new ApiError(response.status, error.error || error.message || "Unknown error", error.code);
    }

    if (response.status === 204) {
      return undefined as T;
    }

    return response.json();
  }

  get<T>(path: string, config: RequestConfig = {}) {
    return this.request<T>(path, { ...config, method: "GET" });
  }

  post<T>(path: string, config: RequestConfig = {}) {
    return this.request<T>(path, { ...config, method: "POST" });
  }

  private async handle401<T>(path: string, config: RequestConfig): Promise<T> {
    if (!this.isRefreshing) {
      this.isRefreshing = true;

      try {
        const newToken = await this.onTokenExpired();

        if (!newToken) {
          throw new Error("Refresh failed");
        }

        this.refreshQueue.forEach(({ resolve }) => resolve());
        this.refreshQueue = [];

        return this.request<T>(path, { ...config, skipAuth: false });
      } catch (error) {
        this.refreshQueue.forEach(({ reject }) => reject(error));
        this.refreshQueue = [];
        throw error;
      } finally {
        this.isRefreshing = false;
      }
    }

    return new Promise((resolve, reject) => {
      this.refreshQueue.push({
        resolve: () => resolve(this.request<T>(path, config)),
        reject,
      });
    });
  }
}

export class ApiError extends Error {
  constructor(
    public status: number,
    message: string,
    public code?: string,
  ) {
    super(message);
    this.name = "ApiError";
  }
}

export const createApiClient = (config: ConstructorParameters<typeof ApiClient>[0]) =>
  new ApiClient(config);
