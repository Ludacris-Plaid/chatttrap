package services

import (
	"log/slog"

	"hushcircuits/api/internal/models"
)

type FreeSwitchService struct {
	sipSvc *SIPService
}

func NewFreeSwitchService(host string, port int, password string) *FreeSwitchService {
	return &FreeSwitchService{}
}

func NewFreeSwitchServiceWithSIP(sipSvc *SIPService) *FreeSwitchService {
	return &FreeSwitchService{sipSvc: sipSvc}
}

func (s *FreeSwitchService) OriginateCall(req *models.OriginateCallRequest) (string, error) {
	if s.sipSvc != nil {
		return s.sipSvc.OriginateCall(req)
	}
	slog.Warn("no SIP service configured, call not originated")
	return "", nil
}

func (s *FreeSwitchService) CaptureDTMF(callID, digit string, timestampMs int) error {
	if s.sipSvc != nil {
		return s.sipSvc.CaptureDTMF(callID, digit, timestampMs)
	}
	return nil
}

func (s *FreeSwitchService) HangupCall(callID string) error {
	if s.sipSvc != nil {
		return s.sipSvc.HangupCall(callID)
	}
	return nil
}

func (s *FreeSwitchService) MuteCall(callID string, muted bool) error {
	if s.sipSvc != nil {
		return s.sipSvc.MuteCall(callID, muted)
	}
	return nil
}
