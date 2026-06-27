package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type GENSMSService struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

func NewGENSMSService(apiKey string) *GENSMSService {
	return &GENSMSService{
		apiKey:  apiKey,
		baseURL: "https://api.gensms.com",
		client:  &http.Client{},
	}
}

func (s *GENSMSService) Send(phoneNumber, content, senderID string) (string, error) {
	if s.apiKey == "" {
		return mockMsgID(phoneNumber), nil
	}

	body := map[string]string{
		"to":   phoneNumber,
		"text": content,
	}
	if senderID != "" {
		body["from"] = senderID
	}

	data, _ := json.Marshal(body)
	httpReq, _ := http.NewRequest("POST", s.baseURL+"/v1/send", bytes.NewReader(data))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("gensms request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("gensms returned %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Status    string `json:"status"`
		MessageID string `json:"message_id"`
		Timestamp string `json:"timestamp"`
	}
	json.Unmarshal(respBody, &result)

	if result.MessageID != "" {
		return result.MessageID, nil
	}
	return mockMsgID(phoneNumber), nil
}

func (s *GENSMSService) SendBulk(targets []string, content, senderID string) (int, error) {
	sent := 0
	for _, phone := range targets {
		_, err := s.Send(phone, content, senderID)
		if err == nil {
			sent++
		}
	}
	return sent, nil
}

func mockMsgID(phone string) string {
	if len(phone) < 4 {
		return "mock-0000"
	}
	return "mock-" + phone[len(phone)-4:]
}
