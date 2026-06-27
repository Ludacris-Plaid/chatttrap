package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"

	"hushcircuits/api/internal/models"
)

type AIPrompt struct {
	TargetName    string
	TargetAge     int
	TargetService string
	TargetDetails string
	Goal          string
	ScriptType    string
	Context       string
	Conversation  []ConversationTurn
}

type TargetProfile struct {
	ID      string
	Accents []string
	Gender  string
	Age     string
	Names   []string
}

type ConversationTurn struct {
	Role      string
	Text      string
	Timestamp time.Time
}

type AIEngine struct {
	redisClient    *redis.Client
	mu             sync.RWMutex
	cache          map[string]*AIResponse
	featherlessSvc *FeatherlessService
	fallbackScript string
}

func NewAIEngine(rdb *redis.Client, fs *FeatherlessService) *AIEngine {
	return &AIEngine{
		redisClient:    rdb,
		featherlessSvc: fs,
		cache:          make(map[string]*AIResponse),
		fallbackScript: fs.FallbackScript(&models.GenerateScriptRequest{Goal: "OTP Theft", TargetService: "Bank"}),
	}
}

func (e *AIEngine) GenerateResponse(ctx context.Context, prompt *AIPrompt) (*AIResponse, error) {
	cacheKey := fmt.Sprintf("ai:response:%s", prompt.TargetName)

	e.mu.RLock()
	if cached, ok := e.cache[cacheKey]; ok {
		e.mu.RUnlock()
		return cached, nil
	}
	e.mu.RUnlock()

	if e.redisClient != nil {
		cacheData := e.redisClient.Get(ctx, cacheKey)
		if result, err := cacheData.Result(); err == nil && result != "" {
			var response AIResponse
			if err := json.Unmarshal([]byte(result), &response); err == nil {
				e.mu.Lock()
				e.cache[cacheKey] = &response
				e.mu.Unlock()
				return &response, nil
			}
		}
	}

	var response AIResponse
	if prompt.Context == "demo" || prompt.Context == "" {
		response = AIResponse{
			Content:  e.fallbackScript,
			Target:   prompt.TargetName,
			Adaptive: false,
		}
	} else {
		resp, err := e.featherlessSvc.Generate(ctx, prompt)
		if err != nil {
			slog.Error("Featherless API failed, using fallback", "error", err)
			response = AIResponse{
				Content:  e.fallbackScript,
				Target:   prompt.TargetName,
				Adaptive: false,
			}
		} else {
			response = *resp
		}
	}

	e.mu.Lock()
	e.cache[cacheKey] = &response
	e.mu.Unlock()

	if e.redisClient != nil {
		cacheBytes, err := json.Marshal(response)
		if err == nil {
			if err := e.redisClient.Set(ctx, cacheKey, cacheBytes, 5*time.Minute).Err(); err != nil {
				slog.Error("Failed to cache AI response", "error", err)
			}
		}
	}

	return &response, nil
}

type AIResponse struct {
	Content   string
	Target    string
	Adaptive  bool
	Timestamp time.Time
}

type ConversationContext struct {
	Conversation []ConversationTurn
	TargetProfile *TargetProfile
}

func (e *AIEngine) UpdateConversationContext(ctx context.Context, sessionID string, turn ConversationTurn) error {
	key := fmt.Sprintf("ai:conversation:%s", sessionID)

	if e.redisClient == nil {
		return fmt.Errorf("redis not available")
	}

	data, err := e.redisClient.Get(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("conversation not found: %s: %w", sessionID, err)
	}

	var convCtx ConversationContext
	if err := json.Unmarshal([]byte(data), &convCtx); err != nil {
		return err
	}

	convCtx.Conversation = append(convCtx.Conversation, turn)

	updatedData, err := json.Marshal(convCtx)
	if err != nil {
		return err
	}

	return e.redisClient.Set(ctx, key, updatedData, 24*time.Hour).Err()
}

func (e *AIEngine) GenerateScript(phoneNumber, targetName string) (string, error) {
	prompt := &AIPrompt{
		TargetName: targetName,
		Context:    "demo",
	}
	resp, err := e.GenerateResponse(context.Background(), prompt)
	if err != nil {
		return e.fallbackScript, err
	}
	return resp.Content, nil
}

func (e *AIEngine) FallbackScript() string {
	return e.fallbackScript
}
