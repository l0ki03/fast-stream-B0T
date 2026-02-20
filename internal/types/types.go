package types

import (
	"fmt"

	"github.com/gotd/td/tg"
)

type HashInfo struct {
	Hash      string `json:"hash"`
	MessageId int    `json:"message_id"`
	ChannelId int64  `json:"channel_id"`
}

type HashResponse struct {
	Data       HashInfo `json:"data"`
	StatusCode int      `json:"status_code"`
}

type File struct {
	Location   *tg.InputDocumentFileLocation
	Size       int64
	AccessHash int64
	MimeType   string
	FileName   string
}

type BroadcastState struct {
	CompletedCountChan chan int
	DoneChan           chan bool
	FailedCountChan    chan int
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type FileResponse struct {
	Title          string
	Size           string
	DownloadLink   string
	StreamLink     string
	IsJustVerified bool
	ExpireTime     string
	AppName        string
}

var (
	ErrorNotFound       = fmt.Errorf("data not found")
	ErrorInternalServer = fmt.Errorf("internal server error. Please try again later or report the issue to developer")
	ErrorDuplicate      = fmt.Errorf("user already exists")
)
