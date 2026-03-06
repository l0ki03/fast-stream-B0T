package bot

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand" // Telegram API को मैसेज भेजने के लिए RandomID चाहिए होता है

	"github.com/gotd/td/tg"
)

// Check if user joined all channels
func IsUserJoined(
	ctx context.Context,
	api *tg.Client,
	userID int64,
	channels []string, // int64 से बदलकर string (Username) कर दिया गया है
) bool {

	for _, username := range channels {

		// 1. Username से चैनल को ढूँढना (Resolve) ताकि असली AccessHash मिल सके
		peer, err := api.ContactsResolveUsername(ctx, username)
		if err != nil || len(peer.Chats) == 0 {
			slog.Error("Failed to resolve channel", "username", username, "error", err)
			return false
		}

		channel, ok := peer.Chats[0].(*tg.Channel)
		if !ok {
			return false
		}

		// 2. अब असली ID और AccessHash का इस्तेमाल करके चेक करें
		participant, err := api.ChannelsGetParticipant(ctx, &tg.ChannelsGetParticipantRequest{
			Channel: &tg.InputChannel{
				ChannelID:  channel.ID,
				AccessHash: channel.AccessHash, // असली AccessHash (जिससे CHANNEL_INVALID एरर नहीं आएगा)
			},
			Participant: &tg.InputPeerUser{
				UserID: userID,
			},
		})

		if err != nil {
			slog.Info("User not joined channel", "user", userID, "channel", username, "error", err)
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
	channels []string, // int64 से बदलकर string कर दिया गया है
) error {

	msg := "🚨 Please join the required channels to use this bot:\n\n"

	var buttons [][]tg.KeyboardButtonClass

	for _, username := range channels {

		// Username के आधार पर लिंक बनाना
		link := fmt.Sprintf("https://t.me/%s", username)

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
		RandomID: rand.Int63(), // ⚠️ IMPORTANT: इसके बिना Telegram (gotd) मैसेज को इग्नोर कर देता है
	})

	return err
}
