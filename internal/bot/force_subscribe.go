package bot

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"

	"github.com/gotd/td/tg"
)

// Check if user joined all channels
func IsUserJoined(
	ctx context.Context,
	api *tg.Client,
	userID int64,
	channels []string,
) bool {

	for _, username := range channels {

		// 1. Username se channel ko dhundhna (Resolve)
		resolved, err := api.ContactsResolveUsername(ctx, &tg.ContactsResolveUsernameRequest{
			Username: username,
		})
		if err != nil {
			slog.Error("Failed to resolve channel", "username", username, "error", err)
			return false
		}

		if len(resolved.Chats) == 0 {
			slog.Error("Channel chats not found", "username", username)
			return false
		}

		channel, ok := resolved.Chats[0].(*tg.Channel)
		if !ok {
			return false
		}

		// 2. Ab asli ID aur AccessHash ka istemaal karke check karein
		participant, err := api.ChannelsGetParticipant(ctx, &tg.ChannelsGetParticipantRequest{
			Channel: &tg.InputChannel{
				ChannelID:  channel.ID,
				AccessHash: channel.AccessHash,
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
	channels []string,
) error {

	// ⚠️ Naya Stylish Font Message
	msg := "🚀 🇧🇴🇹 🇰🇦 🇮🇸🇹🇪🇲🇦🇦🇱 🇰🇦🇷🇳🇪 🇰🇪 🇱🇮🇾🇪 🇪🇰 🇨🇭🇭🇴🇹🇦 🇸🇦 🇸🇹🇪🇵!\n\n" +
		"🇵🇪🇭🇱🇪 🇳🇪🇪🇨🇭🇪 🇩🇮🇾🇪 🇬🇦🇾🇪 🇸🇦🇧🇭🇮 🇷🇪🇶🇺🇮🇷🇪🇩 🇨🇭🇦🇳🇳🇪🇱🇸 🇯🇴🇮🇳 🇰🇦🇷🇳🇦 🇿🇦🇷🇴🇴🇷🇮 🇭🇦🇮.\n" +
		"🇺🇸🇰🇪 🇧🇦🇦🇩 🇦🇵🇳🇮 🇫🇮🇱🇪 🇩🇴🇧🇦🇷🇦 🇸🇪🇳🇩 🇾🇦 🇫🇴🇷🇼🇦🇷🇩 🇰🇦🇷🇪.\n" +
		"✨ 🇫🇮🇷 🇦🇦🇵🇰🇴 🇹🇺🇷🇦🇳🇹 🇸🇹🇷🇪🇦🇲 / 🇩🇴🇼🇳🇱🇴🇦🇩 🇱🇮🇳🇰 🇲🇮🇱 🇯🇦🇾🇪🇬🇦.\n\n"

	var rows []tg.KeyboardButtonRow

	for _, username := range channels {

		link := fmt.Sprintf("https://t.me/%s", username)

		msg += fmt.Sprintf("👉 %s\n", link)

		// Button mein bhi wahi special font daal diya hai (🇯🇴🇮🇳 = JOIN)
		rows = append(rows, tg.KeyboardButtonRow{
			Buttons: []tg.KeyboardButtonClass{
				&tg.KeyboardButtonURL{
					Text: "📢 🇯🇴🇮🇳 " + username,
					URL:  link,
				},
			},
		})
	}

	_, err := api.MessagesSendMessage(ctx, &tg.MessagesSendMessageRequest{
		Peer: &tg.InputPeerUser{
			UserID: userID,
		},
		Message: msg,
		ReplyMarkup: &tg.ReplyInlineMarkup{
			Rows: rows, 
		},
		RandomID: rand.Int63(),
	})

	return err
}
