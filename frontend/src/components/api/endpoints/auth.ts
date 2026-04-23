import type { ApiClient } from "@/components/api/client";
import type {
  LoginPayload,
  RegisterPayload,
  RegisterResponse,
  TokenResponse,
  UserResponse,
} from "@/features/auth/types";

export const createAuthApi = (client: ApiClient) => ({
  login: (payload: LoginPayload) =>
    client.post<TokenResponse>("/auth/login", { body: JSON.stringify(payload), skipAuth: true }),
  register: (payload: RegisterPayload) =>
    client.post<RegisterResponse>("/auth/register", { body: JSON.stringify(payload), skipAuth: true }),
  verification: (email: string, code: string) =>
    client.post("/auth/verification", { body: JSON.stringify({ email, code }), skipAuth: true }),
  me: () => client.get<UserResponse>("/me"),
  refresh: (refreshToken: string) =>
    client.post<TokenResponse>("/refresh", { body: JSON.stringify({ refresh_token: refreshToken }) }),
  logout: () => client.post("/logout"),
});
