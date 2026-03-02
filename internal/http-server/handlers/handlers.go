// Package handlers contains the handlers for the http server
package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"syscall"

	"github.com/biisal/fast-stream-bot/config"
	"github.com/biisal/fast-stream-bot/internal/bot"
	botutils "github.com/biisal/fast-stream-bot/internal/bot/bot-utils"
	"github.com/biisal/fast-stream-bot/internal/http-server/shortner"
	"github.com/biisal/fast-stream-bot/internal/stream"
	"github.com/biisal/fast-stream-bot/internal/types"
)

type StreamHandler struct {
	Worker   *bot.Worker
	Cfg      config.Config
	Shortner shortner.Shortner
}

////////////////////////////////////////////////////////////
// LANDING PAGE (FIXED - REQUIRED FOR BUILD)
////////////////////////////////////////////////////////////

func (h *StreamHandler) LandingPage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		type LandingData struct {
			AppName string
		}

		data := LandingData{
			AppName: h.Cfg.APP_NAME,
		}

		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)

		html := `
		<!DOCTYPE html>
		<html>
		<head>
			<title>` + data.AppName + `</title>
			<style>
				body {
					font-family: Arial, sans-serif;
					background: #111;
					color: #fff;
					text-align: center;
					padding-top: 100px;
				}
				h1 { color: #00ffcc; }
			</style>
		</head>
		<body>
			<h1>` + data.AppName + ` 🚀</h1>
			<p>Server is running successfully.</p>
		</body>
		</html>`

		_, _ = w.Write([]byte(html))
	}
}

////////////////////////////////////////////////////////////
// FILE STREAM
////////////////////////////////////////////////////////////

func (h *StreamHandler) ServerFile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		messageID, channelID, err := botutils.ParseMessageAndChannelId(
			r.PathValue("messageId"),
			r.PathValue("channelId"),
			h.Cfg.DB_CHANNEL_ID,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		downloadQuery := strings.TrimSpace(r.URL.Query().Get("d"))
		isDownload := downloadQuery == "1" || strings.ToLower(downloadQuery) == "true"

		hash := r.PathValue("hash")

		workerBot, err := h.Worker.HireFreeWorker()
		if workerBot == nil {
			http.Error(w, "No worker available", http.StatusInternalServerError)
			return
		}
		defer h.Worker.ReleaseWorker(workerBot)

		fileMsg, err := botutils.GetChannelMessage(
			context.Background(),
			channelID,
			messageID,
			workerBot.Client.API(),
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		file, err := botutils.GetMediaFromMessage(fileMsg)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if !botutils.CheckFileHash(file, hash) {
			http.Error(w, "Invalid hash", http.StatusForbidden)
			return
		}

		reader := stream.NewTgFileReader(
			workerBot.Client.API(),
			context.Background(),
			file.Location,
			file,
			r,
		)

		if err = reader.SetupStream(r, w, isDownload); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		buffer := make([]byte, stream.TelegramChunkSize)

		if _, err = io.CopyBuffer(w, reader, buffer); err != nil {

			if errors.Is(err, syscall.EPIPE) ||
				errors.Is(err, syscall.ECONNRESET) ||
				errors.Is(err, io.EOF) {
				return
			}

			return
		}
	}
}

////////////////////////////////////////////////////////////
// HOME STREAM PAGE
////////////////////////////////////////////////////////////

func (h *StreamHandler) HomeStream() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		errorResp := &types.ErrorResponse{Error: ""}

		messageID, channelID, err := botutils.ParseMessageAndChannelId(
			r.PathValue("messageId"),
			r.PathValue("channelId"),
			h.Cfg.DB_CHANNEL_ID,
		)
		if err != nil {
			errorResp.Error = err.Error()
			renderHTML(w, "error.html", errorResp)
			return
		}

		hash := r.URL.Query().Get("hash")
		streamLink := fmt.Sprintf("/stream/%d/%d/%s", channelID, messageID, hash)

		if strings.Contains(strings.ToLower(r.Header.Get("User-Agent")), "vlc") {
			http.Redirect(w, r, streamLink, http.StatusSeeOther)
			return
		}

		client, err := h.Worker.HireFreeWorker()
		if err != nil {
			errorResp.Error = "Failed to get bots"
			renderHTML(w, "error.html", errorResp)
			return
		}
		defer h.Worker.ReleaseWorker(client)

		fileMsg, err := botutils.GetChannelMessage(
			context.Background(),
			channelID,
			messageID,
			client.Client.API(),
		)
		if err != nil {
			errorResp.Error = "Failed to get file message"
			renderHTML(w, "error.html", errorResp)
			return
		}

		file, err := botutils.GetMediaFromMessage(fileMsg)
		if err != nil {
			errorResp.Error = "Failed to get media"
			renderHTML(w, "error.html", errorResp)
			return
		}

		if !botutils.CheckFileHash(file, hash) {
			errorResp.Error = "Invalid hash"
			renderHTML(w, "error.html", errorResp)
			return
		}

		downloadLink := fmt.Sprintf("%s?d=1", streamLink)

		fileInfo := &types.FileResponse{
			Title:        file.FileName,
			Size:         botutils.MakeSizeReadable(file.Size),
			DownloadLink: downloadLink,
			StreamLink:   streamLink,
			AppName:      h.Cfg.APP_NAME,
		}

		w.Header().Set("Cache-Control", "max-age=1200")
		renderHTML(w, "index.html", fileInfo)
	}
}

////////////////////////////////////////////////////////////
// HASH API
////////////////////////////////////////////////////////////

func (h *StreamHandler) MakeHashByChanMsgID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		messageIdStr := r.PathValue("messageId")
		channelIdStr := r.PathValue("channelId")

		messageId, channelId64, err := botutils.ParseMessageAndChannelId(
			messageIdStr,
			channelIdStr,
			0,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		workerBot, err := h.Worker.HireFreeWorker()
		if workerBot == nil {
			http.Error(w, "No worker available", http.StatusInternalServerError)
			return
		}
		defer h.Worker.ReleaseWorker(workerBot)

		fileMsg, err := botutils.GetChannelMessage(
			context.Background(),
			channelId64,
			messageId,
			workerBot.Client.API(),
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		file, err := botutils.GetMediaFromMessage(fileMsg)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		hash := botutils.MakeHashByFileInfo(file)

		res := &types.HashResponse{
			Data: types.HashInfo{
				Hash:      hash,
				MessageId: messageId,
				ChannelId: channelId64,
			},
			StatusCode: http.StatusOK,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(res)
	}
}

////////////////////////////////////////////////////////////
// PING
////////////////////////////////////////////////////////////

func (h *StreamHandler) Ping() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "pong")
	}
}

////////////////////////////////////////////////////////////
// TEMPLATE RENDER
////////////////////////////////////////////////////////////

func renderHTML(w http.ResponseWriter, htmlTemplate string, data any) {
	t, err := template.ParseFiles("frontend/" + htmlTemplate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = t.Execute(w, data)
}
