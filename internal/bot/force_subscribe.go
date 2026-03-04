package bot

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gotd/td/tg"
)

// Check if user joined all channels
func IsUserJoined(
	ctx context.Context,
	api *tg.Client,
	userID int64,
	channels []int64,
) bool {

	for _, channelID := range channels {

		participant, err := api.ChannelsGetParticipant(ctx, &tg.ChannelsGetParticipantRequest{
			Channel: &tg.InputChannel{
				ChannelID: channelID,
				AccessHash: 0, // ⚠️ Important: If public channel, 0 works
			},
			Participant: &tg.InputPeerUser{
				UserID: userID,
			},
		})

		if err != nil {
			slog.Info("User not joined channel", "user", userID, "channel", channelID, "error", err)
			return false
		}

		if participant == nil {
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

	msg := "🚨 Please join the required channels to use this bot:\n\n"

	var buttons [][]tg.KeyboardButtonClass

	for _, channelID := range channels {

		link := fmt.Sprintf("https://t.me/c/%d", channelID)

		msg += fmt.Sprintf("👉 %s\n", link)

		buttons = append(buttons, []tg.KeyboardButtonClass{
			&tg.KeyboardButtonURL{
				Text: "Join Channel",
				URL:  link,
			},
		})
	}

	_, err := api.MessagesSendMessage(ctx, &tg.MessagesSendMessageRequest{
		Peer: &tg.InputPeerUser{
			UserID: userID,
		},
		Message: msg,
		ReplyMarkup: &tg.ReplyInlineMarkup{
			Rows: []tg.KeyboardButtonRow{
				{
					Buttons: buttons[0],
				},
			},
		},
		RandomID: 0,
	})

	return err
}
