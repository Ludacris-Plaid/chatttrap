"use client";

import { useState, useEffect } from "react";
import { motion } from "framer-motion";
import { api } from "@/lib/api";
import { useToast } from "@/lib/toast";

export default function Settings() {
  const [countryCode, setCountryCode] = useState("+1");
  const [webhook, setWebhook] = useState("");
  const [ringtone, setRingtone] = useState(true);
  const [vibrate, setVibrate] = useState(true);
  const [saved, setSaved] = useState(false);
  const [loading, setLoading] = useState(true);
  const { toast } = useToast();

  useEffect(() => {
    api.getSettings()
      .then(r => {
        setCountryCode(r.country_code || "+1");
        setWebhook(r.webhook_url || "");
        setRingtone(r.ringtone ?? true);
        setVibrate(r.vibrate ?? true);
      })
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  const handleSave = async () => {
    try {
      await api.updateSettings({
        country_code: countryCode,
        webhook_url: webhook,
        ringtone,
        vibrate,
      });
      setSaved(true);
      toast("Settings saved", "success");
      setTimeout(() => setSaved(false), 2000);
    } catch (e: any) {
      toast(e.message || "Failed to save", "error");
    }
  };

  return (
    <div className="flex flex-col gap-5 pt-2">
      <motion.div
        initial={{ opacity: 0, y: -10 }}
        animate={{ opacity: 1, y: 0 }}
        className="text-center"
      >
        <span className="text-2xl font-bold text-white">Settings</span>
      </motion.div>

      {loading ? (
        <div className="space-y-4">
          {[1,2,3].map(i => <div key={i} className="skeleton h-16 rounded-xl" />)}
        </div>
      ) : (
        <>
          {/* General */}
          <motion.div
            initial={{ opacity: 0, x: -20 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ delay: 0.05 }}
            className="card space-y-4"
          >
            <div className="text-[10px] text-[#525252] font-mono tracking-widest uppercase">General</div>
            <div>
              <label className="text-[#525252] text-[10px] font-mono tracking-widest uppercase block mb-2">Default Country Code</label>
              <select
                className="input-field font-mono"
                value={countryCode}
                onChange={e => setCountryCode(e.target.value)}
              >
                <option value="+1">+1 (US/Canada)</option>
                <option value="+44">+44 (United Kingdom)</option>
                <option value="+61">+61 (Australia)</option>
                <option value="+49">+49 (Germany)</option>
                <option value="+33">+33 (France)</option>
                <option value="+81">+81 (Japan)</option>
              </select>
            </div>
            <div>
              <label className="text-[#525252] text-[10px] font-mono tracking-widest uppercase block mb-2">Webhook URL</label>
              <input
                className="input-field font-mono"
                placeholder="https://your-server.com/webhook"
                value={webhook}
                onChange={e => setWebhook(e.target.value)}
              />
              <div className="text-[#404040] text-[10px] font-mono mt-1">POST alerts for calls, OTP captures, and DTMF events</div>
            </div>
          </motion.div>

          {/* Audio */}
          <motion.div
            initial={{ opacity: 0, x: 20 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ delay: 0.1 }}
            className="card space-y-3"
          >
            <div className="text-[10px] text-[#525252] font-mono tracking-widest uppercase">Audio & Haptics</div>
            <label className="flex items-center justify-between cursor-pointer group">
              <span className="text-xs text-[#a3a3a3] font-mono group-hover:text-white transition-colors">
                Play ringtone during outbound calls
              </span>
              <motion.button
                whileTap={{ scale: 0.9 }}
                onClick={() => setRingtone(!ringtone)}
                className={`relative w-10 h-5 rounded-full transition-colors ${ringtone ? "bg-red-600" : "bg-white/10"}`}
              >
                <motion.div
                  animate={{ x: ringtone ? 20 : 2 }}
                  transition={{ type: "spring", stiffness: 500, damping: 30 }}
                  className="absolute top-0.5 w-4 h-4 rounded-full bg-white"
                />
              </motion.button>
            </label>
            <label className="flex items-center justify-between cursor-pointer group">
              <span className="text-xs text-[#a3a3a3] font-mono group-hover:text-white transition-colors">
                Vibrate on DTMF capture
              </span>
              <motion.button
                whileTap={{ scale: 0.9 }}
                onClick={() => setVibrate(!vibrate)}
                className={`relative w-10 h-5 rounded-full transition-colors ${vibrate ? "bg-red-600" : "bg-white/10"}`}
              >
                <motion.div
                  animate={{ x: vibrate ? 20 : 2 }}
                  transition={{ type: "spring", stiffness: 500, damping: 30 }}
                  className="absolute top-0.5 w-4 h-4 rounded-full bg-white"
                />
              </motion.button>
            </label>
          </motion.div>

          {/* Save */}
          <motion.button
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.15 }}
            whileTap={{ scale: 0.97 }}
            onClick={handleSave}
            className="w-full py-4 rounded-xl bg-gradient-to-r from-red-600 to-red-700 text-white font-bold text-sm
              hover:from-red-500 hover:to-red-600 transition-all shadow-lg shadow-red-500/20
              flex items-center justify-center gap-2"
          >
            {saved ? (
              <>
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
                  <polyline points="20 6 9 17 4 12"/>
                </svg>
                SAVED
              </>
            ) : (
              <>
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5">
                  <path d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v11a2 2 0 0 1-2 2z"/>
                  <polyline points="17 21 17 13 7 13 7 21"/>
                  <polyline points="7 3 7 8 15 8"/>
                </svg>
                SAVE SETTINGS
              </>
            )}
          </motion.button>
        </>
      )}
    </div>
  );
}
