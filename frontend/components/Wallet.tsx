"use client";

import { useState, useEffect } from "react";
import { motion } from "framer-motion";
import { api } from "@/lib/api";
import { useToast } from "@/lib/toast";

export default function Wallet() {
  const [balance, setBalance] = useState(0);
  const [isVIP, setIsVIP] = useState(false);
  const [txns, setTxns] = useState<any[]>([]);
  const [code, setCode] = useState("");
  const [loading, setLoading] = useState(false);
  const [currency, setCurrency] = useState("BTC");
  const { toast } = useToast();

  const refreshBalance = () => {
    api.getBalance().then(r => { setBalance(r.balance); setIsVIP(r.is_vip); }).catch(()=>{});
  };

  useEffect(() => {
    refreshBalance();
    api.getTransactions().then(r => setTxns(r.transactions || [])).catch(()=>{});
  }, []);

  const handleVoucher = async () => {
    if (!code) return;
    setLoading(true);
    try {
      const r = await api.redeemVoucher(code);
      refreshBalance();
      api.getTransactions().then(r => setTxns(r.transactions || [])).catch(()=>{});
      setCode("");
      toast(`Redeemed ${r.tokens} tokens`, "success");
    } catch (e: any) { toast(e.message, "error"); }
    setLoading(false);
  };

  const handleVIP = async () => {
    try {
      await api.upgradeVIP();
      setIsVIP(true);
      toast("VIP activated for 7 days", "success");
    } catch (e: any) { toast(e.message, "error"); }
  };

  return (
    <div className="flex flex-col gap-5 pt-2">
      {/* Balance card */}
      <motion.div
        initial={{ opacity: 0, y: -20, scale: 0.95 }}
        animate={{ opacity: 1, y: 0, scale: 1 }}
        className="relative overflow-hidden rounded-2xl p-6 text-center"
        style={{
          background: "linear-gradient(135deg, rgba(239,68,68,0.12) 0%, rgba(168,85,247,0.08) 50%, rgba(239,68,68,0.04) 100%)",
          border: "1px solid rgba(239,68,68,0.15)",
        }}
      >
        <div className="absolute inset-0 bg-gradient-to-br from-red-500/5 to-purple-500/5" />
        <div className="relative z-10">
          <div className="text-[10px] text-[#525252] font-mono tracking-widest uppercase mb-2">Balance</div>
          <motion.div
            key={balance}
            initial={{ scale: 1.2 }}
            animate={{ scale: 1 }}
            transition={{ type: "spring", stiffness: 300, damping: 20 }}
            className="text-5xl font-mono font-bold text-white"
          >
            ${balance.toFixed(2)}
          </motion.div>
          {isVIP && (
            <motion.div
              initial={{ opacity: 0, y: 5 }}
              animate={{ opacity: 1, y: 0 }}
              className="inline-flex items-center gap-1.5 text-yellow-400 text-xs font-mono mt-3 px-3 py-1 rounded-full
                bg-yellow-500/10 border border-yellow-500/20"
            >
              <svg width="12" height="12" viewBox="0 0 24 24" fill="currentColor">
                <path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"/>
              </svg>
              VIP ACTIVE
            </motion.div>
          )}
        </div>
      </motion.div>

      {/* Buy tokens */}
      <motion.div
        initial={{ opacity: 0, y: 10 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.05 }}
        className="card"
      >
        <div className="text-[10px] text-[#525252] font-mono tracking-widest uppercase mb-3">Buy Tokens</div>
        <div className="flex gap-2 mb-3">
          {["BTC", "ETH", "USDT", "LTC"].map(c => (
            <button
              key={c}
              onClick={() => setCurrency(c)}
              className={`px-3 py-1 rounded-lg text-[10px] font-mono uppercase transition-all ${
                currency === c
                  ? "bg-red-600/20 text-red-400 border border-red-500/30"
                  : "bg-white/[0.03] text-[#525252] border border-white/[0.04] hover:text-[#a3a3a3]"
              }`}
            >
              {c}
            </button>
          ))}
        </div>
        <div className="grid grid-cols-4 gap-2">
          {[25, 50, 100, 250].map((a, i) => (
            <motion.button
              key={a}
              whileTap={{ scale: 0.95 }}
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.1 + i * 0.03 }}
              onClick={async () => {
                try {
                  const r = await api.createPayment({ currency, amount: a });
                  toast(`Send ${r.pay_amount} ${currency} to ${r.pay_address}`, "success");
                } catch (e: any) { toast(e.message, "error"); }
              }}
              className="py-4 rounded-xl bg-white/[0.03] border border-white/[0.04] text-white font-mono text-sm font-bold
                hover:border-red-500/30 hover:bg-red-500/5 transition-all"
            >
              ${a}
            </motion.button>
          ))}
        </div>
      </motion.div>

      {/* VIP upgrade */}
      {!isVIP && (
        <motion.div
          initial={{ opacity: 0, y: 10 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.15 }}
          className="relative overflow-hidden rounded-xl p-4"
          style={{
            background: "linear-gradient(135deg, rgba(168,85,247,0.1) 0%, rgba(239,68,68,0.05) 100%)",
            border: "1px solid rgba(168,85,247,0.2)",
          }}
        >
          <div className="text-[10px] text-purple-400 font-mono tracking-widest uppercase mb-1">VIP — $250 FOR 7 DAYS UNLIMITED</div>
          <div className="text-[#525252] text-[10px] font-mono mb-3">Unlimited calls, priority routing, dedicated trunks</div>
          <motion.button
            whileTap={{ scale: 0.97 }}
            onClick={handleVIP}
            className="w-full py-3 rounded-xl bg-gradient-to-r from-purple-600 to-purple-700 text-white text-sm font-bold
              hover:from-purple-500 hover:to-purple-600 transition-all shadow-lg shadow-purple-500/20"
          >
            UPGRADE TO VIP
          </motion.button>
        </motion.div>
      )}

      {/* Voucher */}
      <motion.div
        initial={{ opacity: 0, y: 10 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.2 }}
        className="card"
      >
        <div className="text-[10px] text-[#525252] font-mono tracking-widest uppercase mb-3">Redeem Voucher</div>
        <div className="flex gap-2">
          <input
            className="input-field flex-1 font-mono"
            placeholder="HUSH-XXXXXX"
            value={code}
            onChange={e => setCode(e.target.value)}
          />
          <motion.button
            whileTap={{ scale: 0.95 }}
            onClick={handleVoucher}
            disabled={loading || !code}
            className="btn-primary text-sm px-5 py-2 disabled:opacity-30"
          >
            {loading ? <span className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin inline-block" /> : "REDEEM"}
          </motion.button>
        </div>
      </motion.div>

      {/* Transactions */}
      <motion.div
        initial={{ opacity: 0, y: 10 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.25 }}
        className="card"
      >
        <div className="text-[10px] text-[#525252] font-mono tracking-widest uppercase mb-3">Transactions</div>
        {txns.length === 0 ? (
          <div className="text-center py-6">
            <div className="text-[#262626] text-2xl mb-2">💰</div>
            <div className="text-[#404040] text-xs font-mono">No transactions yet</div>
          </div>
        ) : (
          <div className="space-y-2 max-h-48 overflow-y-auto">
            {txns.map((t: any, i: number) => (
              <motion.div
                key={t.id}
                initial={{ opacity: 0, x: -10 }}
                animate={{ opacity: 1, x: 0 }}
                transition={{ delay: i * 0.02 }}
                className="flex items-center justify-between text-xs bg-white/[0.02] rounded-lg p-3 border border-white/[0.03]"
              >
                <div>
                  <div className="text-[#a3a3a3] font-mono">{t.description || t.type}</div>
                  <div className="text-[#404040] text-[10px] font-mono mt-0.5">{new Date(t.created_at).toLocaleString()}</div>
                </div>
                <div className={`font-mono font-bold ${t.amount >= 0 ? "text-green-400" : "text-red-400"}`}>
                  {t.amount >= 0 ? "+" : ""}${Math.abs(t.amount).toFixed(2)}
                </div>
              </motion.div>
            ))}
          </div>
        )}
      </motion.div>
    </div>
  );
}
