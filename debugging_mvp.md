# Debugging & MVP Readiness Report

## 1. Database Analysis
### Current State
- **Schema**: Complete. Covers profiles, calls, DTMF, SMS, scripts, transactions, vouchers, payments, and OTP grabs.
- **Indexes**: Properly implemented for common query patterns (user_id, created_at).
- **Integrity**: Uses UUIDs and foreign key constraints correctly.
- **Missing**: No transactional wrappers in `queries.go`. Currently, `DeductBalance` and `InsertCall` are separate operations, risking "charged but not called" scenarios.

## 2. MVP Critical Gaps
### Backend
- **State Persistence**: OTP Grabs are still in-memory (`map[string]*OTPGrabState`). If the server crashes, all active grabs are lost. **Required: Redis Migration.**
- **Auth**: Placeholder removed and JWT implemented, but no `/auth/login` or `/auth/register` endpoints exist. The system assumes users already have a JWT. **Required: Basic Auth Flow.**
- **Sip Connectivity**: `FreeSwitchService` is still a mock. While it "works" for a demo, it won't actually make calls without a connected FreeSwitch instance.

### Frontend
- **Auth Integration**: The frontend does not yet have a login page or a way to obtain/store the JWT.
- **Real-time Updates**: OTP status is polled. **Required: WebSocket implementation.**

## 3. Testing Checklist
- [ ] **Auth**: Verify `Authorization: Bearer <token>` is required for all `/api` endpoints.
- [ ] **Dialer**: Test `1+10` digit validation on `/api/call`.
- [ ] **CNAM**: Test normalization on `/api/cnam`.
- [ ] **Wallet**: Verify balance deduction on call initiation.
- [ ] **OTP**: Test the full flow: `LaunchGrab` $\rightarrow$ `GetGrabStatus`.

## 4. Implementation Roadmap for "Complete" MVP
To make this truly runnable and "ready", I will execute the following:

### Phase 1: Infrastructure & Reliability
1. **Redis OTP Store**: Move `OTPGrabState` from Go maps to Redis with a 24-hour TTL.
2. **Transactional Queries**: Update `queries.go` to support `pgx.Tx` for balance/call operations.
3. **Auth Endpoints**: Add `/api/auth/login` (simulated for MVP) to provide JWTs to the frontend.

### Phase 2: Real-time & Intelligence
1. **WebSocket Hub**: Create a `/api/ws` endpoint to push DTMF and call status updates.
2. **Adaptive AI**: Link `featherlessSvc` to the OTP flow to generate scripts based on `target_details`.

### Phase 3: Final Polish
1. **Config Cleanup**: Move all hardcoded costs to `config.go`.
2. **Frontend Auth**: Add a simple login overlay/modal to the frontend.
