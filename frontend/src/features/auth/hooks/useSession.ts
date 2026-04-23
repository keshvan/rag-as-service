"use client";

import { useAuth } from "@/features/auth/providers/auth-provider";

export function useSession() {
  const { user, isAuthenticated, isLoading } = useAuth();

  return {
    data: user,
    isAuthenticated,
    isLoading,
  };
}
