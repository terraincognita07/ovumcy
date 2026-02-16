package api

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(app *fiber.App, handler *Handler) {
	app.Get("/healthz", handler.Health)
	app.Get("/lang/:lang", handler.SetLanguage)

	app.Get("/login", handler.ShowLoginPage)
	app.Get("/", handler.AuthRequired, handler.ShowDashboard)
	app.Get("/calendar", handler.AuthRequired, handler.ShowCalendar)
	app.Get("/calendar/day/:date", handler.AuthRequired, handler.CalendarDayPanel)
	app.Get("/stats", handler.AuthRequired, handler.ShowStats)

	api := app.Group("/api")

	auth := api.Group("/auth")
	auth.Get("/setup-status", handler.SetupStatus)
	auth.Post("/register", handler.Register)
	auth.Post("/login", handler.Login)
	auth.Post("/logout", handler.AuthRequired, handler.Logout)

	days := api.Group("/days", handler.AuthRequired)
	days.Get("", handler.GetDays)
	days.Get("/:date", handler.GetDay)
	days.Post("/:date", handler.OwnerOnly, handler.UpsertDay)

	symptoms := api.Group("/symptoms", handler.AuthRequired)
	symptoms.Get("", handler.GetSymptoms)
	symptoms.Post("", handler.OwnerOnly, handler.CreateSymptom)
	symptoms.Delete("/:id", handler.OwnerOnly, handler.DeleteSymptom)

	stats := api.Group("/stats", handler.AuthRequired)
	stats.Get("/overview", handler.GetStatsOverview)
}
