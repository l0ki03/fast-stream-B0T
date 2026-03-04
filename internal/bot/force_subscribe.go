package bot

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gotd/td/tg"
)

// Check if user joined all required channels (username based)
func IsUserJoined(
	ctx context.Context,
	api *tg.Client,
	userID int64,
	channels []string, // 🔥 change to string (username)
) bool {

	for _, username := range channels {

		resolved, err := api.ContactsResolveUsername(ctx, username)
		if err != nil {
			slog.Info("Channel resolve failed", "channel", username, "error", err)
			return false
		}

		channel := resolved.Chats[0].(*tg.Channel)

		_, err = api.ChannelsGetParticipant(ctx, &tg.ChannelsGetParticipantRequest{
			Channel: &tg.InputChannel{
				ChannelID:  channel.ID,
				AccessHash: channel.AccessHash,
			},
			Participant: &tg.InputPeerUser{
				UserID: userID,
			},
		})

		if err != nil {
			slog.Info("User not joined channel",
				"user", userID,
				"channel", username,
				"error", err,
			)
			return false
		}
	}

	return true
}

func SendForceSubscribeMessage(
	ctx context.Context,
	api *tg.Client,
	userID int64,
	channels []string,
) error {

	text := "🚨 You must join all required channels to continue:\n\n"

	var rows []tg.KeyboardButtonRow

	for _, username := range channels {

		link := fmt.Sprintf("https://t.me/%s", username)

		button := &tg.KeyboardButtonURL{
			Text: "Join Channel",
			URL:  link,
		}

		rows = append(rows, tg.KeyboardButtonRow{
			Buttons: []tg.KeyboardButtonClass{button},
		})
	}

	_, err := api.MessagesSendMessage(ctx, &tg.MessagesSendMessageRequest{
		Peer: &tg.InputPeerUser{
			UserID: userID,
		},
		Message: text,
		RandomID: 0,
	})

	return err
}
