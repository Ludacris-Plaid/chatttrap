# HushCircuits Pro: MVP Sprint Checklist

## 🎯 Mission
Build a production-ready, multi-threaded, Redis-backed **AI Voice Caller** that dominates lead response times with adaptive, conversational agents. Users speak naturally to their own AI caller, and it responds in any voice/age/sex/accent targeting their leads.

---

## 🔥 Digital Issues We're Solving

### The "Missed Opportunity" Plague
- **Lead Response Window**: 5 seconds to 5 minutes. After that, leads cold.
- **Current State**: Humans dial, hesitate, stutter, check phones mid-conversation.
- **The Gap**: AI exists but sounds robotic, monotonous, or limited to pre-recorded scripts.

### The "Generic AI" Antagonist
- **Problem**: Most "AI callers" are just text-to-speech reading templates.
- **Consequence**: Leads detect the pattern within 10 seconds. Disengage.
- **Our Edge**: Natural language interface + adaptive scripting + 4x voice profiles.

### The "Context Collapse" Crisis
- **Issue**: AI doesn't know the lead's age, accent, or recent behavior.
- **Result**: Awkward, unnatural conversations that feel scripted.
- **Fix**: Dynamic profile generation + real-time context injection.

### The "State Chaos" Nightmare
- **Reality**: Redis maps, in-memory flags, race conditions across servers.
- **Impact**: Lost calls, double-charges, stale DTMF states.
- **Solution**: Redis-backed `OTPGrabState` with 24h TTL + atomic `pgx` transactions.

### The "Phone Format" Friction
- **Pain**: `1234567890`, `+1-234-567-890`, `234567890` all break things.
- **Fix**: Enforced `1XXXXXXXXXX` normalization + `phone.go` validation.

### The "Token Economy" Black Hole
- **Risk**: `DeductBalance` before `InsertCall` = "charged but not called" errors.
- **Remedy**: `ExecTx` wrappers + hardened transactional helpers.

---

## 🏗️ Architecture Shift: Supabase-Ready

### Current Stack → Supabase Migration
| Current | Supabase Replacement |
|---------|---------------------|
| `pgxpool` (Postgres) | `pgxpool` (Postgres) |
| `redis.NewClient` | `redis.NewClient` (keep Redis for state) |
| `golang-jwt/jwt` | `golang-jwt/jwt` (keep JWT auth) |
| `github.com/gofiber/fiber` | `github.com/gofiber/fiber` |

### Supabase-Specific Changes
1. **Auth**: Replace `profiles.email` + JWT login with Supabase Auth hooks.
2. **Database**: Keep `profiles` table (standalone), add `supabase_auth.users` sync.
3. **Storage**: `genSMS_API_KEY` → Supabase Storage for call recordings.
4. **Edge Functions**: Migrate Featherless prompt handling to Supabase Edge Functions.

### Migration Checklist
- [ ] Add `auth.users` → `profiles` sync trigger
- [ ] Replace `godotenv` with `supabase/serverless` secrets
- [ ] Add `supabase_url`, `supabase_anon_key`, `supabase_service_key` to `Config`
- [ ] Update `database.Connect()` to use Supabase connection pool
- [ ] Add `pgxpool` health check endpoint

---

## 📞 The Calling Component: AI Voice Caller Architecture

### Core Requirements
1. **Multi-Voice Profiles**: 4 voices (Age/ Sex/ Accent variants)
2. **Natural Language Interface**: Users speak normally, AI adapts.
3. **Adaptive Scripting**: Generate scripts based on real-time target data.
4. **Stateless Design**: Any number can call any number.
5. **Hardcore Resilience**: "If they try, I'll stab them all up."

### Component Breakdown

#### 1. **Voice Profile Service** (`backend/internal/services/voice_profiles.go`)
- **Purpose**: Manage 4x voice profiles (Male/Female, Young/Mature, US/UK/AU/IN accents).
- **Storage**: Redis-backed with per-user voice selection.
- **API**:
  ```go
  type VoiceProfile struct {
      ID          string
      AgeGroup    string // "young_adult", "mature_adult"
      Sex         string // "male", "female"
      Accent      string // "us", "uk", "au", "in"
      VoiceModel  string // "clue_con_v1", "clue_con_v2"
  }
  ```

#### 2. **Natural Language Interface** (`backend/internal/handlers/nli.go`)
- **Purpose**: Transcribe user speech → generate AI response.
- **Stack**:
  - Speech-to-Text: `deepgram.com` or `google-cloud-speech`.
  - NLP: `featherless.ai` (captain-eris-violet-12b) for adaptive responses.
  - Text-to-Speech: `freeSwitch` (`ClueCon` trunk, localhost:8021).
- **API**:
  ```go
  type NLISession struct {
      SessionID string
      VoiceProfile *VoiceProfile
      Script *Script
  }
  ```

#### 3. **Script Generator** (`backend/internal/handlers/script.go`)
- **Purpose**: Generate/Adapt scripts based on target profile.
- **Inputs**: `target_name`, `target_age`, `target_service`, `target_details`.
- **Outputs**: Dynamic scripts (fallback script for demo mode).
- **API**:
  ```go
  type ScriptRequest struct {
      TargetName  string
      TargetAge   int
      TargetService string
      TargetDetails string
      Goal         string
  }
  ```

#### 4. **Call Origination** (`backend/internal/handlers/dialer.go`)
- **Purpose**: FreeSwitch → `ClueCon` trunk → destination.
- **State**: Redis-backed `OTPGrabState` with 24h TTL.
- **API**:
  ```go
  type CallRequest struct {
      Destination string
      VoiceProfileID string
      ScriptID     string // Optional, fallback if missing
      NLI_SessionID string
  }
  ```

#### 5. **DTMF Capture** (`backend/internal/handlers/dtmf.go`)
- **Purpose**: Capture real-time DTMF from `ClueCon` trunk.
- **State**: Redis-backed with 24h TTL.
- **API**:
  ```go
  type DTMFRequest struct {
      CallID string
      Digit string
      TimestampMs int
  }
  ```

---

## 🔄 Data Flow: From User Speech to AI Response

### 1. User Speaks
```
User: "Hey Sarah, I saw you were interested in our premium plan..."
     ↓
Speech-to-Text (Deepgram/Google)
     ↓
"Hey Sarah, I saw you were interested in our premium plan..."
```

### 2. NLI Response Generation
```
NLI Service → Featherless.ai
     ↓
"Sure! Tell me more about what you'd like to know..."
     ↓
Text-to-Speech (FreeSwitch/ClueCon)
     ↓
Audio Stream → Destination Phone
```

### 3. Script Adaptation
```
Target: "Sarah, 28, Tech Startup, US"
     ↓
Script Generator → "Hey Sarah! I saw you were interested..."
     ↓
Dynamic Context Injection:
- Age: 28 → "young_adult"
- Service: "Tech Startup" → "tech_industry"
- Details: "Premium Plan" → "premium_plan_interest"
```

---

## 🧪 MVP Testing Checklist

### Phase 1: Core Voice Profiles
- [ ] 4x voice profiles load correctly from Redis.
- [ ] User selects voice profile via UI dropdown.
- [ ] Voice profile persists across sessions.

### Phase 2: Natural Language Interface
- [ ] User speaks to AI → AI responds naturally.
- [ ] AI adapts to target profile (age, sex, accent).
- [ ] AI handles interruptions gracefully.

### Phase 3: Script Generation
- [ ] Script generates based on target profile.
- [ ] Fallback script works for demo mode.
- [ ] Script persists in Redis/Postgres.

### Phase 4: Call Origination
- [ ] FreeSwitch → `ClueCon` trunk → destination.
- [ ] DTMF captured and streamed in real-time.
- [ ] Redis state cleanup after 24h.

### Phase 5: Supabase Migration
- [ ] `profiles` table syncs with `auth.users`.
- [ ] `godotenv` replaced with Supabase secrets.
- [ ] `pgxpool` health check endpoint.

---

## 🚀 Hardened Resilience: "Stab Them All Up"

### 1. **Stateless Design**
- Any number can call any number.
- Redis-backed state with 24h TTL.
- No persistent sessions (stateless).

### 2. **Atomic Transactions**
- `DeductBalanceTx` + `InsertCallTx` prevent race conditions.
- `ExecTx` wrappers ensure consistency.

### 3. **Fallback Flows**
- **Featherless API Down**: `fallbackScript()` returns hardcoded script.
- **FreeSwitch Down**: Retry loop with exponential backoff.
- **Redis Down**: In-memory maps with graceful degradation.

### 4. **Error Handling**
- **Speech-to-Text Failure**: Auto-retry 3x, then fallback to TTS.
- **Text-to-Speech Failure**: Stream previous audio buffer.
- **API Failure**: Log to `logs/` and alert admin.

---

## 📦 Final Build Commands

### Backend
```bash
cd /home/dysthemix/projects/ultimate_spoof/backend
go mod tidy
go build -o api ./cmd/api
```

### Frontend
```bash
cd /home/dysthemix/projects/ultimate_spoof/frontend
npm run dev
```

### Database
```bash
cd /home/dysthemix/projects/ultimate_spoof/backend
psql $DATABASE_URL -f migrations/001_schema.sql
```

### Redis
```bash
redis-cli ping
```

---

## 🎯 Next Steps

1. **Voice Profile Service**: Create `backend/internal/services/voice_profiles.go`.
2. **NLI Service**: Implement `backend/internal/services/nli.go`.
3. **Script Generator**: Extend `backend/internal/handlers/script.go`.
4. **Supabase Migration**: Update `backend/internal/config/config.go` and `backend/internal/database/database.go`.
5. **Frontend Integration**: Update `frontend/src/components/OTPGrab.tsx` to include voice profile selection.

---

## 🧠 Key Design Principles

1. **Stateless**: Any number can call any number.
2. **Atomic**: `DeductBalanceTx` + `InsertCallTx` prevent race conditions.
3. **Adaptive**: Scripts adapt to real-time target data.
4. **Resilient**: Fallback flows for every failure mode.
5. **Hardcore**: "If they try, I'll stab them all up."

---

## 📝 Quick Reference

- **Phone Format**: `1XXXXXXXXXX` (11 digits, 1 prefix).
- **Redis TTL**: 24h for active grabs.
- **JWT Expiry**: 24h.
- **Default Costs**: OTP Grab = $5.00 + 10 tokens; Default Call = $2.50 + 5 tokens.
- **Fallback AI**: `fallbackScript()` returns hardcoded, high-quality script if API key is missing.
- **FreeSwitch**: Uses `ClueCon` trunk at `localhost:8021`.
- **Token Service**: Uses Redis with `SetNX` locking for 5s.
- **Database**: Uses `pgx` with transactional helpers.

---

## 🔗 Relevant Files

- `backend/internal/services/voice_profiles.go`: New voice profile management.
- `backend/internal/services/nli.go`: New NLI (Natural Language Interface) service.
- `backend/internal/handlers/script.go`: Script generation with dynamic context.
- `backend/internal/config/config.go`: Add `supabase_url`, `supabase_anon_key`, `supabase_service_key`.
- `backend/internal/database/database.go`: Update for Supabase connection pool.
- `frontend/src/components/OTPGrab.tsx`: Add voice profile selection.
- `backend/cmd/api/main.go`: Register new routes and services.
