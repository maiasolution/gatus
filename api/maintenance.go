package api

import (
	"github.com/TwiN/gatus/v5/config/maintenance"
	fiber "github.com/gofiber/fiber/v2"
)

// MaintenanceEvents returns all scheduled maintenance events keyed by endpoint key.
func MaintenanceEvents(c *fiber.Ctx) error {
	store := maintenance.GetEventsStore()
	if store == nil {
		return c.JSON(map[string]interface{}{})
	}
	return c.JSON(store.GetAllEvents())
}
