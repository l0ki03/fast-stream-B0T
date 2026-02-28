package routers

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/biisal/fast-stream-bot/config"
	"github.com/biisal/fast-stream-bot/internal/bot"
	"github.com/biisal/fast-stream-bot/internal/http-server/handlers"
	"github.com/biisal/fast-stream-bot/internal/http-server/shortner"
)

func GET(path string) string {
	return fmt.Sprintf("GET %s", path)
}

func SetUpRouters(worker *bot.Worker, Cfg config.Config, shortner shortner.Shortner) *http.ServeMux {
	slog.Info("Setting up routers")

	mux := http.NewServeMux()
	h := handlers.StreamHandler{
		Worker:   worker,
		Cfg:      Cfg,
		Shortner: shortner,
	}

	mux.HandleFunc(GET("/ping"), h.Ping())

	// 🔹 STREAM ROUTES
	mux.Handle(GET("/stream/{messageId}/{hash}"), h.ServerFile())
	mux.Handle(GET("/stream/{channelId}/{messageId}/{hash}"), h.ServerFile())

	// 🔹 WATCH ROUTES
	mux.Handle(GET("/watch/{messageId}"), h.HomeStream())
	mux.Handle(GET("/watch/{channelId}/{messageId}"), h.HomeStream())

	// 🔥 DOWNLOAD ROUTES (NEW ADDED)
	mux.Handle(GET("/download/{messageId}/{hash}"), h.DownloadFile())
	mux.Handle(GET("/download/{channelId}/{messageId}/{hash}"), h.DownloadFile())

	// 🔹 API
	mux.Handle(GET("/api/v1/hash/{channelId}/{messageId}"), h.MakeHashByChanMsgID())

	// 🔹 LANDING
	mux.Handle(GET("/"), h.LandingPage())

	// 🔹 STATIC FILES
	fs := http.FileServer(http.Dir("frontend/assets"))
	mux.Handle(GET("/static/"), http.StripPrefix("/static/", fs))

	mux.HandleFunc(GET("/favicon.ico"), func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	return mux
}
