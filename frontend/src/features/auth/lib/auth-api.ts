import { createApiClient } from "@/components/api/client";
import { createAuthApi } from "@/components/api/endpoints/auth";
import { clearTokens, getAccessToken, getRefreshToken, saveTokens } from "@/features/auth/lib/auth-storage";

const baseUrl = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1";

export const apiClient = createApiClient({
  baseUrl,
  getAccessToken,
  onTokenExpired: async () => {
    const refreshToken = getRefreshToken();

    if (!refreshToken) {
      clearTokens();
      return null;
    }

    try {
      const response = await fetch(`${baseUrl}/refresh`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${getAccessToken()}`,
        },
        body: JSON.stringify({ refresh_token: refreshToken }),
      });

      if (!response.ok) {
        clearTokens();
        return null;
      }

      const data = await response.json();
      saveTokens(data.access_token, data.refresh_token);
      return data.access_token as string;
    } catch {
      clearTokens();
      return null;
    }
  },
});

export const authApi = createAuthApi(apiClient);
