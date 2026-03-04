package bot

import (
	"context"
	"log/slog"

	"github.com/gotd/td/tg"
)

func IsUserJoined(
	ctx context.Context,
	api *tg.Client,
	userID int64,
	channels []int64,
) bool {

	for _, channelID := range channels {

		// Resolve channel entity
		full, err := api.ChannelsGetFullChannel(ctx, &tg.ChannelsGetFullChannelRequest{
			Channel: &tg.InputChannel{
				ChannelID:  channelID,
				AccessHash: 0,
			},
		})

		if err != nil {
			slog.Info("Channel resolve failed",
				"channel", channelID,
				"error", err,
			)
			return false
		}

		channel := full.Chats[0].(*tg.Channel)

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
				"channel", channelID,
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
	channels []int64,
) error {

	text := "🚨 You must join required channels to use this bot."

	_, err := api.MessagesSendMessage(ctx, &tg.MessagesSendMessageRequest{
		Peer: &tg.InputPeerUser{
			UserID: userID,
		},
		Message: text,
		RandomID: 0,
	})

	return err
}
