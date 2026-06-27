"use client";

import { useState, createContext, useContext, useEffect, ReactNode } from "react";
import {
  supabase,
  isSupabaseConfigured,
  signInWithEmail as supabaseSignIn,
  signUpWithEmail as supabaseSignUp,
  signOut as supabaseSignOut,
  getSession,
} from "@/lib/supabase";

interface AuthState {
  token: string | null;
  email: string | null;
  userId: string | null;
  isAdmin: boolean;
  login: (email: string, password: string) => Promise<void>;
  signup: (email: string, password: string) => Promise<void>;
  logout: () => void;
  isLoading: boolean;
  isSupabase: boolean;
}

const AuthContext = createContext<AuthState>({
  token: null,
  email: null,
  userId: null,
  isAdmin: false,
  login: async () => {},
  signup: async () => {},
  logout: () => {},
  isLoading: true,
  isSupabase: false,
});

export function useAuth() {
  return useContext(AuthContext);
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [token, setToken] = useState<string | null>(null);
  const [email, setEmail] = useState<string | null>(null);
  const [userId, setUserId] = useState<string | null>(null);
  const [isAdmin, setIsAdmin] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const isSupabase = isSupabaseConfigured();

  useEffect(() => {
    const init = async () => {
      if (isSupabase && supabase) {
        // Supabase mode: get session, with 3s timeout fallback to demo
        try {
          const session = await Promise.race([
            getSession(),
            new Promise<null>((_, reject) => setTimeout(() => reject(new Error("timeout")), 3000)),
          ]);
          if (session) {
            setToken(session.access_token);
            setEmail(session.user.email || null);
            setUserId(session.user.id);
            localStorage.setItem("auth_token", session.access_token);
            localStorage.setItem("auth_email", session.user.email || "");
            localStorage.setItem("auth_user_id", session.user.id);
          }
        } catch {
          // Supabase unreachable — fall back to demo mode localStorage
          const saved = localStorage.getItem("auth_token");
          if (saved) {
            setToken(saved);
            setEmail(localStorage.getItem("auth_email"));
            setUserId(localStorage.getItem("auth_user_id"));
            setIsAdmin(localStorage.getItem("auth_is_admin") === "true");
          }
        }

        // Listen for auth changes (don't block loading on this)
        try {
          const { data: { subscription } } = supabase.auth.onAuthStateChange(
            (_event, session) => {
              if (session) {
                setToken(session.access_token);
                setEmail(session.user.email || null);
                setUserId(session.user.id);
                localStorage.setItem("auth_token", session.access_token);
                localStorage.setItem("auth_email", session.user.email || "");
                localStorage.setItem("auth_user_id", session.user.id);
              } else {
                setToken(null);
                setEmail(null);
                setUserId(null);
                localStorage.removeItem("auth_token");
                localStorage.removeItem("auth_email");
                localStorage.removeItem("auth_user_id");
              }
            }
          );
          setIsLoading(false);
          return () => subscription?.unsubscribe();
        } catch {
          setIsLoading(false);
        }
      } else {
        // Demo mode: use localStorage
        const saved = localStorage.getItem("auth_token");
        const savedEmail = localStorage.getItem("auth_email");
        const savedUserId = localStorage.getItem("auth_user_id");
        const savedIsAdmin = localStorage.getItem("auth_is_admin") === "true";
        if (saved) {
          setToken(saved);
          setEmail(savedEmail);
          setUserId(savedUserId);
          setIsAdmin(savedIsAdmin);
        }
        setIsLoading(false);
      }
    };

    init();
  }, [isSupabase]);

  const login = async (email: string, password: string) => {
    if (isSupabase && supabase) {
      try {
        const data = await supabaseSignIn(email, password);
        setToken(data.session?.access_token || null);
        setEmail(data.user?.email || null);
        setUserId(data.user?.id || null);
        return;
      } catch {
        // Supabase login failed, fall through to demo mode
      }
    }
    // Demo mode: call backend API
    const res = await fetch("/api/auth/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email, password }),
    });
    if (!res.ok) {
      const err = await res.json().catch(() => ({ error: "Login failed" }));
      throw new Error(err.error || "Login failed");
    }
    const data = await res.json();
    localStorage.setItem("auth_token", data.token);
    localStorage.setItem("auth_email", data.email || data.user_id);
    localStorage.setItem("auth_user_id", data.user_id);
    localStorage.setItem("auth_is_admin", data.is_admin ? "true" : "false");
    setToken(data.token);
    setEmail(data.email || data.user_id);
    setUserId(data.user_id);
    setIsAdmin(data.is_admin || false);
  };

  const signup = async (email: string, password: string) => {
    if (isSupabase && supabase) {
      try {
        const data = await supabaseSignUp(email, password);
        setToken(data.session?.access_token || null);
        setEmail(data.user?.email || null);
        setUserId(data.user?.id || null);
        return;
      } catch {
        // Supabase signup failed, fall through to demo mode
      }
    }
    await login(email, password);
  };

  const logout = async () => {
    if (isSupabase && supabase) {
      await supabaseSignOut();
    }
    localStorage.removeItem("auth_token");
    localStorage.removeItem("auth_email");
    localStorage.removeItem("auth_user_id");
    localStorage.removeItem("auth_is_admin");
    setToken(null);
    setEmail(null);
    setUserId(null);
    setIsAdmin(false);
  };

  return (
    <AuthContext.Provider value={{ token, email, userId, isAdmin, login, signup, logout, isLoading, isSupabase }}>
      {children}
    </AuthContext.Provider>
  );
}
