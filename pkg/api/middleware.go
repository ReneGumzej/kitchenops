package api

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/requestid"

	"gitlab.com/ReneGumzej/kitchenops/pkg/logging"
)

// RequestLogger liefert eine Fiber-v3-Middleware, die jede eingehende Anfrage
// strukturiert über slog protokolliert (Methode, Pfad, Status, Dauer, IP,
// Request-ID). Fehler aus dem Handler werden mit geloggt.
func RequestLogger() fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		attrs := []any{
			"method", c.Method(),
			"path", c.Path(),
			"status", c.Response().StatusCode(),
			"duration_ms", time.Since(start).Milliseconds(),
			"ip", c.IP(),
			"request_id", requestid.FromContext(c),
		}
		if err != nil {
			attrs = append(attrs, "fehler", err.Error())
			logging.Error("http request", attrs...)
			return err
		}
		logging.Info("http request", attrs...)
		return nil
	}
}
