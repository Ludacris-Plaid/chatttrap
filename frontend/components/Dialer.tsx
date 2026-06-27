"use client";

import { useState, useRef, useEffect } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { api } from "@/lib/api";
import { normalizePhone, formatPhone, isCompletePhone } from "@/lib/phone";
import { useToast } from "@/lib/toast";
import ActiveCall from "./ActiveCall";

const keypad = [
  ["1","2","3"],
  ["4","5","6"],
  ["7","8","9"],
  ["*","0","#"],
];

export default function Dialer() {
  const [spoofedCID, setSpoofedCID] = useState("");
  const [destination, setDestination] = useState("");
  const [cnam, setCnam] = useState("");
  const [cnamLoading, setCnamLoading] = useState(false);
  const [callActive, setCallActive] = useState(false);
  const [callId, setCallId] = useState("");
  const [error, setError] = useState("");
  const lastLookedUp = useRef("");
  const { toast } = useToast();

  const [focusedField, setFocusedField] = useState<"cid" | "dest">("cid");

  const handleKey = (k: string) => {
    if (focusedField === "dest") setDestination(p => p + k);
    else setSpoofedCID(p => p + k);
  };
  const handleBackspace = () => {
    if (focusedField === "dest") setDestination(p => p.slice(0, -1));
    else setSpoofedCID(p => p.slice(0, -1));
  };
  const handleClear = () => { setSpoofedCID(""); setDestination(""); setCnam(""); lastLookedUp.current = ""; };

  const handleCIDChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setSpoofedCID(e.target.value.replace(/[^0-9+]/g, ""));
  };

  const handleDestChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setDestination(e.target.value.replace(/[^0-9+]/g, ""));
  };

  useEffect(() => {
    if (!isCompletePhone(spoofedCID)) { setCnam(""); return; }
    const n = normalizePhone(spoofedCID);
    if (n === lastLookedUp.current) return;
    lastLookedUp.current = n;
    setCnamLoading(true);
    api.lookupCNAM(n)
      .then(res => setCnam(res.name))
      .catch(() => setCnam(""))
      .finally(() => setCnamLoading(false));
  }, [spoofedCID]);

  const handleCall = async () => {
    if (!isCompletePhone(spoofedCID) || !isCompletePhone(destination)) {
      setError("Enter complete numbers (1+ area+local)");
      return;
    }
    setError("");
    const cid = normalizePhone(spoofedCID);
    const dest = normalizePhone(destination);
    try {
      const res = await api.originateCall({
        spoofed_caller_id: cid,
        spoofed_name: cnam || "Unknown",
        destination_number: dest,
      });
      setCallId(res.call_id);
      setCallActive(true);
      toast("Call initiated", "success");
      } catch (e: any) {
        const msg = e.message || "Call failed";
        if (msg.includes("403") || msg.includes("rejected") || msg.includes("call failed")) {
          setError("SIP trunk rejected spoofed ID — use account caller ID or check trunk config");
        } else {
          setError(msg);
        }
        toast(msg, "error");
      }
  };

  if (callActive) {
    return <ActiveCall callId={callId} spoofedCID={formatPhone(spoofedCID)} cnam={cnam} destination={formatPhone(destination)} onEnd={() => { setCallActive(false); setCallId(""); }} />;
  }

  const cidFmt = formatPhone(spoofedCID);
  const destFmt = formatPhone(destination);

  return (
    <div className="flex flex-col items-center gap-5 pt-2 w-full max-w-sm mx-auto">
      {/* Trunk status */}
      <motion.div
        initial={{ opacity: 0, y: -10 }}
        animate={{ opacity: 1, y: 0 }}
        className="card w-full text-center py-3"
      >
        <div className="flex items-center justify-center gap-2">
          <div className="status-dot" />
          <span className="text-[#525252] text-[10px] font-mono tracking-widest uppercase">Connected to premium spoofing trunks</span>
        </div>
      </motion.div>

      {/* Spoofed CID */}
      <motion.div
        initial={{ opacity: 0, x: -20 }}
        animate={{ opacity: 1, x: 0 }}
        transition={{ delay: 0.05 }}
        className="w-full space-y-2"
      >
        <label className="text-[#525252] text-[10px] font-mono tracking-widest uppercase block">Spoofed Caller ID</label>
        <input
          className={`input-field text-lg text-center tracking-[0.2em] font-mono ${focusedField === "cid" ? "!border-red-500/40 !bg-red-500/[0.03]" : ""}`}
          placeholder="1+555-555-5555"
          value={cidFmt}
          onChange={handleCIDChange}
          onFocus={() => setFocusedField("cid")}
        />
        <AnimatePresence>
          {cnam && (
            <motion.div
              initial={{ opacity: 0, scale: 0.95, y: -5 }}
              animate={{ opacity: 1, scale: 1, y: 0 }}
              exit={{ opacity: 0 }}
              className="text-center"
            >
              <span className="inline-flex items-center gap-1.5 text-green-400 text-xs font-mono bg-green-500/5 px-3 py-1 rounded-full border border-green-500/10">
                <span className="w-1.5 h-1.5 rounded-full bg-green-400" />
                Will display as: {cnam}
              </span>
            </motion.div>
          )}
          {cnamLoading && (
            <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} className="text-[#404040] text-[10px] text-center font-mono">
              Resolving CNAM...
            </motion.div>
          )}
        </AnimatePresence>
      </motion.div>

      {/* Destination */}
      <motion.div
        initial={{ opacity: 0, x: 20 }}
        animate={{ opacity: 1, x: 0 }}
        transition={{ delay: 0.1 }}
        className="w-full"
      >
        <label className="text-[#525252] text-[10px] font-mono tracking-widest uppercase block mb-2">Destination Number</label>
        <input
          className={`input-field text-lg text-center tracking-[0.2em] font-mono ${focusedField === "dest" ? "!border-red-500/40 !bg-red-500/[0.03]" : ""}`}
          placeholder="1+555-555-5555"
          value={destFmt}
          onChange={handleDestChange}
          onFocus={() => setFocusedField("dest")}
        />
      </motion.div>

      {/* Keypad */}
      <motion.div
        initial={{ opacity: 0, scale: 0.95 }}
        animate={{ opacity: 1, scale: 1 }}
        transition={{ delay: 0.15 }}
        className="flex flex-col items-center gap-2.5 w-full"
      >
        <div className="text-[#404040] text-[9px] font-mono tracking-widest uppercase mb-1">
          Editing: <span className="text-red-400">{focusedField === "cid" ? "Caller ID" : "Destination"}</span>
          {' '}&mdash; click input box above to switch
        </div>
        {keypad.map((row, i) => (
          <div key={i} className="flex gap-3 w-full">
            {row.map(k => (
              <motion.button
                key={k}
                whileTap={{ scale: 0.85 }}
                className="dtmf-btn dtmf-btn-wide"
                onClick={() => handleKey(k)}
              >
                {k}
              </motion.button>
            ))}
          </div>
        ))}
      </motion.div>

      {/* Action buttons */}
      <motion.div
        initial={{ opacity: 0, y: 10 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.2 }}
        className="flex gap-3 w-full"
      >
        <motion.button
          whileTap={{ scale: 0.95 }}
          onClick={handleBackspace}
          className="btn-secondary flex-1 text-sm py-3"
        >
          ⌫
        </motion.button>
        <motion.button
          whileTap={{ scale: 0.97 }}
          onClick={handleCall}
          className="flex-1 py-3 rounded-lg bg-gradient-to-r from-green-600 to-green-700 text-white font-bold text-sm
            hover:from-green-500 hover:to-green-600 transition-all disabled:opacity-30 disabled:cursor-not-allowed
            shadow-lg shadow-green-500/20 flex items-center justify-center gap-2"
          disabled={!isCompletePhone(spoofedCID) || !isCompletePhone(destination)}
        >
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
            <path d="M22 16.92v3a2 2 0 0 1-2.18 2 19.79 19.79 0 0 1-8.63-3.07 19.5 19.5 0 0 1-6-6 19.79 19.79 0 0 1-3.07-8.67A2 2 0 0 1 4.11 2h3a2 2 0 0 1 2 1.72c.127.96.361 1.903.7 2.81a2 2 0 0 1-.45 2.11L8.09 9.91a16 16 0 0 0 6 6l1.27-1.27a2 2 0 0 1 2.11-.45c.907.339 1.85.573 2.81.7A2 2 0 0 1 22 16.92z"/>
          </svg>
          CALL
        </motion.button>
        <motion.button
          whileTap={{ scale: 0.95 }}
          onClick={handleClear}
          className="btn-secondary flex-1 text-sm py-3"
        >
          CLR
        </motion.button>
      </motion.div>

      <AnimatePresence>
        {error && (
          <motion.div
            initial={{ opacity: 0, y: -5 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0 }}
            className="text-red-400 text-xs font-mono"
          >
            {error}
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}
