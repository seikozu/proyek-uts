package main

import (
	"context"
	"log/slog"
	"os"

	"proyek-uts/app/repository"
	"proyek-uts/app/service"
	"proyek-uts/config"
	"proyek-uts/database"
	"proyek-uts/helper"
	"proyek-uts/route"
	"time"
)

const minSecretLength = 32

func main() {
	config.LoadEnv()
	logger := config.NewLogger()

	jwtSecret := config.GetEnv("JWT_SECRET", "")
	if len(jwtSecret) < minSecretLength {
		logger.Error("JWT_SECRET tidak diisi atau terlalu pendek",
			slog.Int("minimal_karakter", minSecretLength))
		os.Exit(1)
	}

	pool, err := database.NewPool(context.Background())
	if err != nil {
		logger.Error("gagal terhubung ke database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer pool.Close()

	jwtManager := helper.NewJWTManager(
		jwtSecret,
		config.GetEnv("JWT_ISSUER", "manga-api"),
		time.Duration(config.GetEnvInt("JWT_ACCESS_TTL_MINUTES", 15))*time.Minute,
	)

	comicRepository := repository.NewComicRepository(pool)
	userRepository := repository.NewUserRepository(pool)
	tokenRepository := repository.NewTokenRepository(pool)
	roleRepository := repository.NewRoleRepository(pool)

	rawPermissions, err := roleRepository.LoadPermissions(context.Background())
	if err != nil {
		logger.Error("gagal memuat permission", slog.String("error", err.Error()))
		os.Exit(1)
	}
	permissions := helper.NewPermissionSet(rawPermissions)

	comicService := service.NewComicService(comicRepository, permissions)
	authService := service.NewAuthService(
		userRepository, tokenRepository, jwtManager, permissions,
		time.Duration(config.GetEnvInt("JWT_REFRESH_TTL_DAYS", 7))*24*time.Hour,
	)

	app := config.NewApp(logger)

	route.Register(app, route.Dependencies{
		Pool:         pool,
		JWT:          jwtManager,
		Permissions:  permissions,
		ComicService: comicService,
		AuthService:  authService,
	})

	port := config.GetEnv("APP_PORT", "3000")
	if err := app.Listen(":" + port); err != nil {
		logger.Error("gagal menjalankan server", slog.String("error", err.Error()))
	}
}