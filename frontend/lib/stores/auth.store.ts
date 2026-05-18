"use client";

import { create } from "zustand";
import { persist } from "zustand/middleware";

export type AuthUser = {
  id: string;
  name: string;
  email: string;
  platformRole: "user" | "super_admin";
};

type AuthState = {
  accessToken: string | null;
  user: AuthUser | null;
  setSession: (payload: { accessToken: string; user: AuthUser }) => void;
  clearSession: () => void;
};

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      accessToken: null,
      user: null,
      setSession: ({ accessToken, user }) => set({ accessToken, user }),
      clearSession: () => set({ accessToken: null, user: null })
    }),
    {
      name: "umkm-auth"
    }
  )
);
