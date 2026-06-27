package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

type CNAMService struct {
	apiKey    string
	baseURL   string
	client    *http.Client
	redis     *redis.Client
	cacheTTL  time.Duration
}

func NewCNAMService(apiKey string, rdb *redis.Client) *CNAMService {
	if apiKey == "" {
		apiKey = "demo"
	}
	return &CNAMService{
		apiKey:   apiKey,
		baseURL:  "https://freecnam.org",
		client:   &http.Client{Timeout: 5 * time.Second},
		redis:    rdb,
		cacheTTL: 1 * time.Hour,
	}
}

func (s *CNAMService) Lookup(number string) (string, error) {
	number = normalizePhone(number)

	if s.apiKey == "" || s.apiKey == "demo" {
		return mockCNAM(number), nil
	}

	// Check Redis cache first
	if s.redis != nil {
		val, err := s.redis.Get(context.Background(), "cnam:"+number).Result()
		if err == nil && val != "" {
			return val, nil
		}
	}

	name, err := s.fetchFromFreeCNAM(number)
	if err != nil {
		return "", err
	}

	// Cache result in Redis
	if s.redis != nil && name != "" {
		s.redis.Set(context.Background(), "cnam:"+number, name, s.cacheTTL)
	}

	return name, nil
}

func (s *CNAMService) fetchFromFreeCNAM(number string) (string, error) {
	url := fmt.Sprintf("%s/dip?q=%s", s.baseURL, number)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Accept", "application/json")
	if s.apiKey != "demo" {
		req.Header.Set("Authorization", "Bearer "+s.apiKey)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("cnam request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Name string `json:"name"`
	}
	json.Unmarshal(body, &result)

	if result.Name != "" {
		return result.Name, nil
	}
	return mockCNAM(number), nil
}

func mockCNAM(number string) string {
	names := map[string]string{
		"800": "TOLL FREE INFO",
		"555": "GENERAL SERVICE",
		"844": "TOLL FREE USA",
	}
	if len(number) >= 4 {
		area := number[1:4]
		if name, ok := names[area]; ok {
			return name
		}
	}
	return "VERIZON WIRELESS"
}

func normalizePhone(raw string) string {
	var digits []byte
	for _, r := range raw {
		if r >= '0' && r <= '9' {
			digits = append(digits, byte(r))
		}
	}
	s := string(digits)
	if len(s) == 10 {
		return "1" + s
	}
	if len(s) >= 11 && s[0] == '1' {
		return s[:11]
	}
	if len(s) > 0 {
		return "1" + s
	}
	return s
}
