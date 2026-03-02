package stream

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"

	"github.com/alist-org/alist/v3/pkg/http_range"
	"github.com/biisal/fast-stream-bot/internal/types"
	"github.com/gotd/td/tg"
)

const (
	TelegramChunkSize = 1024 * 1024
)

type TgFileReader struct {
	ctx          context.Context
	cachedChunk  []byte
	cachedOffset int64
	TgAPI        *tg.Client
	start        int64
	end          int64
	FileLocation *tg.InputDocumentFileLocation
	File         *types.File
	finished     bool
	mu           sync.RWMutex
}

func (r *TgFileReader) isFinished() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.finished

}
func (r *TgFileReader) setFinished() {
	r.mu.Lock()
	r.finished = true
	r.mu.Unlock()
}

func (r *TgFileReader) SetupStream(req *http.Request, w http.ResponseWriter, isDownload bool) error {

	// ✅ IMPORTANT FIX: Always send filename to client (VLC fix)
	if isDownload {
		w.Header().Set("Content-Disposition", "attachment; filename=\""+r.File.FileName+"\"")
	} else {
		w.Header().Set("Content-Disposition", "inline; filename=\""+r.File.FileName+"\"")
	}

	rangeHeader := req.Header.Get("Range")
	isFull := rangeHeader == ""

	if isFull {
		r.start = 0
		r.end = r.File.Size - 1
	} else {
		ranges, err := http_range.ParseRange(rangeHeader, r.File.Size)
		if err != nil {
			slog.Error("Failed to parse range", "error", err)
			return fmt.Errorf("failed to parse range: %w", err)
		}

		if len(ranges) == 0 {
			return fmt.Errorf("no valid ranges found")
		}

		length := ranges[0].Length
		r.start = ranges[0].Start
		r.end = r.start + length - 1

		if r.end >= r.File.Size {
			r.end = r.File.Size - 1
		}

		w.Header().Set("Content-Range",
			fmt.Sprintf("bytes %d-%d/%d", r.start, r.end, r.File.Size))
	}

	if r.File.MimeType == "" {
		r.File.MimeType = "application/octet-stream"
	}

	w.Header().Set("Content-Type", r.File.MimeType)
	w.Header().Set("Accept-Ranges", "bytes")

	contentLength := r.end - r.start + 1
	w.Header().Set("Content-Length", fmt.Sprintf("%d", contentLength))

	if isFull {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusPartialContent)
	}

	return nil
}
