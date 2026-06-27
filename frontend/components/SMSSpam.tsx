"use client";

import { useState } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { api } from "@/lib/api";
import { formatPhone } from "@/lib/phone";
import { useToast } from "@/lib/toast";

export default function SMSSpam() {
  const [mode, setMode] = useState<"single"|"bulk">("single");
  const [phone, setPhone] = useState("");
  const [targets, setTargets] = useState("");
  const [content, setContent] = useState("");
  const [senderId, setSenderId] = useState("Service");
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState("");
  const [history, setHistory] = useState<{msg: string; ok: boolean; ts: number}[]>([]);
  const { toast } = useToast();

  const handleSend = async () => {
    setLoading(true); setResult("");
    try {
      if (mode === "single") {
        const r = await api.sendSMS({ phone_number: phone, content, sender_id: senderId });
        const msg = `Sent — ID: ${r.message_id}`;
        setResult(msg);
        setHistory(p => [{ msg, ok: true, ts: Date.now() }, ...p].slice(0, 20));
        toast("SMS delivered", "success");
      } else {
        const r = await api.sendBulkSMS({ targets, content, sender_id: senderId });
        const msg = `Blasted ${r.sent}/${r.targets} messages`;
        setResult(msg);
        setHistory(p => [{ msg, ok: true, ts: Date.now() }, ...p].slice(0, 20));
        toast(`Blasted ${r.sent} messages`, "success");
      }
    } catch (e: any) {
      setResult(e.message);
      setHistory(p => [{ msg: e.message, ok: false, ts: Date.now() }, ...p].slice(0, 20));
      toast(e.message, "error");
    }
    setLoading(false);
  };

  return (
    <div className="flex flex-col gap-5 pt-2">
      {/* Mode toggle */}
      <motion.div
        initial={{ opacity: 0, y: -10 }}
        animate={{ opacity: 1, y: 0 }}
        className="flex gap-2 justify-center"
      >
        {(["single","bulk"] as const).map(m => (
          <motion.button
            key={m}
            whileTap={{ scale: 0.95 }}
            onClick={() => setMode(m)}
            className={`relative px-5 py-2 rounded-lg text-xs font-mono uppercase tracking-wider transition-all ${
              mode === m
                ? "text-white"
                : "text-[#525252] bg-white/[0.02] border border-white/[0.04] hover:border-white/[0.08]"
            }`}
          >
            {mode === m && (
              <motion.div
                layoutId="sms-mode"
                className="absolute inset-0 bg-gradient-to-r from-red-600 to-red-700 rounded-lg"
                transition={{ type: "spring", stiffness: 400, damping: 30 }}
              />
            )}
            <span className="relative z-10">{m === "single" ? "Single SMS" : "Bulk Blast"}</span>
          </motion.button>
        ))}
      </motion.div>

      {/* Phone / Targets */}
      <AnimatePresence mode="wait">
        {mode === "single" ? (
          <motion.div
            key="single"
            initial={{ opacity: 0, x: -20 }}
            animate={{ opacity: 1, x: 0 }}
            exit={{ opacity: 0, x: 20 }}
            className="card"
          >
            <label className="text-[#525252] text-[10px] font-mono tracking-widest uppercase block mb-2">Target Number</label>
            <input
              className="input-field text-center tracking-widest font-mono"
              placeholder="1+555-555-5555"
              value={formatPhone(phone)}
              onChange={e => setPhone(e.target.value.replace(/[^0-9+]/g, ""))}
            />
          </motion.div>
        ) : (
          <motion.div
            key="bulk"
            initial={{ opacity: 0, x: 20 }}
            animate={{ opacity: 1, x: 0 }}
            exit={{ opacity: 0, x: -20 }}
            className="card"
          >
            <label className="text-[#525252] text-[10px] font-mono tracking-widest uppercase block mb-2">
              Targets <span className="text-[#404040]">— one per line</span>
            </label>
            <textarea
              className="textarea-field font-mono text-sm"
              rows={4}
              placeholder={"1+555-555-5555\n1+555-999-8888\n1+555-123-4567"}
              value={targets}
              onChange={e => setTargets(e.target.value)}
            />
            <div className="text-[#404040] text-[10px] font-mono mt-2">
              {targets.split("\n").filter(l => l.trim()).length} numbers loaded
            </div>
          </motion.div>
        )}
      </AnimatePresence>

      {/* Sender ID */}
      <motion.div
        initial={{ opacity: 0, y: 10 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.05 }}
        className="card"
      >
        <label className="text-[#525252] text-[10px] font-mono tracking-widest uppercase block mb-2">Sender ID</label>
        <input
          className="input-field font-mono"
          placeholder="WellsFargo, IRS, Coinbase..."
          value={senderId}
          onChange={e => setSenderId(e.target.value)}
        />
        <div className="text-[#404040] text-[10px] font-mono mt-2">
          Appears as the sender name on the victim's phone
        </div>
      </motion.div>

      {/* Message */}
      <motion.div
        initial={{ opacity: 0, y: 10 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.1 }}
        className="card"
      >
        <label className="text-[#525252] text-[10px] font-mono tracking-widest uppercase block mb-2">Message Content</label>
        <textarea
          className="textarea-field text-sm"
          rows={4}
          placeholder="Your account has been compromised. Call immediately at 1-800-555-0199 to verify your identity..."
          value={content}
          onChange={e => setContent(e.target.value)}
        />
        <div className="flex justify-between mt-2">
          <span className="text-[#404040] text-[10px] font-mono">{content.length}/160 chars</span>
          <span className={`text-[10px] font-mono ${content.length > 160 ? "text-yellow-400" : "text-[#404040]"}`}>
            {content.length > 160 ? "Multi-segment SMS" : "Single segment"}
          </span>
        </div>
      </motion.div>

      {/* Send button */}
      <motion.button
        whileTap={{ scale: 0.97 }}
        onClick={handleSend}
        disabled={loading || !content || (!phone && mode === "single") || (!targets && mode === "bulk")}
        className="w-full py-4 rounded-xl bg-gradient-to-r from-red-600 to-red-700 text-white font-bold text-sm
          hover:from-red-500 hover:to-red-600 transition-all disabled:opacity-30 disabled:cursor-not-allowed
          shadow-lg shadow-red-500/20 flex items-center justify-center gap-2"
      >
        {loading ? (
          <span className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
        ) : (
          <>
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
              <path d="M22 2L11 13M22 2l-7 20-4-9-9-4 20-7z"/>
            </svg>
            {mode === "single" ? "SEND SMS" : "BLAST CAMPAIGN"}
          </>
        )}
      </motion.button>

      {/* Result */}
      <AnimatePresence>
        {result && (
          <motion.div
            initial={{ opacity: 0, y: -5, scale: 0.98 }}
            animate={{ opacity: 1, y: 0, scale: 1 }}
            exit={{ opacity: 0 }}
            className={`text-center text-xs font-mono py-2 rounded-lg border ${
              history.length > 0 && history[0].ok
                ? "text-green-400 bg-green-500/5 border-green-500/10"
                : "text-red-400 bg-red-500/5 border-red-500/10"
            }`}
          >
            {result}
          </motion.div>
        )}
      </AnimatePresence>

      {/* History */}
      {history.length > 0 && (
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          className="card"
        >
          <div className="text-[10px] text-[#525252] font-mono tracking-widest uppercase mb-3">Send History</div>
          <div className="space-y-1.5 max-h-40 overflow-y-auto">
            {history.map((h, i) => (
              <motion.div
                key={h.ts}
                initial={{ opacity: 0, x: -10 }}
                animate={{ opacity: 1, x: 0 }}
                transition={{ delay: i * 0.02 }}
                className="flex items-center justify-between text-[11px] font-mono py-1.5 border-b border-white/[0.03] last:border-0"
              >
                <span className={h.ok ? "text-green-400" : "text-red-400"}>{h.msg}</span>
                <span className="text-[#404040] text-[10px]">{new Date(h.ts).toLocaleTimeString()}</span>
              </motion.div>
            ))}
          </div>
        </motion.div>
      )}
    </div>
  );
}
