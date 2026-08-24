package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"gocamping/internal/config"
	"gocamping/internal/httpx"
	"gocamping/internal/service"
	"gocamping/internal/tiles"
)

func NewRouter(cfg config.Config, app *service.App, ready func() error) http.Handler {
	r := chi.NewRouter()
	r.Use(httpx.Recover)
	r.Use(httpx.RequestID)
	r.Use(httpx.AccessLog)
	r.Use(httpx.CORS(cfg.CORSOrigins))

	r.Get("/healthz", func(w http.ResponseWriter, req *http.Request) {
		httpx.JSON(w, req, 200, map[string]string{"status": "ok"})
	})
	r.Get("/readyz", func(w http.ResponseWriter, req *http.Request) {
		if err := ready(); err != nil {
			httpx.Fail(w, req, httpx.Internal("not ready"))
			return
		}
		httpx.JSON(w, req, 200, map[string]string{"status": "ready"})
	})
	r.Get("/tiles/{z}/{x}/{y}.png", tiles.Handler)

	h := &Handler{Cfg: cfg, App: app}
	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/auth/register", h.Register)
		r.Post("/auth/login", h.Login)
		r.Get("/ws", h.WS)

		r.Group(func(r chi.Router) {
			r.Use(httpx.RequireAuth(cfg.JWTSecret))
			r.Get("/me", h.Me)
			r.Get("/route-books", h.ListRoutes)
			r.Post("/route-books", h.SaveRoute)
			r.Get("/route-books/{id}", h.GetRoute)
			r.Put("/route-books/{id}", h.UpdateRoute)
			r.Post("/route-books/{id}/elevation", h.Elevation)
			r.Get("/route-books/{id}/export", h.Export)
			r.Post("/route-books/import", h.Import)
			r.Get("/teams", h.ListTeams)
			r.Post("/teams", h.CreateTeam)
			r.Get("/teams/{id}", h.GetTeam)
			r.Post("/teams/join", h.JoinTeam)
			r.Post("/teams/{id}/kick", h.Kick)
			r.Post("/trips", h.CreateTrip)
			r.Get("/trips/{id}", h.GetTrip)
			r.Get("/teams/{id}/trips", h.ListTrips)
			r.Post("/trips/{id}/start", h.StartTrip)
			r.Post("/trips/{id}/pause", h.PauseTrip)
			r.Post("/trips/{id}/finish", h.FinishTrip)
			r.Post("/trips/{id}/positions", h.PostPosition)
			r.Post("/trips/{id}/tracks/batch", h.BatchTracks)
			r.Get("/trips/{id}/tracks", h.GetTracks)
			r.Get("/trips/{id}/eta", h.ETA)
			r.Post("/trips/{id}/sos", h.SOS)
			r.Get("/trips/{id}/sos", h.ListSOS)
			r.Post("/sos/{id}/resolve", h.ResolveSOS)
			r.Get("/trips/{id}/risk", h.Risk)
			r.Get("/trips/{id}/backtrack", h.Backtrack)
			r.Get("/trips/{id}/replay", h.Replay)
			r.Post("/trips/{id}/simulate", h.Simulate)
			r.Group(func(r chi.Router) {
				r.Use(httpx.RequireAdmin)
				r.Get("/admin/users", h.AdminUsers)
				r.Get("/admin/metrics", h.AdminMetrics)
				r.Get("/admin/route-books", h.AdminRoutes)
			})
		})
	})
	return r
}
