"use client";

import { useState } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { useAuth } from "@/lib/auth";
import ParticleBackground from "./ParticleBackground";

export default function LoginOverlay() {
  const { login, signup, isLoading, isSupabase } = useAuth();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const [mode, setMode] = useState<"login" | "signup">("login");

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!email) { setError("Email required"); return; }
    setLoading(true);
    setError("");
    try {
      if (mode === "signup") {
        await signup(email, password || "demo123456");
      } else {
        await login(email, password || "demo");
      }
    } catch (e: any) {
      setError(e.message || "Login failed");
    }
    setLoading(false);
  };

  if (isLoading) {
    return (
      <div className="fixed inset-0 bg-[#0a0a0a] flex items-center justify-center z-50">
        <motion.div
          animate={{ rotate: 360 }}
          transition={{ duration: 1, repeat: Infinity, ease: "linear" }}
          className="w-8 h-8 border-2 border-red-500/30 border-t-red-500 rounded-full"
        />
      </div>
    );
  }

  return (
    <div className="fixed inset-0 bg-[#0a0a0a] flex items-center justify-center z-50 overflow-hidden">
      <ParticleBackground />

      {/* Gradient orbs */}
      <motion.div
        animate={{
          x: [0, 100, -50, 0],
          y: [0, -80, 60, 0],
          scale: [1, 1.2, 0.9, 1],
        }}
        transition={{ duration: 20, repeat: Infinity, ease: "linear" }}
        className="absolute -top-32 -left-32 w-96 h-96 bg-red-600/10 rounded-full blur-[120px]"
      />
      <motion.div
        animate={{
          x: [0, -80, 50, 0],
          y: [0, 60, -40, 0],
          scale: [1, 0.8, 1.1, 1],
        }}
        transition={{ duration: 25, repeat: Infinity, ease: "linear" }}
        className="absolute -bottom-32 -right-32 w-96 h-96 bg-purple-600/10 rounded-full blur-[120px]"
      />

      <motion.div
        initial={{ opacity: 0, scale: 0.9, y: 20 }}
        animate={{ opacity: 1, scale: 1, y: 0 }}
        transition={{ duration: 0.5, ease: [0.22, 1, 0.36, 1] }}
        className="w-full max-w-sm mx-4 relative z-10"
      >
        {/* Logo */}
        <motion.div
          initial={{ opacity: 0, y: -20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.2, duration: 0.6 }}
          className="text-center mb-8"
        >
          <div className="inline-flex items-center gap-3 mb-4">
            <motion.div
              className="relative"
              whileHover={{ scale: 1.05 }}
              transition={{ type: "spring", stiffness: 400, damping: 20 }}
            >
              <div className="w-14 h-14 rounded-2xl bg-gradient-to-br from-red-500 via-red-600 to-red-700 flex items-center justify-center shadow-lg shadow-red-500/30">
                <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2.5">
                  <path d="M22 16.92v3a2 2 0 0 1-2.18 2 19.79 19.79 0 0 1-8.63-3.07 19.5 19.5 0 0 1-6-6 19.79 19.79 0 0 1-3.07-8.67A2 2 0 0 1 4.11 2h3a2 2 0 0 1 2 1.72c.127.96.361 1.903.7 2.81a2 2 0 0 1-.45 2.11L8.09 9.91a16 16 0 0 0 6 6l1.27-1.27a2 2 0 0 1 2.11-.45c.907.339 1.85.573 2.81.7A2 2 0 0 1 22 16.92z"/>
                </svg>
              </div>
              <motion.div
                animate={{ scale: [1, 1.4, 1], opacity: [0.3, 0.1, 0.3] }}
                transition={{ duration: 3, repeat: Infinity }}
                className="absolute -inset-2 rounded-2xl bg-red-500/20 blur-md"
              />
            </motion.div>
            <div className="text-left">
              <motion.div
                initial={{ opacity: 0, x: -10 }}
                animate={{ opacity: 1, x: 0 }}
                transition={{ delay: 0.3 }}
                className="text-white font-bold text-2xl tracking-tight"
              >
                HUSH<span className="text-red-500">CIRCUITS</span>
              </motion.div>
              <motion.div
                initial={{ opacity: 0, x: -10 }}
                animate={{ opacity: 1, x: 0 }}
                transition={{ delay: 0.4 }}
                className="text-[#525252] text-[10px] tracking-[0.3em] uppercase font-mono"
              >
                Pro v2 • AI Voice Platform
              </motion.div>
            </div>
          </div>

          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 0.5 }}
            className="text-[#525252] text-xs"
          >
            {isSupabase ? "Sign in with your Supabase account" : "Enter your credentials to access the platform"}
          </motion.div>
        </motion.div>

        {/* Login Form */}
        <motion.form
          onSubmit={handleSubmit}
          initial={{ opacity: 0, y: 10 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.4, duration: 0.5 }}
          className="space-y-3"
        >
          {/* Mode toggle */}
          <div className="flex gap-2 mb-4">
            {(["login", "signup"] as const).map(m => (
              <motion.button
                key={m}
                type="button"
                whileTap={{ scale: 0.95 }}
                onClick={() => { setMode(m); setError(""); }}
                className={`relative flex-1 py-2 rounded-lg text-xs font-mono uppercase tracking-wider transition-all ${
                  mode === m ? "text-white" : "text-[#525252] bg-white/[0.02] border border-white/[0.04]"
                }`}
              >
                {mode === m && (
                  <motion.div
                    layoutId="auth-mode"
                    className="absolute inset-0 bg-gradient-to-r from-red-600 to-red-700 rounded-lg"
                    transition={{ type: "spring", stiffness: 400, damping: 30 }}
                  />
                )}
                <span className="relative z-10">{m === "login" ? "Sign In" : "Sign Up"}</span>
              </motion.button>
            ))}
          </div>

          <div>
            <label className="text-[#525252] text-[10px] font-mono tracking-widest uppercase block mb-1.5">Email</label>
            <input
              type="email"
              className="input-field text-center font-mono"
              placeholder="agent@hushcircuits.io"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              autoFocus
            />
          </div>
          <div>
            <label className="text-[#525252] text-[10px] font-mono tracking-widest uppercase block mb-1.5">Password</label>
            <input
              type="password"
              className="input-field text-center font-mono"
              placeholder={mode === "signup" ? "Min 6 characters" : "••••••••"}
              value={password}
              onChange={(e) => setPassword(e.target.value)}
            />
          </div>

          <motion.button
            type="submit"
            disabled={loading || !email}
            whileHover={{ scale: 1.01 }}
            whileTap={{ scale: 0.98 }}
            className="w-full py-4 rounded-xl bg-gradient-to-r from-red-600 to-red-700 text-white font-bold text-sm tracking-wide
              hover:from-red-500 hover:to-red-600 transition-all disabled:opacity-40 disabled:cursor-not-allowed
              shadow-lg shadow-red-500/20 flex items-center justify-center gap-2"
          >
            {loading ? (
              <motion.span
                animate={{ rotate: 360 }}
                transition={{ duration: 1, repeat: Infinity, ease: "linear" }}
                className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full"
              />
            ) : (
              <>
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
                  <rect x="3" y="11" width="18" height="11" rx="2" ry="2"/>
                  <path d="M7 11V7a5 5 0 0 1 10 0v4"/>
                </svg>
                {mode === "login" ? "ACCESS PLATFORM" : "CREATE ACCOUNT"}
              </>
            )}
          </motion.button>

          <AnimatePresence>
            {error && (
              <motion.div
                initial={{ opacity: 0, y: -5, scale: 0.98 }}
                animate={{ opacity: 1, y: 0, scale: 1 }}
                exit={{ opacity: 0, y: -5 }}
                className="text-red-400 text-xs text-center font-mono py-2 rounded-lg bg-red-500/5 border border-red-500/10"
              >
                {error}
              </motion.div>
            )}
          </AnimatePresence>
        </motion.form>

        {/* Footer */}
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ delay: 0.7 }}
          className="text-center mt-8 space-y-2"
        >
          <div className="flex items-center justify-center gap-4 text-[#3a3a3a] text-[10px] font-mono">
            <span className="flex items-center gap-1">
              <span className="w-1.5 h-1.5 rounded-full bg-green-500/50" />
              {isSupabase ? "Supabase Auth" : "Demo Mode"}
            </span>
            <span>•</span>
            <span>JWT Encrypted</span>
          </div>
          <div className="text-[#2a2a2a] text-[9px] font-mono">
            {mode === "signup" ? "Account will be created via Supabase Auth" : "No password? Use any email to enter demo mode"}
          </div>
        </motion.div>
      </motion.div>
    </div>
  );
}
