package models

import "time"

type Profile struct {
	ID                string    `json:"id"`
	Email             string    `json:"email,omitempty"`
	Phone             string    `json:"phone"`
	Balance           float64   `json:"balance"`
	IsVIP             bool      `json:"is_vip"`
	VIPExpiresAt      *time.Time `json:"vip_expires_at,omitempty"`
	TokensUsed        int64     `json:"tokens_used"`
	TotalCalls        int64     `json:"total_calls"`
	ReferralCode      string    `json:"referral_code"`
	ReferredBy        *string   `json:"referred_by,omitempty"`
	FirstDeposit      bool      `json:"first_deposit"`
	OnboardingCompleted bool    `json:"onboarding_completed"`
	CreatedAt         time.Time `json:"created_at"`
}

type Call struct {
	ID                string    `json:"id"`
	UserID            string    `json:"user_id"`
	SpoofedCallerID   string    `json:"spoofed_caller_id"`
	SpoofedName       string    `json:"spoofed_name"`
	DestinationNumber string    `json:"destination_number"`
	Status            string    `json:"status"`
	DurationSeconds   int       `json:"duration_seconds"`
	TokensCost        int       `json:"tokens_cost"`
	CostUSD           float64   `json:"cost_usd"`
	CNAMResult        string    `json:"cnam_result"`
	TrunkUsed         string    `json:"trunk_used"`
	RecordingURL      string    `json:"recording_url,omitempty"`
	DTMFCaptured      string    `json:"dtmf_captured"`
	HangupCause       string    `json:"hangup_cause"`
	CreatedAt         time.Time `json:"created_at"`
	EndedAt           *time.Time `json:"ended_at,omitempty"`
}

type DTMFLog struct {
	ID          string    `json:"id"`
	CallID      string    `json:"call_id"`
	UserID      string    `json:"user_id"`
	Digit       string    `json:"digit"`
	TimestampMs int       `json:"timestamp_ms"`
	CreatedAt   time.Time `json:"created_at"`
}

type SMSCampaign struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	SenderID  string    `json:"sender_id"`
	Content   string    `json:"content"`
	Targets   int       `json:"targets"`
	SentCount int       `json:"sent_count"`
	Status    string    `json:"status"`
	CostTokens int      `json:"cost_tokens"`
	CreatedAt time.Time `json:"created_at"`
}

type Script struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	Title         string    `json:"title"`
	TargetName    string    `json:"target_name"`
	TargetAge     int       `json:"target_age"`
	TargetService string    `json:"target_service"`
	TargetDetails string    `json:"target_details"`
	Goal          string    `json:"goal"`
	ScriptType    string    `json:"script_type"`
	Content       string    `json:"content"`
	TokensCost    int       `json:"tokens_cost"`
	IsLibrary     bool      `json:"is_library"`
	CreatedAt     time.Time `json:"created_at"`
}

type Transaction struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Type        string    `json:"type"`
	Amount      float64   `json:"amount"`
	Tokens      int       `json:"tokens"`
	Currency    string    `json:"currency"`
	Status      string    `json:"status"`
	Reference   string    `json:"reference,omitempty"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type Voucher struct {
	ID        string     `json:"id"`
	Code      string     `json:"code"`
	Tokens    int        `json:"tokens"`
	IsUsed    bool       `json:"is_used"`
	UsedBy    *string    `json:"used_by,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
}

type Payment struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	PaymentID   string    `json:"payment_id"`
	Currency    string    `json:"currency"`
	Amount      float64   `json:"amount"`
	Tokens      int       `json:"tokens"`
	Status      string    `json:"status"`
	PayAddress  string    `json:"pay_address"`
	PayAmount   float64   `json:"pay_amount"`
	TxID        string    `json:"txid,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	ConfirmedAt *time.Time `json:"confirmed_at,omitempty"`
}

// Request types
type OriginateCallRequest struct {
	SpoofedCallerID   string `json:"spoofed_caller_id"`
	SpoofedName       string `json:"spoofed_name"`
	DestinationNumber string `json:"destination_number"`
}

type SendSMSRequest struct {
	PhoneNumber string `json:"phone_number"`
	Content     string `json:"content"`
	SenderID    string `json:"sender_id"`
	Campaign    bool   `json:"campaign,omitempty"`
	Targets     string `json:"targets,omitempty"`
}

type GenerateScriptRequest struct {
	TargetName    string `json:"target_name"`
	TargetAge     int    `json:"target_age"`
	TargetService string `json:"target_service"`
	TargetDetails string `json:"target_details"`
	Goal          string `json:"goal"`
	ScriptType    string `json:"script_type"`
}

type TopUpRequest struct {
	Amount float64 `json:"amount"`
}

type RedeemVoucherRequest struct {
	Code string `json:"code"`
}

type CreatePaymentRequest struct {
	Currency string  `json:"currency"`
	Amount   float64 `json:"amount"`
}

type NowPaymentsWebhook struct {
	PaymentID  string  `json:"payment_id"`
	PaymentStatus string `json:"payment_status"`
	PayAddress string  `json:"pay_address"`
	PriceAmount float64 `json:"price_amount"`
	PriceCurrency string `json:"price_currency"`
	PayAmount   float64 `json:"pay_amount"`
	ActuallyPaid float64 `json:"actually_paid"`
	PayCurrency  string  `json:"pay_currency"`
	OrderID     string  `json:"order_id"`
	OrderDescription string `json:"order_description"`
	TxID        string  `json:"txid"`
}

type CNAMResponse struct {
	Number string `json:"number"`
	Name   string `json:"name"`
	Error  string `json:"error,omitempty"`
}
