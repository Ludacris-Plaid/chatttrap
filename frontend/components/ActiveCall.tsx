"use client";

import { useState, useEffect, useRef } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { api } from "@/lib/api";

interface Props {
  callId: string;
  spoofedCID: string;
  cnam: string;
  destination: string;
  onEnd: () => void;
}

export default function ActiveCall({ callId, spoofedCID, cnam, destination, onEnd }: Props) {
  const [elapsed, setElapsed] = useState(0);
  const [muted, setMuted] = useState(false);
  const [dtmf, setDtmf] = useState("");
  const [ended, setEnded] = useState(false);
  const [cost, setCost] = useState(0);
  const timer = useRef<ReturnType<typeof setInterval>>();
  const endTimerRef = useRef<ReturnType<typeof setTimeout>>();

  useEffect(() => {
    timer.current = setInterval(() => {
      setElapsed(t => t + 1);
      setCost(c => c + 0.0083);
    }, 1000);

    // Poll call state every 3 seconds
    const poll = setInterval(async () => {
      try {
        const call = await api.getCall(callId);
        if (call && (call.status === "completed" || call.status === "failed" || call.status === "cancelled")) {
          clearInterval(poll);
          clearInterval(timer.current);
          setEnded(true);
          endTimerRef.current = setTimeout(onEnd, 5000);
        }
      } catch {}
    }, 3000);

    return () => {
      clearInterval(timer.current);
      clearInterval(poll);
      if (endTimerRef.current) clearTimeout(endTimerRef.current);
    };
  }, [callId, onEnd]);

  const fmt = (s: number) => `${Math.floor(s/60).toString().padStart(2,"0")}:${(s%60).toString().padStart(2,"0")}`;

  const handleDTMF = async (digit: string) => {
    setDtmf(p => p + digit);
    try { await api.submitDTMF(callId, digit, Date.now()); } catch {}
  };

  const handleHangup = async () => {
    clearInterval(timer.current);
    try { await api.endCall(callId); } catch {}
    setEnded(true);
    endTimerRef.current = setTimeout(onEnd, 5000);
  };

  if (ended) {
    const tokens = Math.max(1, Math.ceil(elapsed / 60));
    return (
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        className="flex flex-col items-center justify-center h-full gap-6 text-center"
      >
        <motion.div
          initial={{ scale: 0 }}
          animate={{ scale: 1 }}
          transition={{ type: "spring", stiffness: 300, damping: 20 }}
          className="w-16 h-16 rounded-full bg-green-500/10 flex items-center justify-center border border-green-500/20"
        >
          <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="#22c55e" strokeWidth="2.5">
            <polyline points="20 6 9 17 4 12"/>
          </svg>
        </motion.div>
        <div className="text-xl font-semibold text-white">Call Completed</div>
        <div className="card w-full max-w-sm">
          <div className="grid grid-cols-2 gap-6 text-sm">
            <div>
              <div className="text-[#525252] text-[10px] font-mono uppercase tracking-wider mb-1">Duration</div>
              <div className="text-white font-mono text-lg">{fmt(elapsed)}</div>
            </div>
            <div>
              <div className="text-[#525252] text-[10px] font-mono uppercase tracking-wider mb-1">Cost</div>
              <div className="text-white font-mono text-lg">${cost.toFixed(2)}</div>
            </div>
            <div>
              <div className="text-[#525252] text-[10px] font-mono uppercase tracking-wider mb-1">Tokens</div>
              <div className="text-yellow-400 font-mono text-lg">{tokens}</div>
            </div>
            <div>
              <div className="text-[#525252] text-[10px] font-mono uppercase tracking-wider mb-1">DTMF</div>
              <div className="text-green-400 font-mono text-lg">{dtmf || "—"}</div>
            </div>
          </div>
        </div>
        <div className="text-[#404040] text-xs font-mono">Returning to dialer...</div>
      </motion.div>
    );
  }

  return (
    <motion.div
      initial={{ opacity: 0, scale: 0.98 }}
      animate={{ opacity: 1, scale: 1 }}
      className="flex flex-col items-center gap-5 pt-2"
    >
      {/* Live indicator */}
      <motion.div
        initial={{ opacity: 0, y: -10 }}
        animate={{ opacity: 1, y: 0 }}
        className="flex items-center gap-2"
      >
        <span className="recording-dot" />
        <span className="text-red-400 text-xs font-mono font-semibold tracking-widest">LIVE</span>
      </motion.div>

      {/* Timer */}
      <div className="text-center">
        <motion.div
          key={elapsed}
          className="text-5xl font-mono tracking-widest text-white font-light"
        >
          {fmt(elapsed)}
        </motion.div>
        <motion.div
          key={cost.toFixed(2)}
          initial={{ scale: 1.1 }}
          animate={{ scale: 1 }}
          className="text-sm text-red-400 font-mono mt-1"
        >
          ${cost.toFixed(4)}
        </motion.div>
      </div>

      {/* Call info */}
      <div className="card w-full max-w-sm text-center py-4">
        <div className="text-[10px] text-[#525252] font-mono tracking-widest uppercase mb-1">Spoofed CID</div>
        <div className="text-white font-mono">{spoofedCID}</div>
        {cnam && <div className="text-green-400 text-xs mt-1">{cnam}</div>}
        <div className="w-full h-px bg-white/[0.04] my-3" />
        <div className="text-[10px] text-[#525252] font-mono tracking-widest uppercase mb-1">Destination</div>
        <div className="text-white font-mono">{destination}</div>
      </div>

      {/* Waveform */}
      <div className="flex items-end gap-[3px] h-10">
        {[1,2,3,4,5,6,5,4,3,4,5,6,5,4,3,2,1].map((h,i) => (
          <motion.div
            key={i}
            animate={{
              height: [`${12 + h * 4}px`, `${16 + h * 6}px`, `${12 + h * 4}px`],
            }}
            transition={{
              duration: 1.2,
              repeat: Infinity,
              delay: i * 0.08,
              ease: "easeInOut",
            }}
            className="w-1 bg-gradient-to-t from-red-600 to-red-400 rounded-full"
          />
        ))}
      </div>

      {/* DTMF */}
      <div className="card w-full max-w-sm">
        <div className="text-[10px] text-[#525252] font-mono tracking-widest uppercase mb-3 text-center">DTMF Keypad</div>
        {dtmf && (
          <motion.div
            key={dtmf}
            initial={{ scale: 1.05 }}
            animate={{ scale: 1 }}
            className="text-center mb-3"
          >
            {dtmf.split("").map((d,i) => (
              <motion.span
                key={i}
                initial={{ opacity: 0, y: -8 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: i * 0.03, type: "spring", stiffness: 400 }}
                className="inline-block mx-0.5 text-green-400 font-mono text-lg font-semibold"
              >
                {d}
              </motion.span>
            ))}
          </motion.div>
        )}
        <div className="flex flex-col items-center gap-2">
          {[["1","2","3"],["4","5","6"],["7","8","9"],["*","0","#"]].map((row, ri) => (
            <div key={ri} className="flex gap-3">
              {row.map(k => (
                <motion.button
                  key={k}
                  whileTap={{ scale: 0.85 }}
                  className="dtmf-btn !w-14 !h-14 !text-base"
                  onClick={() => handleDTMF(k)}
                >
                  {k}
                </motion.button>
              ))}
            </div>
          ))}
        </div>
      </div>

      {/* Controls */}
      <div className="flex gap-3 w-full max-w-sm">
        <motion.button
          whileTap={{ scale: 0.95 }}
          onClick={() => { setMuted(!muted); api.muteCall(callId, !muted).catch(()=>{}); }}
          className={`btn-secondary flex-1 py-3 ${muted ? "!bg-red-500/10 !border-red-500/30 !text-red-400" : ""}`}
        >
          {muted ? "Unmute" : "Mute"}
        </motion.button>
        <motion.button
          whileTap={{ scale: 0.95 }}
          onClick={() => setDtmf("")}
          className="btn-secondary flex-1 py-3"
        >
          Clear
        </motion.button>
      </div>

      <motion.button
        whileTap={{ scale: 0.97 }}
        onClick={handleHangup}
        className="w-full max-w-sm py-4 rounded-xl bg-gradient-to-r from-red-600 to-red-700 text-white font-bold text-base
          hover:from-red-500 hover:to-red-600 transition-all shadow-lg shadow-red-500/20"
      >
        End Call
      </motion.button>
    </motion.div>
  );
}
