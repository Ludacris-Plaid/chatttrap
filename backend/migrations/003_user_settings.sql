-- User settings
CREATE TABLE IF NOT EXISTS public.user_settings (
  user_id TEXT PRIMARY KEY REFERENCES public.profiles(id) ON DELETE CASCADE,
  country_code TEXT DEFAULT '+1',
  webhook_url TEXT DEFAULT '',
  ringtone BOOLEAN DEFAULT true,
  vibrate BOOLEAN DEFAULT true,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);
