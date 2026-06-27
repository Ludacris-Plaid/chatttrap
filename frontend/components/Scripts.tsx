"use client";

import { useState, useEffect } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { api } from "@/lib/api";
import { useToast } from "@/lib/toast";

const goals = ["OTP Theft", "Card Details", "Wire Fraud", "Account Takeover", "Impersonation"];

export default function Scripts() {
  const [name, setName] = useState("");
  const [age, setAge] = useState(30);
  const [service, setService] = useState("");
  const [details, setDetails] = useState("");
  const [goal, setGoal] = useState("OTP Theft");
  const [script, setScript] = useState("");
  const [history, setHistory] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);
  const { toast } = useToast();

  useEffect(() => { api.listScripts().then(r => setHistory(r.scripts || [])).catch(()=>{}); }, []);

  const handleGenerate = async () => {
    if (!name) return;
    setLoading(true);
    try {
      const r = await api.generateScript({ target_name: name, target_age: age, target_service: service, target_details: details, goal, script_type: "custom" });
      setScript(r.script);
      setHistory(prev => [{ id: r.id, content: r.script, goal, created_at: new Date() }, ...prev]);
      toast("Script generated", "success");
    } catch (e: any) { setScript(`Error: ${e.message}`); toast(e.message, "error"); }
    setLoading(false);
  };

  return (
    <div className="flex flex-col gap-5 pt-2">
      {/* Header */}
      <motion.div
        initial={{ opacity: 0, y: -10 }}
        animate={{ opacity: 1, y: 0 }}
        className="text-center"
      >
        <span className="text-2xl font-bold bg-gradient-to-r from-red-500 to-purple-500 bg-clip-text text-transparent">
          ScriptForge
        </span>
        <span className="text-[#525252] text-xs font-mono ml-2">AI ENGINE</span>
      </motion.div>

      {/* Target profile */}
      <motion.div
        initial={{ opacity: 0, x: -20 }}
        animate={{ opacity: 1, x: 0 }}
        transition={{ delay: 0.05 }}
        className="card space-y-3"
      >
        <div className="text-[10px] text-[#525252] font-mono tracking-widest uppercase">Target Profile</div>
        <div>
          <label className="text-[#525252] text-[10px] font-mono block mb-1">VICTIM NAME</label>
          <input
            className="input-field font-mono"
            placeholder="John Smith"
            value={name}
            onChange={e => setName(e.target.value)}
          />
        </div>
        <div className="flex gap-3">
          <div className="flex-1">
            <label className="text-[#525252] text-[10px] font-mono block mb-1">AGE</label>
            <input
              type="number"
              className="input-field font-mono"
              value={age}
              onChange={e => setAge(+e.target.value)}
            />
          </div>
          <div className="flex-1">
            <label className="text-[#525252] text-[10px] font-mono block mb-1">TARGET SERVICE</label>
            <input
              className="input-field font-mono"
              placeholder="Chase, Coinbase..."
              value={service}
              onChange={e => setService(e.target.value)}
            />
          </div>
        </div>
        <div>
          <label className="text-[#525252] text-[10px] font-mono block mb-1">KNOWN DETAILS</label>
          <textarea
            className="textarea-field font-mono text-sm"
            rows={2}
            placeholder="Recent transactions, fears, vulnerabilities..."
            value={details}
            onChange={e => setDetails(e.target.value)}
          />
        </div>
      </motion.div>

      {/* Goal selector */}
      <motion.div
        initial={{ opacity: 0, y: 10 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.1 }}
        className="card"
      >
        <label className="text-[#525252] text-[10px] font-mono tracking-widest uppercase block mb-3">Objective</label>
        <div className="flex flex-wrap gap-2">
          {goals.map(g => (
            <motion.button
              key={g}
              whileTap={{ scale: 0.95 }}
              onClick={() => setGoal(g)}
              className={`relative px-3 py-1.5 rounded-lg text-[11px] font-mono transition-all ${
                goal === g
                  ? "text-white"
                  : "text-[#525252] bg-white/[0.02] border border-white/[0.04] hover:border-white/[0.08]"
              }`}
            >
              {goal === g && (
                <motion.div
                  layoutId="script-goal"
                  className="absolute inset-0 bg-gradient-to-r from-red-600 to-purple-600 rounded-lg"
                  transition={{ type: "spring", stiffness: 400, damping: 30 }}
                />
              )}
              <span className="relative z-10">{g}</span>
            </motion.button>
          ))}
        </div>
      </motion.div>

      {/* Generate */}
      <motion.button
        whileTap={{ scale: 0.97 }}
        onClick={handleGenerate}
        disabled={loading || !name}
        className="w-full py-4 rounded-xl bg-gradient-to-r from-red-600 to-purple-600 text-white font-bold text-sm
          hover:from-red-500 hover:to-purple-500 transition-all disabled:opacity-30 disabled:cursor-not-allowed
          shadow-lg shadow-red-500/20 flex items-center justify-center gap-2"
      >
        {loading ? (
          <span className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
        ) : (
          <>
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
              <path d="M12 2a7 7 0 0 1 7 7c0 2.38-1.19 4.47-3 5.74V17a2 2 0 0 1-2 2h-4a2 2 0 0 1-2-2v-2.26C6.19 13.47 5 11.38 5 9a7 7 0 0 1 7-7z"/>
              <path d="M9 22h6M10 2v1M14 2v1"/>
            </svg>
            GENERATE SCRIPT
          </>
        )}
      </motion.button>

      {/* Generated script */}
      <AnimatePresence>
        {script && (
          <motion.div
            initial={{ opacity: 0, y: 20, scale: 0.98 }}
            animate={{ opacity: 1, y: 0, scale: 1 }}
            exit={{ opacity: 0, y: -20 }}
            className="card"
          >
            <div className="flex items-center justify-between mb-3">
              <div className="text-[10px] text-[#525252] font-mono tracking-widest uppercase">Generated Script</div>
              <motion.button
                whileTap={{ scale: 0.95 }}
                onClick={() => { navigator.clipboard.writeText(script); toast("Copied to clipboard", "success"); }}
                className="text-[10px] font-mono text-red-400 hover:text-red-300 transition-colors"
              >
                COPY
              </motion.button>
            </div>
            <div className="bg-white/[0.02] rounded-lg p-4 border border-white/[0.04] max-h-80 overflow-y-auto">
              <pre className="text-xs text-[#a3a3a3] leading-relaxed whitespace-pre-wrap font-mono">{script}</pre>
            </div>
          </motion.div>
        )}
      </AnimatePresence>

      {/* Script library */}
      {history.length > 0 && (
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          className="card"
        >
          <div className="text-[10px] text-[#525252] font-mono tracking-widest uppercase mb-3">
            Library <span className="text-[#404040]">({history.length})</span>
          </div>
          <div className="space-y-2 max-h-48 overflow-y-auto">
            {history.map((s: any, i: number) => (
              <motion.div
                key={s.id || i}
                initial={{ opacity: 0, x: -10 }}
                animate={{ opacity: 1, x: 0 }}
                transition={{ delay: i * 0.02 }}
                className="text-xs bg-white/[0.02] rounded-lg p-3 border border-white/[0.03] cursor-pointer
                  hover:border-red-500/20 hover:bg-red-500/[0.02] transition-all"
                onClick={() => setScript(s.content)}
              >
                <div className="flex justify-between items-center">
                  <span className="text-red-400 font-mono text-[11px]">{s.goal || "Script"}</span>
                  <span className="text-[#404040] text-[10px] font-mono">{new Date(s.created_at).toLocaleDateString()}</span>
                </div>
                <div className="truncate text-[#525252] mt-1 font-mono">{s.content?.slice(0, 80)}...</div>
              </motion.div>
            ))}
          </div>
        </motion.div>
      )}
    </div>
  );
}
