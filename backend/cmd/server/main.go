package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"

	"github.com/your-org/webapp-template/internal/domain"
	"github.com/your-org/webapp-template/internal/handler"
	"github.com/your-org/webapp-template/internal/infrastructure"
	"github.com/your-org/webapp-template/migrations"
	"github.com/your-org/webapp-template/internal/repository"
	"github.com/your-org/webapp-template/internal/usecase"
)

func main() {
	// .env 読み込み（本番では無視される）
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	ctx := context.Background()

	// --- DB 接続 ---
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		slog.Error("DATABASE_URL is required")
		os.Exit(1)
	}
	pool, err := infrastructure.NewDB(ctx, dbURL)
	if err != nil {
		slog.Error("failed to connect to database", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	migrations, err := loadMigrations()
	if err != nil {
		slog.Error("failed to load migrations", "err", err)
		os.Exit(1)
	}
	if err := infrastructure.RunMigrations(ctx, pool, migrations); err != nil {
		slog.Error("failed to run migrations", "err", err)
		os.Exit(1)
	}

	// --- セッションサービス ---
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		slog.Error("JWT_SECRET is required")
		os.Exit(1)
	}
	sessionSvc := infrastructure.NewJWTSessionService(jwtSecret, 24*time.Hour)

	// --- OAuth プロバイダ ---
	providers := make(map[string]domain.OAuthProvider)

	if id := os.Getenv("GOOGLE_CLIENT_ID"); id != "" {
		providers["google"] = infrastructure.NewGoogleOAuthProvider(
			id,
			os.Getenv("GOOGLE_CLIENT_SECRET"),
			os.Getenv("GOOGLE_REDIRECT_URL"),
		)
	}
	if id := os.Getenv("GITHUB_CLIENT_ID"); id != "" {
		providers["github"] = infrastructure.NewGitHubOAuthProvider(
			id,
			os.Getenv("GITHUB_CLIENT_SECRET"),
			os.Getenv("GITHUB_REDIRECT_URL"),
		)
	}
	if id := os.Getenv("MICROSOFT_CLIENT_ID"); id != "" {
		providers["microsoft"] = infrastructure.NewMicrosoftOAuthProvider(
			id,
			os.Getenv("MICROSOFT_CLIENT_SECRET"),
			os.Getenv("MICROSOFT_REDIRECT_URL"),
		)
	}

	// --- リポジトリ ---
	userRepo := repository.NewPostgresUserRepository(pool)
	providerRepo := repository.NewPostgresUserProviderRepository(pool)

	// --- ユースケース ---
	authUC := usecase.NewAuthUsecase(userRepo, providerRepo, sessionSvc, providers)

	// --- ハンドラ・ミドルウェア ---
	authHandler := handler.NewAuthHandler(authUC)
	authMW := handler.NewAuthMiddleware(sessionSvc)

	// --- ルーター ---
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{os.Getenv("FRONTEND_ORIGIN")},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	r.Get("/health", handler.Health)

	r.Route("/api/v1", func(r chi.Router) {
		// 公開エンドポイント: OAuth フロー
		r.Mount("/auth", authHandler.Router())

		// 認証必須エンドポイント
		r.Group(func(r chi.Router) {
			r.Use(authMW.Handler())

			r.Get("/auth/me", authHandler.HandleMe)
			r.Post("/auth/logout", authHandler.HandleLogout)

			// admin のみ
			r.Group(func(r chi.Router) {
				r.Use(handler.RequireRole(domain.RoleAdmin))
				r.Put("/users/{id}/role", authHandler.HandleUpdateRole)
			})
		})
	})

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// グレースフルシャットダウン
	go func() {
		slog.Info("server starting", "port", port, "providers", len(providers))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown error", "err", err)
	}
	slog.Info("server stopped")
}

func loadMigrations() ([]infrastructure.MigrationFile, error) {
	entries, err := migrations.FS.ReadDir(".")
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations dir: %w", err)
	}
	var files []infrastructure.MigrationFile
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		content, err := migrations.FS.ReadFile(e.Name())
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", e.Name(), err)
		}
		files = append(files, infrastructure.MigrationFile{
			Name:    e.Name(),
			Content: string(content),
		})
	}
	return files, nil
}
