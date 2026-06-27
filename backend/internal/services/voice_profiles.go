package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type VoiceProfile struct {
	ID         string  `json:"id"`
	AgeGroup   string  `json:"age_group"`
	Sex        string  `json:"sex"`
	Accent     string  `json:"accent"`
	VoiceModel string  `json:"voice_model"`
	SSMLVoice  string  `json:"ssml_voice"`
	Speed      float64 `json:"speed"`
	Pitch      float64 `json:"pitch"`
	Emotion    string  `json:"emotion"`
	Tone       string  `json:"tone"`
	Background string  `json:"background"`
	NoiseLevel int     `json:"noise_level"`
}

type VoiceProfilesManager struct {
	redisClient *redis.Client
	mu          sync.RWMutex
	profiles    map[string]*VoiceProfile
}

func NewVoiceProfilesManager(rdb *redis.Client) *VoiceProfilesManager {
	return &VoiceProfilesManager{
		redisClient: rdb,
		profiles:    make(map[string]*VoiceProfile),
	}
}

func (m *VoiceProfilesManager) LoadProfiles(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.redisClient == nil {
		return nil
	}

	pattern := "voice:profiles:*"
	cursor := uint64(0)
	var keys []string

	for {
		result, nextCursor, err := m.redisClient.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			break
		}
		keys = append(keys, result...)
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	for _, key := range keys {
		var profile VoiceProfile
		data, err := m.redisClient.Get(ctx, key).Bytes()
		if err != nil {
			continue
		}
		if err := json.Unmarshal(data, &profile); err != nil {
			slog.Error("Failed to unmarshal voice profile", "key", key, "error", err)
			continue
		}
		m.profiles[key] = &profile
	}

	return nil
}

func (m *VoiceProfilesManager) GetProfile(ctx context.Context, profileID string) (*VoiceProfile, error) {
	m.mu.RLock()
	if profile, ok := m.profiles[profileID]; ok {
		m.mu.RUnlock()
		return profile, nil
	}
	m.mu.RUnlock()

	if m.redisClient == nil {
		return nil, fmt.Errorf("profile not found: %s", profileID)
	}

	data, err := m.redisClient.Get(ctx, fmt.Sprintf("voice:profiles:%s", profileID)).Bytes()
	if err != nil {
		return nil, fmt.Errorf("profile not found: %s", profileID)
	}

	var profile VoiceProfile
	if err := json.Unmarshal(data, &profile); err != nil {
		return nil, err
	}

	m.mu.Lock()
	m.profiles[profileID] = &profile
	m.mu.Unlock()

	return &profile, nil
}

func (m *VoiceProfilesManager) CreateProfile(ctx context.Context, profile *VoiceProfile) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := fmt.Sprintf("voice:profiles:%s", profile.ID)
	data, err := json.Marshal(profile)
	if err != nil {
		return err
	}

	if m.redisClient != nil {
		if err := m.redisClient.Set(ctx, key, data, 24*time.Hour).Err(); err != nil {
			return err
		}
	}

	m.profiles[profile.ID] = profile
	return nil
}

func (m *VoiceProfilesManager) UpdateProfile(ctx context.Context, profileID string, updates map[string]interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Fetch profile directly from map or Redis without re-acquiring lock
	profile, ok := m.profiles[profileID]
	if !ok {
		if m.redisClient == nil {
			return fmt.Errorf("profile not found: %s", profileID)
		}
		data, err := m.redisClient.Get(ctx, fmt.Sprintf("voice:profiles:%s", profileID)).Bytes()
		if err != nil {
			return fmt.Errorf("profile not found: %s", profileID)
		}
		profile = &VoiceProfile{}
		if err := json.Unmarshal(data, profile); err != nil {
			return err
		}
	}

	// Apply updates
	for k, v := range updates {
		switch k {
		case "age_group":
			profile.AgeGroup = v.(string)
		case "sex":
			profile.Sex = v.(string)
		case "accent":
			profile.Accent = v.(string)
		case "voice_model":
			profile.VoiceModel = v.(string)
		case "ssml_voice":
			profile.SSMLVoice = v.(string)
		case "speed":
			profile.Speed = v.(float64)
		case "pitch":
			profile.Pitch = v.(float64)
		case "emotion":
			profile.Emotion = v.(string)
		case "tone":
			profile.Tone = v.(string)
		case "background":
			profile.Background = v.(string)
		case "noise_level":
			profile.NoiseLevel = v.(int)
		}
	}

	data, err := json.Marshal(profile)
	if err != nil {
		return err
	}

	if m.redisClient != nil {
		key := fmt.Sprintf("voice:profiles:%s", profileID)
		if err := m.redisClient.Set(ctx, key, data, 24*time.Hour).Err(); err != nil {
			return err
		}
	}

	m.profiles[profileID] = profile
	return nil
}

func (m *VoiceProfilesManager) DeleteProfile(ctx context.Context, profileID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.redisClient != nil {
		key := fmt.Sprintf("voice:profiles:%s", profileID)
		if err := m.redisClient.Del(ctx, key).Err(); err != nil {
			return err
		}
	}

	delete(m.profiles, profileID)
	return nil
}

func (m *VoiceProfilesManager) ListProfiles(ctx context.Context) ([]*VoiceProfile, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var profiles []*VoiceProfile
	for _, profile := range m.profiles {
		profiles = append(profiles, profile)
	}

	return profiles, nil
}

func (m *VoiceProfilesManager) GetAvailableProfiles(ctx context.Context) ([]*VoiceProfile, error) {
	return m.ListProfiles(ctx)
}

func PredefinedVoiceProfiles() []*VoiceProfile {
	return []*VoiceProfile{
		{
			ID:         "default_male_us",
			AgeGroup:   "young_adult",
			Sex:        "male",
			Accent:     "us",
			VoiceModel: "clue_con_v1",
			SSMLVoice:  "en-US-Alex",
			Speed:      1.0,
			Pitch:      1.0,
			Emotion:    "professional",
			Tone:       "confident",
			Background: "none",
			NoiseLevel: 10,
		},
		{
			ID:         "default_female_uk",
			AgeGroup:   "mature_adult",
			Sex:        "female",
			Accent:     "uk",
			VoiceModel: "clue_con_v1",
			SSMLVoice:  "en-GB-Sophie",
			Speed:      1.0,
			Pitch:      1.0,
			Emotion:    "friendly",
			Tone:       "calm",
			Background: "none",
			NoiseLevel: 10,
		},
		{
			ID:         "default_male_au",
			AgeGroup:   "young_adult",
			Sex:        "male",
			Accent:     "au",
			VoiceModel: "clue_con_v1",
			SSMLVoice:  "en-AU-Nick",
			Speed:      1.0,
			Pitch:      1.0,
			Emotion:    "cheerful",
			Tone:       "playful",
			Background: "none",
			NoiseLevel: 10,
		},
		{
			ID:         "default_female_in",
			AgeGroup:   "young_adult",
			Sex:        "female",
			Accent:     "in",
			VoiceModel: "clue_con_v1",
			SSMLVoice:  "en-IN-Aditi",
			Speed:      1.0,
			Pitch:      1.0,
			Emotion:    "professional",
			Tone:       "confident",
			Background: "none",
			NoiseLevel: 10,
		},
	}
}

func (m *VoiceProfilesManager) CreatePredefinedProfiles(ctx context.Context) error {
	profiles := PredefinedVoiceProfiles()

	for _, profile := range profiles {
		key := fmt.Sprintf("voice:profiles:%s", profile.ID)
		data, err := json.Marshal(profile)
		if err != nil {
			return err
		}

		if m.redisClient != nil {
			if err := m.redisClient.Set(ctx, key, data, 24*time.Hour).Err(); err != nil {
				slog.Error("Failed to create predefined profile", "id", profile.ID, "error", err)
				continue
			}
		}

		m.profiles[profile.ID] = profile
		slog.Info("Created predefined profile", "id", profile.ID)
	}

	return nil
}

func (m *VoiceProfilesManager) GetRandomProfile(ctx context.Context) (*VoiceProfile, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.profiles) == 0 {
		return nil, fmt.Errorf("no profiles available")
	}

	var ids []string
	for id := range m.profiles {
		ids = append(ids, id)
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	profileID := ids[rng.Intn(len(ids))]

	return m.GetProfile(ctx, profileID)
}
