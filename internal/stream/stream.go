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

func NewTgFileReader(tgAPI *tg.Client, ctx context.Context, fileLocation *tg.InputDocumentFileLocation, file *types.File, req *http.Request) *TgFileReader {
	reader := &TgFileReader{
		ctx:          ctx,
		TgAPI:        tgAPI,
		FileLocation: fileLocation,
		File:         file,
		mu:           sync.RWMutex{},
	}
	return reader
}

func (r *TgFileReader) SetupStream(req *http.Request, w http.ResponseWriter, isDownload bool) error {
	if isDownload {
		w.Header().Set("Content-Disposition", "attachment; filename=\""+r.File.FileName+"\"")
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

		w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", r.start, r.end, r.File.Size))
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

func (r *TgFileReader) Read(p []byte) (n int, err error) {
	if r.isFinished() || r.start > r.end {
		return 0, io.EOF
	}
	if r.cachedChunk == nil || r.start < r.cachedOffset || r.start >= r.cachedOffset+int64(len(r.cachedChunk)) {
		chunkStart := (r.start / TelegramChunkSize) * TelegramChunkSize
		request := &tg.UploadGetFileRequest{
			Location: r.File.Location,
			Offset:   chunkStart,
			Limit:    TelegramChunkSize,
		}

		res, err := r.TgAPI.UploadGetFile(r.ctx, request)
		if err != nil {
			r.setFinished()
			return 0, err
		}

		file, ok := res.(*tg.UploadFile)
		if !ok {
			r.setFinished()
			return 0, fmt.Errorf("unable to cast")
		}

		r.cachedChunk = file.Bytes
		r.cachedOffset = chunkStart
	}

	positionInChunk := int(r.start - r.cachedOffset)

	availableBytes := len(r.cachedChunk) - positionInChunk
	if availableBytes <= 0 {
		return 0, io.EOF
	}

	bytesToCopy := min(len(p), availableBytes)
	n = copy(p, r.cachedChunk[positionInChunk:positionInChunk+bytesToCopy])
	r.start += int64(n)

	return n, nil
}
