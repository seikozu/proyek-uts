package helper

import (
	"context"
	"time"

	"proyek-uts/app/model"

	"github.com/gofiber/fiber/v2"
)

const LocalsAuthUser = "authUser"

func CurrentUser(c *fiber.Ctx) (model.AuthUser, bool) {
	user, ok := c.Locals(LocalsAuthUser).(model.AuthUser)
	return user, ok
}

func RequestContext(c *fiber.Ctx) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.UserContext(), 5*time.Second)
}

func RequestID(c *fiber.Ctx) string {
	id, _ := c.Locals("requestid").(string)
	return id
}

func ParamID(c *fiber.Ctx) (int, bool) {
	id, err := c.ParamsInt("id")
	if err != nil || id < 1 {
		return 0, false
	}
	return id, true
}

func ParseCursorQuery(c *fiber.Ctx) (model.CursorQuery, error) {
	limit := c.QueryInt("limit", 10)
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	q := model.CursorQuery{
		Limit:  limit,
		Search: c.Query("search"),
		Type:   c.Query("type"),
	}

	if cursorStr := c.Query("cursor"); cursorStr != "" {
		cursor, err := DecodeCursor(cursorStr)
		if err != nil {
			return q, BadRequest("cursor tidak valid")
		}
		q.After = &cursor
	}

	return q, nil
}