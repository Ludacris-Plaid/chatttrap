package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"hushcircuits/api/internal/models"
)

type NowPaymentsService struct {
	apiKey     string
	ipnSecret  string
	callbackURL string
	client     *http.Client
}

func NewNowPaymentsService(apiKey, ipnSecret, callbackURL string) *NowPaymentsService {
	return &NowPaymentsService{
		apiKey:      apiKey,
		ipnSecret:   ipnSecret,
		callbackURL: callbackURL,
		client:      &http.Client{},
	}
}

func (s *NowPaymentsService) CreatePayment(userID string, req *models.CreatePaymentRequest) (*models.Payment, error) {
	if s.apiKey == "" {
		return s.mockPayment(userID, req), nil
	}

	payload := map[string]any{
		"price_amount":     req.Amount,
		"price_currency":   "USD",
		"pay_currency":     strings.ToUpper(req.Currency),
		"ipn_callback_url": s.callbackURL,
		"order_id":         userID,
		"order_description": fmt.Sprintf("HushCircuits deposit - %s", userID),
	}

	data, _ := json.Marshal(payload)
	httpReq, _ := http.NewRequest("POST", "https://api.nowpayments.io/v1/invoice", strings.NewReader(string(data)))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", s.apiKey)

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("nowpayments request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("nowpayments returned %d: %s", resp.StatusCode, string(body))
	}

	var npResp struct {
		PaymentID  string  `json:"payment_id"`
		PayAddress string  `json:"pay_address"`
		PayAmount  float64 `json:"pay_amount"`
		PayCurrency string `json:"pay_currency"`
		PriceAmount float64 `json:"price_amount"`
	}
	json.Unmarshal(body, &npResp)

	tokens := int(req.Amount / 0.50)
	return &models.Payment{
		UserID:    userID,
		PaymentID: npResp.PaymentID,
		Currency:  npResp.PayCurrency,
		Amount:    npResp.PriceAmount,
		Tokens:    tokens,
		Status:    "pending",
		PayAddress: npResp.PayAddress,
		PayAmount:  npResp.PayAmount,
	}, nil
}

func (s *NowPaymentsService) VerifyWebhook(body []byte, signature string) bool {
	if s.ipnSecret == "" {
		slog.Warn("webhook verification skipped: ipnSecret not configured")
		return false
	}
	mac := hmac.New(sha256.New, []byte(s.ipnSecret))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}

func (s *NowPaymentsService) mockPayment(userID string, req *models.CreatePaymentRequest) *models.Payment {
	tokens := int(req.Amount / 0.50)
	return &models.Payment{
		UserID:    userID,
		PaymentID: fmt.Sprintf("mock_np_%s", userID),
		Currency:  strings.ToUpper(req.Currency),
		Amount:    req.Amount,
		Tokens:    tokens,
		Status:    "waiting",
		PayAddress: "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa",
		PayAmount:  req.Amount * 0.00005,
	}
}
