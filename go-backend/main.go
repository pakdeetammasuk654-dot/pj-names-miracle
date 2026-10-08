package main

import (
	"context"
	"log"
	"os"
	"time"

	"go-backend/config"
	"go-backend/delivery/http/handler"
	"go-backend/delivery/http/router"
	"go-backend/domain"
	"go-backend/repository"
	"go-backend/usecase"
)

func main() {
	// 1. Initialize PostgreSQL Connection
	db, err := config.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// 2. Initialize Repository
	nameRepo := repository.NewNameRepository(db)
	articleRepo := repository.NewArticleRepository(db)
	dimensionRepo := repository.NewDimensionLabelRepository(db)
	auspiciousRepo := repository.NewAuspiciousRuleRepository(db)
	auspiciousContentRepo := repository.NewAuspiciousContentRepository(db)
	calibrationProfile, calibrationWarm := warmSoundCalibrationCache(nameRepo)

	// 3. Initialize UseCases
	healthUseCase := usecase.NewHealthUseCase("Go-Clean-Architecture-Service")
	nameSearchUseCase := usecase.NewNameSearchUseCase(nameRepo)
	auspiciousService := usecase.NewAuspiciousService(auspiciousRepo, auspiciousContentRepo)

	// 4. Initialize Handlers
	healthHandler := handler.NewHealthHandler(healthUseCase)
	readinessHandler := handler.NewReadinessHandler(db, calibrationProfile, calibrationWarm)
	nameHandler := handler.NewNameHandler(nameSearchUseCase, dimensionRepo)
	entitlementService := usecase.NewEntitlementServiceFromEnv()
	nameHandler.SetEntitlementService(entitlementService)
	entitlementHandler := handler.NewEntitlementHandler(entitlementService)
	if !entitlementService.Configured() {
		log.Print("[Startup] Purchase verification is not configured; premium search results will remain locked")
	}
	articleHandler, err := handler.NewArticleHandler(articleRepo)
	if err != nil {
		log.Fatalf("Failed to initialize article handler: %v", err)
	}
	articleAdminHandler, err := handler.NewArticleAdminHandler(
		articleRepo,
		os.Getenv("ARTICLE_ADMIN_USERNAME"),
		os.Getenv("ARTICLE_ADMIN_PASSWORD"),
	)
	if err != nil {
		log.Fatalf("Failed to initialize article admin handler: %v", err)
	}
	articleMediaHandler, err := handler.NewArticleMediaHandler(articleRepo, articleAdminHandler)
	if err != nil {
		log.Fatalf("Failed to initialize article media handler: %v", err)
	}
	calendarHandler := handler.NewAuspiciousCalendarHandler(auspiciousService)
	adminDashboardHandler, err := handler.NewAdminDashboardHandler(articleRepo, articleAdminHandler)
	if err != nil {
		log.Fatalf("Failed to initialize admin dashboard handler: %v", err)
	}
	dimensionAdminHandler, err := handler.NewDimensionAdminHandler(dimensionRepo, articleAdminHandler)
	if err != nil {
		log.Fatalf("Failed to initialize dimension admin handler: %v", err)
	}
	auspiciousAdminHandler, err := handler.NewAuspiciousAdminHandler(auspiciousContentRepo, articleAdminHandler)
	if err != nil {
		log.Fatalf("Failed to initialize auspicious admin handler: %v", err)
	}
	supportChatRepo := repository.NewSupportChatRepository(db)
	pushService := usecase.NewPushNotificationService(supportChatRepo)
	supportChatHandler, err := handler.NewSupportChatHandler(supportChatRepo, articleAdminHandler, pushService)
	if err != nil {
		log.Fatalf("Failed to initialize support chat handler: %v", err)
	}

	nameAdminRepo := repository.NewNameAdminRepository(db)
	nameAdminUseCase := usecase.NewNameAdminUseCase(nameAdminRepo)
	nameAdminHandler, err := handler.NewNameAdminHandler(nameAdminUseCase, articleAdminHandler)
	if err != nil {
		log.Fatalf("Failed to initialize name admin handler: %v", err)
	}
	adminDashboardHandler.SetNameAdminRepo(nameAdminRepo)

	// 5. Setup Gin Engine & Routes
	r := router.SetupRouter(
		healthHandler,
		readinessHandler,
		nameHandler,
		articleHandler,
		articleAdminHandler,
		articleMediaHandler,
		calendarHandler,
		adminDashboardHandler,
		supportChatHandler,
		dimensionAdminHandler,
		auspiciousAdminHandler,
		nameAdminHandler,
		entitlementHandler,
	)

	port := getEnv("PORT", "8080")
	addr := getEnv("SERVER_ADDR", ":"+port)
	log.Printf("Starting Gin server on %s...", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// warmSoundCalibrationCache moves the expensive provisional corpus scan out of
// the first user's search request. GetSoundCalibrationProfile keeps the result
// in the process-local repository cache, so subsequent searches only clone the
// already prepared profile. A warm-up failure is non-fatal: the normal request
// path will retry and return its existing error handling to the client.
func warmSoundCalibrationCache(nameRepo repository.NameRepository) (domain.SoundCalibrationProfile, bool) {
	const timeout = 30 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	startedAt := time.Now()
	profile, err := nameRepo.GetSoundCalibrationProfile(ctx)
	if err != nil {
		log.Printf("[Startup] Sound calibration cache warm-up failed after %s: %v", time.Since(startedAt).Round(time.Millisecond), err)
		return domain.SoundCalibrationProfile{}, false
	}
	log.Printf(
		"[Startup] Sound calibration cache ready duration=%s source=%s approved=%t reference_version=%s versions=%d",
		time.Since(startedAt).Round(time.Millisecond),
		profile.Source,
		profile.Approved,
		profile.ReferenceVersion,
		len(profile.Versions),
	)
	return profile, true
}
