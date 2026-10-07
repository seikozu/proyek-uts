package helper

import (
	"encoding/csv"

	"proyek-uts/app/model"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

const (
	FormatJSON = fiber.MIMEApplicationJSON
	FormatCSV  = "text/csv"
)

func Negotiate(c *fiber.Ctx, offered ...string) (string, error) {
	accept := strings.TrimSpace(c.Get(fiber.HeaderAccept))

	if accept == "" {
		return offered[0], nil
	}

	chosen := c.Accepts(offered...)
	if chosen == "" {
		return "", NotAcceptable("format yang diminta tidak tersedia, pilih salah satu dari: " + strings.Join(offered, ", "))
	}

	return chosen, nil
}

func WriteComicsCSV(c *fiber.Ctx, comics []model.Comic) error {
	c.Set(fiber.HeaderContentType, FormatCSV+"; charset=utf-8")
	c.Set(fiber.HeaderContentDisposition, `attachment; filename="comics.csv"`)

	var buffer strings.Builder
	writer := csv.NewWriter(&buffer)

	header := []string{"id", "title", "type", "author", "stock", "is_active", "created_at"}
	if err := writer.Write(header); err != nil {
		return Internal(err)
	}

	for _, item := range comics {
		row := []string{
			strconv.Itoa(item.ID),
			item.Title,
			item.Type,
			item.Author,
			strconv.Itoa(item.Stock),
			strconv.FormatBool(item.IsActive),
			item.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		}
		if err := writer.Write(row); err != nil {
			return Internal(err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return Internal(err)
	}

	return c.SendString(buffer.String())
}