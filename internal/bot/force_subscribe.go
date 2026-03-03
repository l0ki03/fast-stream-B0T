package bot

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gotd/td/tg"
)

// Check if user joined all required channels
func IsUserJoined(
	ctx context.Context,
	api *tg.Client,
	userID int64,
	channels []int64,
) bool {

	for _, channelID := range channels {

		_, err := api.ChannelsGetParticipant(ctx, &tg.ChannelsGetParticipantRequest{
			Channel: &tg.InputChannel{
				ChannelID:  channelID,
				AccessHash: 0,
			},
			Participant: &tg.InputPeerUser{
				UserID:     userID,
				AccessHash: 0,
			},
		})

		if err != nil {
			slog.Info("User not joined channel",
				"user", userID,
				"channel", channelID,
				"error", err,
			)
			return false
		}
	}

	return true
}

// Send force subscribe message
func SendForceSubscribeMessage(
	ctx context.Context,
	api *tg.Client,
	userID int64,
	channels []int64,
) error {

	text := "🚨 You must join all required channels to continue:\n\n"

	var rows []tg.KeyboardButtonRow

	for _, channelID := range channels {

		link := fmt.Sprintf("https://t.me/c/%d", -channelID)

		button := &tg.KeyboardButtonURL{
			Text: "Join Channel",
			URL:  link,
		}

		rows = append(rows, tg.KeyboardButtonRow{
			Buttons: []tg.KeyboardButtonClass{button},
		})
	}

	// Recheck button
	checkButton := &tg.KeyboardButtonCallback{
		Text: "✅ I Joined – Check Again",
		Data: []byte("force_check"),
	}

	rows = append(rows, tg.KeyboardButtonRow{
		Buttons: []tg.KeyboardButtonClass{checkButton},
	})

	_, err := api.MessagesSendMessage(ctx, &tg.MessagesSendMessageRequest{
		Peer: &tg.InputPeerUser{
			UserID:     userID,
			AccessHash: 0,
		},
		Message: text,
		ReplyMarkup: &tg.ReplyInlineMarkup{
			Rows: rows,
		},
		RandomID: 0,
	})

	return err
}
