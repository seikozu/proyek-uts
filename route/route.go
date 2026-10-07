package route

import (
	"context"
	"time"

	"proyek-uts/app/service"
	"proyek-uts/helper"
	"proyek-uts/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Dependencies struct {
	Pool         *pgxpool.Pool
	JWT          *helper.JWTManager
	Permissions  *helper.PermissionSet
	ComicService *service.ComicService
	AuthService  *service.AuthService
}

func Register(app *fiber.App, deps Dependencies) {
	api := app.Group("/api/v1")

	api.Get("/health", healthCheck(deps.Pool))

	auth := api.Group("/auth", middleware.RequireJSON)
	auth.Post("/register", deps.AuthService.Register)
	auth.Post("/login", middleware.LoginRateLimiter(), deps.AuthService.Login)
	auth.Post("/refresh", deps.AuthService.Refresh)
	auth.Post("/logout", deps.AuthService.Logout)
	auth.Get("/me", middleware.RequireAuth(deps.JWT), deps.AuthService.Me)
	auth.Patch("/users/:id/role", middleware.RequireAuth(deps.JWT), middleware.RequirePermission(deps.Permissions, "role:assign"), deps.AuthService.AssignRole)

	comics := api.Group("/comics", middleware.RequireJSON, middleware.RequireAuth(deps.JWT))

	perms := deps.Permissions

	comics.Get("/", middleware.RequirePermission(perms, "comic:list"), deps.ComicService.List)
	comics.Post("/", middleware.RequirePermission(perms, "comic:create"), deps.ComicService.Create)
	comics.Get("/:id", deps.ComicService.Get)
	comics.Put("/:id", deps.ComicService.Replace)
	comics.Patch("/:id", deps.ComicService.Patch)
	comics.Delete("/:id", deps.ComicService.Delete)

	app.Use(func(c *fiber.Ctx) error {
		return helper.NotFound("endpoint tidak ditemukan")
	})
}

func healthCheck(pool *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			return helper.Internal(err)
		}

		return helper.Success(c, fiber.StatusOK, "server dan database berjalan", fiber.Map{
			"timestamp": time.Now(),
		})
	}
}