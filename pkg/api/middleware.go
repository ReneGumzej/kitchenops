package api

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v3"

	"gitlab.com/ReneGumzej/kitchenops/pkg/logging"
)

// RequestLogger protokolliert jede HTTP-Anfrage strukturiert:
// Methode, Pfad, Status, Dauer, IP und Request-ID.
// Fehler aus Handlern (fiber.NewError) werden mit Fehlermeldung auf
// Error-Ebene geloggt - nicht nur als Statuscode.
// /health wird ausgespart, um die Logs ruhig zu halten.
func RequestLogger() fiber.Handler {
	return func(ctx fiber.Ctx) error {
		start := time.Now()

		err := ctx.Next()

		if ctx.Path() == "/health" {
			return err
		}

		status := ctx.Response().StatusCode()
		if err != nil {
			var fiberErr *fiber.Error
			if errors.As(err, &fiberErr) {
				status = fiberErr.Code
			}
		}

		attrs := []any{
			"methode", ctx.Method(),
			"pfad", ctx.Path(),
			"status", status,
			"dauer_ms", time.Since(start).Milliseconds(),
			"ip", ctx.IP(),
			// Request-ID aus dem Response-Header (fiber v3 setzt keine Locals)
			"request_id", string(ctx.Response().Header.Peek("X-Request-ID")),
		}

		if err != nil {
			attrs = append(attrs, "fehler", err.Error())
			logging.Error("anfrage fehlgeschlagen", attrs...)
		} else {
			logging.Info("anfrage", attrs...)
		}
		return err
	}
}
