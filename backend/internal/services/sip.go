package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/emiago/sipgo"
	"github.com/emiago/sipgo/sip"

	"hushcircuits/api/internal/models"
)

type DialogInfo struct {
	CallID    string
	LocalTag  string
	RemoteTag string
	RemoteURI string
	LocalURI  string
	CSeqNo    uint32
	RouteSet  []string
}

type SIPCall struct {
	CallID    string
	Dest      string
	From      string
	StartTime time.Time
	Done      chan struct{}
	HungUp    bool
	Dialog    *DialogInfo
}

type SIPService struct {
	host        string
	port        int
	username    string
	password    string
	transport   string
	displayName string
	callerID    string

	ua     *sipgo.UserAgent
	client *sipgo.Client

	mu          sync.RWMutex
	activeCalls map[string]*SIPCall
	registered  bool
}

func NewSIPService(host, port, username, password, transport, displayName, callerID string) *SIPService {
	portInt, _ := strconv.Atoi(port)
	if portInt == 0 {
		portInt = 5060
	}
	if callerID == "" {
		callerID = username
	}
	if displayName == "" {
		displayName = "HushCircuits"
	}
	return &SIPService{
		host:        host,
		port:        portInt,
		username:    username,
		password:    password,
		transport:   transport,
		displayName: displayName,
		callerID:    callerID,
		activeCalls: make(map[string]*SIPCall),
	}
}

func (s *SIPService) Start() error {
	slog.Info("starting SIP UA",
		"host", s.host,
		"port", s.port,
		"username", s.username,
		"transport", s.transport,
	)

	ua, err := sipgo.NewUA(
		sipgo.WithUserAgent("HushCircuits/2.0"),
	)
	if err != nil {
		return fmt.Errorf("failed to create SIP UA: %w", err)
	}
	s.ua = ua

	client, err := sipgo.NewClient(ua,
		sipgo.WithClientPort(0),
	)
	if err != nil {
		return fmt.Errorf("failed to create SIP client: %w", err)
	}
	s.client = client

	// Handle incoming SIP requests (OPTIONS keepalives, etc.)
	s.ua.TransactionLayer().OnRequest(func(req *sip.Request, tx *sip.ServerTx) {
		if req.Method == "OPTIONS" {
			resp := sip.NewResponseFromRequest(req, 200, "OK", nil)
			resp.AppendHeader(sip.NewHeader("User-Agent", "HushCircuits/2.0"))
			tx.Respond(resp)
			return
		}
	})

	// Register with SIP provider
	if err := s.register(); err != nil {
		slog.Warn("SIP registration failed, calls may still work", "error", err)
	} else {
		s.registered = true
		slog.Info("SIP registered successfully")
	}

	// Re-register every 30 minutes (registration expires in 3600s)
	go func() {
		ticker := time.NewTicker(30 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			if err := s.register(); err != nil {
				slog.Warn("SIP re-registration failed", "error", err)
				s.registered = false
			} else {
				s.registered = true
				slog.Info("SIP re-registered successfully")
			}
		}
	}()

	return nil
}

func (s *SIPService) register() error {
	uri := sip.Uri{
		Scheme: "sip",
		User:   s.username,
		Host:   s.host,
		Port:   s.port,
	}

	regReq := sip.NewRequest(sip.REGISTER, uri)
	maxFwd := sip.MaxForwardsHeader(70)
	regReq.AppendHeader(&maxFwd)

	localIP := getLocalIP()
	contactHDR := &sip.ContactHeader{
		Address: sip.Uri{
			User: s.username,
			Host: localIP,
			Port: s.port,
		},
	}
	regReq.AppendHeader(contactHDR)

	expires := sip.ExpiresHeader(3600)
	regReq.AppendHeader(&expires)
	regReq.AppendHeader(sip.NewHeader("User-Agent", "HushCircuits/2.0"))

	if err := sipgo.ClientRequestRegisterBuild(s.client, regReq); err != nil {
		return fmt.Errorf("build REGISTER: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := s.client.Do(ctx, regReq)
	if err != nil {
		return fmt.Errorf("send REGISTER: %w", err)
	}

	if res.StatusCode == 401 || res.StatusCode == 407 {
		digestAuth := sipgo.DigestAuth{
			Username: s.username,
			Password: s.password,
		}
		res, err = s.client.DoDigestAuth(ctx, regReq, res, digestAuth)
		if err != nil {
			return fmt.Errorf("digest auth REGISTER: %w", err)
		}
	}

	if res.StatusCode >= 200 && res.StatusCode < 300 {
		slog.Info("SIP registration successful", "status", res.StatusCode)
		return nil
	}

	return fmt.Errorf("registration failed with status %d", res.StatusCode)
}

func (s *SIPService) OriginateCall(req *models.OriginateCallRequest) (string, error) {
	if s.client == nil {
		return "", fmt.Errorf("SIP client not initialized")
	}

	dest := normalizeSIPNumber(req.DestinationNumber)
	callerID := s.callerID
	if req.SpoofedCallerID != "" {
		callerID = normalizeSIPNumber(req.SpoofedCallerID)
	}

	slog.Info("originating SIP call",
		"destination", dest,
		"caller_id", callerID,
		"spoofed_name", req.SpoofedName,
	)

	recipient := sip.Uri{
		Scheme: "sip",
		User:   dest,
		Host:   s.host,
		Port:   s.port,
	}

	inviteReq := sip.NewRequest(sip.INVITE, recipient)

	maxFwd := sip.MaxForwardsHeader(70)
	inviteReq.AppendHeader(&maxFwd)

	// From: our registered user (provider expects this)
	fromHeader := &sip.FromHeader{
		DisplayName: s.displayName,
		Address: sip.Uri{
			Scheme: "sip",
			User:   s.username,
			Host:   s.host,
		},
	}
	fromTag := generateTag()
	if fromHeader.Params == nil {
		fromHeader.Params = sip.HeaderParams{}
	}
	fromHeader.Params.Add("tag", fromTag)
	inviteReq.AppendHeader(fromHeader)

	// To: destination
	toHeader := &sip.ToHeader{
		Address: sip.Uri{
			Scheme: "sip",
			User:   dest,
			Host:   s.host,
		},
	}
	inviteReq.AppendHeader(toHeader)

	// Contact: our actual IP
	localIP := getLocalIP()
	contactHDR := &sip.ContactHeader{
		Address: sip.Uri{
			Scheme: "sip",
			User:   s.username,
			Host:   localIP,
			Port:   s.port,
		},
	}
	inviteReq.AppendHeader(contactHDR)

	inviteReq.AppendHeader(sip.NewHeader("User-Agent", "HushCircuits/2.0"))

	// P-Asserted-Identity for caller ID spoofing
	paiHeader := fmt.Sprintf("\"%s\" <sip:%s@%s>", req.SpoofedName, callerID, s.host)
	inviteReq.AppendHeader(sip.NewHeader("P-Asserted-Identity", paiHeader))

	// SDP — use a non-signaling port for RTP media
	sdpBody := buildSDP(localIP, "10006")
	ct := sip.ContentTypeHeader("application/sdp")
	inviteReq.AppendHeader(&ct)
	inviteReq.SetBody([]byte(sdpBody))

	// Send INVITE — handle auth challenges
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	slog.Info("sending INVITE", "destination", dest)
	res, err := s.client.Do(ctx, inviteReq)
	if err != nil {
		return "", fmt.Errorf("INVITE failed: %w", err)
	}

	// Handle 401/407 auth challenges
	if res.StatusCode == 401 || res.StatusCode == 407 {
		slog.Info("INVITE auth challenge, retrying with digest auth", "status", res.StatusCode)
		digestAuth := sipgo.DigestAuth{
			Username: s.username,
			Password: s.password,
		}
		res, err = s.client.DoDigestAuth(ctx, inviteReq, res, digestAuth)
		if err != nil {
			return "", fmt.Errorf("INVITE digest auth failed: %w", err)
		}
	}

	slog.Info("INVITE response", "status", res.StatusCode, "destination", dest)

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return "", fmt.Errorf("call rejected with status %d", res.StatusCode)
	}

	// Capture dialog info from 200 OK for BYE and future requests
	callID := fmt.Sprintf("sip-%d", time.Now().UnixMilli())

	dialog := &DialogInfo{
		CallID:   callID,
		LocalTag: fromTag,
		LocalURI: fmt.Sprintf("sip:%s@%s:%d", s.username, localIP, s.port),
		CSeqNo:   inviteReq.CSeq().SeqNo,
	}
	if from := res.From(); from != nil {
		if t, ok := from.Params.Get("tag"); ok {
			dialog.LocalTag = t
		}
	}
	if to := res.To(); to != nil {
		if t, ok := to.Params.Get("tag"); ok {
			dialog.RemoteTag = t
		}
	}
	if contact := res.Contact(); contact != nil {
		dialog.RemoteURI = contact.Address.String()
	}
	if rrs := res.GetHeaders("Record-Route"); len(rrs) > 0 {
		for _, rr := range rrs {
			dialog.RouteSet = append(dialog.RouteSet, rr.Value())
		}
	}

	// Determine ACK target: use remote Contact URI if available, fallback to original recipient
	ackTarget := recipient
	if dialog.RemoteURI != "" {
		var parsed sip.Uri
		if err := sip.ParseUri(dialog.RemoteURI, &parsed); err == nil {
			ackTarget = parsed
		}
	}

	ackReq := sip.NewRequest(sip.ACK, ackTarget)
	cid := sip.CallIDHeader(callID)
	ackReq.AppendHeader(&cid)
	cseq := &sip.CSeqHeader{SeqNo: dialog.CSeqNo, MethodName: sip.ACK}
	ackReq.AppendHeader(cseq)
	if via := inviteReq.Via(); via != nil {
		viaCopy := *via
		ackReq.AppendHeader(&viaCopy)
	}
	fromCopy := inviteReq.From()
	ackReq.AppendHeader(fromCopy)
	// Use the To from the INVITE (will be updated by provider in response)
	toCopy := inviteReq.To()
	ackReq.AppendHeader(toCopy)

	if err := s.client.WriteRequest(ackReq); err != nil {
		slog.Warn("ACK send failed", "call_id", callID, "error", err)
	}

	// Track the call with full dialog state
	sipCall := &SIPCall{
		CallID:    callID,
		Dest:      dest,
		From:      callerID,
		StartTime: time.Now(),
		Done:      make(chan struct{}),
		Dialog:    dialog,
	}

	s.mu.Lock()
	s.activeCalls[callID] = sipCall
	s.mu.Unlock()

	slog.Info("call established", "call_id", callID, "destination", dest, "status", res.StatusCode)

	return callID, nil
}

func (s *SIPService) HangupCall(callID string) error {
	s.mu.RLock()
	sipCall, exists := s.activeCalls[callID]
	s.mu.RUnlock()

	if !exists {
		slog.Warn("hangup: call not found", "call_id", callID)
		return nil
	}

	sipCall.HungUp = true

	dialog := sipCall.Dialog
	if dialog == nil || dialog.RemoteTag == "" {
		slog.Warn("hangup: no dialog state, sending blind BYE", "call_id", callID)
		// Blind BYE — best-effort, likely fails
		recipient := sip.Uri{
			Scheme: "sip",
			User:   sipCall.Dest,
			Host:   s.host,
			Port:   s.port,
		}
		byeReq := sip.NewRequest(sip.BYE, recipient)
		maxFwd := sip.MaxForwardsHeader(70)
		byeReq.AppendHeader(&maxFwd)
		byeReq.AppendHeader(sip.NewHeader("User-Agent", "HushCircuits/2.0"))
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		s.client.Do(ctx, byeReq)
	} else {
		// Proper BYE in dialog — construct with dialog identifiers per RFC 3261 §15
		byeTarget := s.buildByeTarget(dialog)
		byeReq := sip.NewRequest(sip.BYE, byeTarget)

		maxFwd := sip.MaxForwardsHeader(70)
		byeReq.AppendHeader(&maxFwd)

		cid := sip.CallIDHeader(dialog.CallID)
		byeReq.AppendHeader(&cid)

		var localURI sip.Uri
		if err := sip.ParseUri(dialog.LocalURI, &localURI); err != nil {
			slog.Warn("hangup: invalid local URI", "uri", dialog.LocalURI, "error", err)
		}
		fromHeader := &sip.FromHeader{
			DisplayName: s.displayName,
			Address:     localURI,
		}
		fromHeader.Params = sip.HeaderParams{}
		fromHeader.Params.Add("tag", dialog.LocalTag)
		byeReq.AppendHeader(fromHeader)

		var toURI sip.Uri
		if err := sip.ParseUri(fmt.Sprintf("sip:%s@%s", sipCall.Dest, s.host), &toURI); err != nil {
			slog.Warn("hangup: invalid to URI", "error", err)
		}
		toHeader := &sip.ToHeader{
			Address: toURI,
		}
		toHeader.Params = sip.HeaderParams{}
		toHeader.Params.Add("tag", dialog.RemoteTag)
		byeReq.AppendHeader(toHeader)

		cseq := &sip.CSeqHeader{SeqNo: dialog.CSeqNo + 1, MethodName: sip.BYE}
		byeReq.AppendHeader(cseq)

		localIP := getLocalIP()
		branch, _ := generateBranch()
		byeReq.AppendHeader(sip.NewHeader("Via", fmt.Sprintf("SIP/2.0/UDP %s:%d;branch=%s", localIP, s.port, branch)))

		byeReq.AppendHeader(sip.NewHeader("User-Agent", "HushCircuits/2.0"))

		// Add Route headers for strict/loose routing
		for _, route := range dialog.RouteSet {
			byeReq.AppendHeader(sip.NewHeader("Route", route))
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := s.client.Do(ctx, byeReq)
		if err != nil {
			slog.Error("BYE failed", "call_id", callID, "error", err)
		} else {
			slog.Info("call hung up", "call_id", callID)
		}
	}

	s.mu.Lock()
	delete(s.activeCalls, callID)
	s.mu.Unlock()

	close(sipCall.Done)
	return nil
}

func (s *SIPService) buildByeTarget(dialog *DialogInfo) sip.Uri {
	if dialog.RemoteURI != "" {
		var parsed sip.Uri
		if err := sip.ParseUri(dialog.RemoteURI, &parsed); err == nil {
			return parsed
		}
	}
	return sip.Uri{
		Scheme: "sip",
		Host:   s.host,
		Port:   s.port,
	}
}

func (s *SIPService) CaptureDTMF(callID, digit string, timestampMs int) error {
	slog.Info("DTMF captured", "call_id", callID, "digit", digit, "timestamp_ms", timestampMs)
	return nil
}

func (s *SIPService) MuteCall(callID string, muted bool) error {
	s.mu.RLock()
	_, exists := s.activeCalls[callID]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("call not found: %s", callID)
	}

	slog.Info("mute toggled", "call_id", callID, "muted", muted)
	return nil
}

func (s *SIPService) IsRegistered() bool {
	return s.registered
}

func (s *SIPService) ActiveCallCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.activeCalls)
}

func (s *SIPService) Close() {
	s.mu.Lock()
	for callID := range s.activeCalls {
		s.HangupCall(callID)
	}
	s.mu.Unlock()

	if s.client != nil {
		s.client.Close()
	}
	if s.ua != nil {
		s.ua.Close()
	}
}

func normalizeSIPNumber(num string) string {
	var digits []byte
	for _, r := range num {
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
	return s
}

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			return ipnet.IP.String()
		}
	}
	return "127.0.0.1"
}

func generateTag() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		for i := range b {
			b[i] = byte(time.Now().UnixNano() % 256)
		}
	}
	return hex.EncodeToString(b)
}

func generateBranch() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "z9hG4bK" + hex.EncodeToString(b), nil
}

func buildSDP(ip, port string) string {
	return fmt.Sprintf(`v=0
o=- %d %d IN IP4 %s
s=HushCircuits
c=IN IP4 %s
t=0 0
m=audio %s RTP/AVP 0 8 101
a=rtpmap:0 PCMU/8000
a=rtpmap:8 PCMA/8000
a=rtpmap:101 telephone-event/8000
a=fmtp:101 0-16
a=ptime:20
a=sendrecv
`, time.Now().UnixNano(), time.Now().UnixNano(), ip, ip, port)
}

