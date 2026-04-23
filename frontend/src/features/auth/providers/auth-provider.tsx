"use client";

import {
  createContext,
  useContext,
  useEffect,
  useState,
  type PropsWithChildren,
} from "react";
import { usePathname, useRouter } from "next/navigation";
import { ApiError } from "@/components/api/client";
import { authApi } from "@/features/auth/lib/auth-api";
import { clearTokens, getAccessToken, saveTokens } from "@/features/auth/lib/auth-storage";
import type { LoginPayload, RegisterPayload, SessionUser } from "@/features/auth/types";

type AuthContextValue = {
  user: SessionUser | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (payload: LoginPayload) => Promise<void>;
  register: (payload: RegisterPayload) => Promise<void>;
  verifyEmail: (email: string, code: string) => Promise<void>;
  logout: () => Promise<void>;
  refreshSession: () => Promise<void>;
};

const AuthContext = createContext<AuthContextValue | null>(null);

const AUTH_PAGES = new Set(["/login", "/register", "/verification"]);

export function AuthProvider({ children }: PropsWithChildren) {
  const pathname = usePathname();
  const router = useRouter();
  const [user, setUser] = useState<SessionUser | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  const refreshSession = async () => {
    const token = getAccessToken();

    if (!token) {
      setUser(null);
      setIsLoading(false);
      return;
    }

    try {
      const response = await authApi.me();
      setUser({ guid: response.guid });
    } catch (error) {
      if (error instanceof ApiError && error.status === 401) {
        clearTokens();
        setUser(null);
      } else {
        throw error;
      }
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    void refreshSession();
  }, []);

  useEffect(() => {
    if (isLoading) {
      return;
    }

    if (!user && !AUTH_PAGES.has(pathname)) {
      router.replace(`/login?redirect=${encodeURIComponent(pathname)}`);
      return;
    }

    if (user && AUTH_PAGES.has(pathname)) {
      router.replace("/profile");
    }
  }, [isLoading, pathname, router, user]);

  const login = async (payload: LoginPayload) => {
    const response = await authApi.login(payload);
    saveTokens(response.access_token, response.refresh_token);
    await refreshSession();
  };

  const register = async (payload: RegisterPayload) => {
    await authApi.register(payload);
  };

  const verifyEmail = async (email: string, code: string) => {
    await authApi.verification(email, code);
  };

  const logout = async () => {
    try {
      await authApi.logout();
    } catch {
      // Ignore logout transport errors and clear local session anyway.
    } finally {
      clearTokens();
      setUser(null);
      router.replace("/login");
    }
  };

  return (
    <AuthContext.Provider
      value={{
        user,
        isAuthenticated: Boolean(user),
        isLoading,
        login,
        register,
        verifyEmail,
        logout,
        refreshSession,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);

  if (!context) {
    throw new Error("useAuth must be used within AuthProvider");
  }

  return context;
}
