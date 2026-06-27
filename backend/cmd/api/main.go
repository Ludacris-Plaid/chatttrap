package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/redis/go-redis/v9"

	"hushcircuits/api/internal/config"
	"hushcircuits/api/internal/database"
	"hushcircuits/api/internal/handlers"
	"hushcircuits/api/internal/middleware"
	"hushcircuits/api/internal/services"
)

func main() {
	cfg := config.Load()
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	// Database
	ctx := context.Background()
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		slog.Error("db connect failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	slog.Info("database connected")

	// Redis
	redisOpts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		slog.Warn("redis parse failed, continuing without", "error", err)
		redisOpts = nil
	}
	var rdb *redis.Client
	if redisOpts != nil {
		rdb = redis.NewClient(redisOpts)
		if err := rdb.Ping(ctx).Err(); err != nil {
			slog.Warn("redis ping failed, continuing without", "error", err)
			rdb = nil
		} else {
			slog.Info("redis connected")
		}
	}

	// Services
	smsSvc := services.NewGENSMSService(cfg.GensmsAPIKey)

	// SIP Telephony
	var sipSvc *services.SIPService
	if cfg.SIPUsername != "" {
		sipSvc = services.NewSIPService(
			cfg.SIPHost, cfg.SIPPort, cfg.SIPUsername, cfg.SIPPassword,
			cfg.SIPTransport, cfg.SIPDisplayName, cfg.SIPCallerID,
		)
		if err := sipSvc.Start(); err != nil {
			slog.Warn("SIP service failed to start", "error", err)
		}
	}

	fsSvc := services.NewFreeSwitchServiceWithSIP(sipSvc)
	cnamSvc := services.NewCNAMService(cfg.CNAMAPIKey, rdb)
	pool := db.Pool
	tokenSvc := services.NewTokenService(rdb)
	tokenSvc.SetPool(pool)
	featherlessSvc := services.NewFeatherlessService(cfg.FeatherlessAPIKey, cfg.FeatherlessModel, cfg.FeatherlessPrompt)
	pmtSvc := services.NewNowPaymentsService(cfg.NowPaymentsAPIKey, cfg.NowPaymentsIPNSecret, cfg.NowPaymentsCallbackURL)
	vouchSvc := services.NewVoucherService()
	vouchSvc.SetPool(pool)

	// Handlers
	auth := handlers.NewAuthHandler(cfg.JWTSecret, pool)
	dialer := handlers.NewDialerHandler(pool, fsSvc, cnamSvc, tokenSvc)
	scripts := handlers.NewScriptHandler(pool, featherlessSvc, 50)
	sms := handlers.NewSMSHandler(pool, smsSvc)
	wallet := handlers.NewWalletHandler(pool, pmtSvc, vouchSvc, cfg.TokenRate, cfg.VIPPrice, cfg.VIPDurationDays)
	stats := handlers.NewStatsHandler(pool)
	admin := handlers.NewAdminHandler(pool, vouchSvc)
	webhook := handlers.NewWebhookHandler(pool, pmtSvc)
	otp := handlers.NewOTPHandler(pool, rdb, fsSvc, smsSvc, featherlessSvc, tokenSvc, cfg.OTPGrabCostUSD, cfg.OTPGrabTokens, cfg.OTPGrabCallCostUSD, cfg.OTPGrabCallTokens)
	settings := handlers.NewSettingsHandler(pool)
	sipRegistered := func() bool { return false }
	if sipSvc != nil {
		sipRegistered = sipSvc.IsRegistered
	}
	health := handlers.NewHealthHandler(pool, cfg.GensmsAPIKey, cfg.FeatherlessAPIKey, cfg.CNAMAPIKey, cfg.NowPaymentsAPIKey, sipRegistered)

	// Fiber
	app := fiber.New(fiber.Config{
		AppName:           "HushCircuits Pro",
		EnablePrintRoutes: true,
		ServerHeader:      "HushCircuits",
		ReadBufferSize:    32768,
	})

	// Global middleware
	app.Use(recover.New())
	app.Use(compress.New())
	app.Use(middleware.CORS(cfg.AllowedOrigins))

	// Public routes (no auth required)
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "version": "2.0.0"})
	})
	app.Post("/api/auth/login", auth.Login)

	// Webhook routes (no auth - verified by signature)
	app.Post("/api/webhooks/nowpayments", webhook.NowPayments)

	// Auth middleware for all other /api routes
	app.Use(middleware.Auth(cfg.JWTSecret))

	// API routes (auth required)
	api := app.Group("/api")

	// Dialer
	api.Get("/cnam", dialer.LookupCNAM)
	api.Post("/call", dialer.OriginateCall)
	api.Post("/call/:callId/end", dialer.EndCall)
	api.Get("/call/:callId", dialer.GetCall)
	api.Post("/call/:callId/dtmf", dialer.SubmitDTMF)
	api.Get("/call/:callId/dtmf", dialer.GetDTMF)
	api.Post("/call/:callId/mute", dialer.MuteCall)

	// OTP Grab
	api.Post("/otp/grab", otp.LaunchGrab)
	api.Get("/otp/grab/:id", otp.GetGrabStatus)
	api.Get("/otp/grabs", otp.ListGrabs)

	// Scripts
	api.Post("/script/generate", scripts.Generate)
	api.Get("/scripts", scripts.List)
	api.Delete("/script/:id", scripts.Delete)

	// SMS
	api.Post("/sms/send", sms.SendSingle)
	api.Post("/sms/bulk", sms.SendBulk)

	// Wallet
	api.Get("/wallet/balance", wallet.GetBalance)
	api.Post("/wallet/payment", wallet.CreatePayment)
	api.Post("/wallet/voucher", wallet.RedeemVoucher)
	api.Post("/wallet/vip", wallet.UpgradeVIP)
	api.Get("/wallet/transactions", wallet.GetTransactions)

	// Stats
	api.Get("/stats", stats.GetStats)
	api.Get("/stats/calls", stats.GetRecentCalls)

	// Admin
	adminGroup := api.Group("/admin")
	adminGroup.Use(middleware.AdminEmail(cfg.AdminEmail))
	adminGroup.Get("/dashboard", admin.Dashboard)
	adminGroup.Get("/users", admin.ListUsers)
	adminGroup.Post("/user/:userId/balance", admin.AdjustBalance)
	adminGroup.Post("/voucher", admin.GenerateVoucher)
	adminGroup.Get("/dtmf-logs", admin.GetAllDTMFLogs)

	// Settings
	api.Get("/settings", settings.GetSettings)
	api.Put("/settings", settings.UpdateSettings)

	// Service health
	api.Get("/health/services", health.GetServices)

	// Start
	go func() {
		slog.Info("server starting", "port", cfg.Port)
		if err := app.Listen(":" + cfg.Port); err != nil {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	app.ShutdownWithContext(shutdownCtx)

	if sipSvc != nil {
		sipSvc.Close()
	}
	if rdb != nil {
		rdb.Close()
	}
	db.Close()
}
