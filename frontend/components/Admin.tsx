"use client";

import { useState, useEffect } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { api } from "@/lib/api";
import { useToast } from "@/lib/toast";

export default function Admin() {
  const [dash, setDash] = useState<any>({});
  const [tab, setTab] = useState<"overview"|"users"|"dtmf">("overview");
  const [users, setUsers] = useState<any[]>([]);
  const [dtmfLogs, setDtmfLogs] = useState<any[]>([]);
  const [services, setServices] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const { toast } = useToast();

  useEffect(() => {
    Promise.all([
      api.getDashboard().catch(() => ({})),
      api.listUsers().catch(() => ({ users: [] })),
      api.getDTMFLogs().catch(() => ({ dtmf_logs: [] })),
      api.getServices().catch(() => ({ services: [] })),
    ]).then(([d, u, l, s]) => {
      setDash(d);
      setUsers(u.users || []);
      setDtmfLogs(l.dtmf_logs || []);
      setServices(s.services || []);
    }).finally(() => setLoading(false));
  }, []);

  const handleAdjust = async (userId: string, amount: number) => {
    try {
      await api.adjustBalance(userId, amount, "Admin adjustment");
      const r = await api.listUsers();
      setUsers(r.users || []);
      toast(`Adjusted $${amount}`, "success");
    } catch (e: any) { toast(e.message, "error"); }
  };

  const handleVoucher = async () => {
    try {
      const r = await api.generateVoucher("HUSH-25");
      toast(`Voucher: ${r.code} — ${r.tokens} tokens`, "success");
    } catch (e: any) { toast(e.message, "error"); }
  };

  const metrics = [
    { label: "Users", value: dash.total_users ?? "—", color: "text-blue-400", icon: "👥" },
    { label: "Active Calls", value: dash.active_calls ?? "—", color: "text-green-400", icon: "📞" },
    { label: "Total Calls", value: dash.total_calls ?? "—", color: "text-yellow-400", icon: "📊" },
    { label: "Revenue", value: dash.total_revenue ? `$${dash.total_revenue}` : "$0", color: "text-red-400", icon: "💰" },
  ];

  return (
    <div className="flex flex-col gap-5 pt-2">
      {/* Header */}
      <motion.div
        initial={{ opacity: 0, y: -10 }}
        animate={{ opacity: 1, y: 0 }}
        className="text-center"
      >
        <span className="text-2xl font-bold bg-gradient-to-r from-yellow-500 to-red-500 bg-clip-text text-transparent">
          Admin Panel
        </span>
      </motion.div>

      {/* Metrics */}
      <div className="grid grid-cols-4 gap-2">
        {metrics.map((m, i) => (
          <motion.div
            key={m.label}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: i * 0.05 }}
            className="card text-center py-3"
          >
            <div className="text-lg mb-1">{m.icon}</div>
            <div className={`font-mono text-lg font-bold ${m.color}`}>{m.value}</div>
            <div className="text-[#525252] text-[9px] font-mono tracking-wider uppercase mt-1">{m.label}</div>
          </motion.div>
        ))}
      </div>

      {/* Tab bar */}
      <motion.div
        initial={{ opacity: 0, y: 10 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.1 }}
        className="flex gap-2"
      >
        {(["overview","users","dtmf"] as const).map(t => (
          <motion.button
            key={t}
            whileTap={{ scale: 0.95 }}
            onClick={() => setTab(t)}
            className={`relative px-4 py-2 rounded-lg text-xs font-mono uppercase tracking-wider transition-all ${
              tab === t ? "text-white" : "text-[#525252] bg-white/[0.02] border border-white/[0.04]"
            }`}
          >
            {tab === t && (
              <motion.div
                layoutId="admin-tab"
                className="absolute inset-0 bg-gradient-to-r from-yellow-600 to-red-600 rounded-lg"
                transition={{ type: "spring", stiffness: 400, damping: 30 }}
              />
            )}
            <span className="relative z-10">{t}</span>
          </motion.button>
        ))}
        <motion.button
          whileTap={{ scale: 0.95 }}
          onClick={handleVoucher}
          className="px-4 py-2 rounded-lg text-xs font-mono bg-purple-600/20 text-purple-400 border border-purple-500/30
            hover:bg-purple-600/30 transition-all"
        >
          + Voucher
        </motion.button>
      </motion.div>

      {/* Tab content */}
      <AnimatePresence mode="wait">
        {tab === "users" && (
          <motion.div
            key="users"
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -10 }}
            className="card max-h-96 overflow-y-auto"
          >
            {loading ? (
              <div className="space-y-2">
                {[1,2,3].map(i => <div key={i} className="skeleton h-12 rounded-lg" />)}
              </div>
            ) : users.length === 0 ? (
              <div className="text-center py-8 text-[#404040] text-xs font-mono">No users</div>
            ) : (
              <div className="space-y-2">
                {users.map((u: any, i: number) => (
                  <motion.div
                    key={u.id}
                    initial={{ opacity: 0, x: -10 }}
                    animate={{ opacity: 1, x: 0 }}
                    transition={{ delay: i * 0.02 }}
                    className="flex items-center justify-between text-xs py-3 border-b border-white/[0.03] last:border-0"
                  >
                    <div>
                      <div className="text-[#a3a3a3] font-mono">{u.email}</div>
                      <div className="text-[#525252] text-[10px] font-mono">${u.balance} • {u.total_calls} calls</div>
                    </div>
                    <div className="flex gap-2">
                      <motion.button
                        whileTap={{ scale: 0.9 }}
                        onClick={() => handleAdjust(u.id, 50)}
                        className="text-green-400 text-[10px] font-mono hover:text-green-300 transition-colors"
                      >
                        +$50
                      </motion.button>
                      <motion.button
                        whileTap={{ scale: 0.9 }}
                        onClick={() => handleAdjust(u.id, -50)}
                        className="text-red-400 text-[10px] font-mono hover:text-red-300 transition-colors"
                      >
                        -$50
                      </motion.button>
                    </div>
                  </motion.div>
                ))}
              </div>
            )}
          </motion.div>
        )}

        {tab === "dtmf" && (
          <motion.div
            key="dtmf"
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -10 }}
            className="card max-h-96 overflow-y-auto"
          >
            <div className="text-[10px] text-[#525252] font-mono tracking-widest uppercase mb-3">All DTMF Captures</div>
            {loading ? (
              <div className="space-y-2">
                {[1,2,3].map(i => <div key={i} className="skeleton h-8 rounded-lg" />)}
              </div>
            ) : dtmfLogs.length === 0 ? (
              <div className="text-center py-8 text-[#404040] text-xs font-mono">No DTMF captures</div>
            ) : (
              <div className="space-y-1.5">
                {dtmfLogs.map((l: any, i: number) => (
                  <motion.div
                    key={l.id}
                    initial={{ opacity: 0, x: -10 }}
                    animate={{ opacity: 1, x: 0 }}
                    transition={{ delay: i * 0.02 }}
                    className="flex items-center justify-between text-[11px] font-mono py-2 border-b border-white/[0.03] last:border-0"
                  >
                    <span className="text-[#a3a3a3]">{l.spoofed_caller_id} → {l.destination}</span>
                    <span className="text-green-400 font-bold">{l.digit} <span className="text-[#404040]">@ {l.timestamp_ms}ms</span></span>
                  </motion.div>
                ))}
              </div>
            )}
          </motion.div>
        )}

        {tab === "overview" && (
          <motion.div
            key="overview"
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -10 }}
            className="card"
          >
            <div className="text-[10px] text-[#525252] font-mono tracking-widest uppercase mb-3">System Overview</div>
            {loading ? (
              <div className="grid grid-cols-2 gap-4">
                {[1,2,3,4].map(i => <div key={i} className="skeleton h-16 rounded-lg" />)}
              </div>
            ) : (
              <div className="grid grid-cols-2 gap-4">
                {services.map((s: any, i: number) => (
                  <motion.div
                    key={s.name}
                    initial={{ opacity: 0, scale: 0.95 }}
                    animate={{ opacity: 1, scale: 1 }}
                    transition={{ delay: i * 0.05 }}
                    className="bg-white/[0.02] rounded-lg p-3 border border-white/[0.03]"
                  >
                    <div className="text-[#525252] text-[10px] font-mono uppercase">{s.name}</div>
                    <div className={`text-xs font-mono mt-1 ${
                      s.status === "connected" || s.status === "ready"
                        ? s.mode === "live" ? "text-green-400" : "text-yellow-400"
                        : "text-red-400"
                    }`}>
                      {s.status === "connected" ? "Connected" : s.status === "ready" ? "Ready" : "Disconnected"}
                      <span className="text-[#404040] ml-1">({s.mode})</span>
                    </div>
                  </motion.div>
                ))}
                {services.length === 0 && (
                  <div className="col-span-2 text-center py-4 text-[#404040] text-xs font-mono">No services detected</div>
                )}
              </div>
            )}
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}
