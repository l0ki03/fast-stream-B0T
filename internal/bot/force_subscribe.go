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

	// ⚠️ Safe Stylish Font & NO LINKS IN TEXT
	msg := "🚀 𝗕𝗼𝘁 𝗸𝗮 𝗶𝘀𝘁𝗲𝗺𝗮𝗮𝗹 𝗸𝗮𝗿𝗻𝗲 𝗸𝗲 𝗹𝗶𝘆𝗲 𝗲𝗸 𝗰𝗵𝗵𝗼𝘁𝗮 𝘀𝗮 𝘀𝘁𝗲𝗽!\n\n" +
		"𝗣𝗲𝗵𝗹𝗲 𝗻𝗲𝗲𝗰𝗵𝗲 𝗱𝗶𝘆𝗲 𝗴𝗮𝘆𝗲 𝘀𝗮𝗯𝗵𝗶 𝗿𝗲𝗾𝘂𝗶𝗿𝗲𝗱 𝗰𝗵𝗮𝗻𝗻𝗲𝗹𝘀 𝗝𝗼𝗶𝗻 𝗸𝗮𝗿𝗻𝗮 𝘇𝗮𝗿𝗼𝗼𝗿𝗶 𝗵𝗮𝗶.\n" +
		"𝗨𝘀𝗸𝗲 𝗯𝗮𝗮𝗱 𝗮𝗽𝗻𝗶 𝗳𝗶𝗹𝗲 𝗱𝗼𝗯𝗮𝗿𝗮 𝘀𝗲𝗻𝗱 𝘆𝗮 𝗳𝗼𝗿𝘄𝗮𝗿𝗱 𝗸𝗮𝗿𝗲𝗶𝗻.\n" +
		"✨ 𝗙𝗶𝗿 𝗮𝗮𝗽𝗸𝗼 𝘁𝘂𝗿𝗮𝗻𝘁 𝗦𝘁𝗿𝗲𝗮𝗺 / 𝗗𝗼𝘄𝗻𝗹𝗼𝗮𝗱 𝗹𝗶𝗻𝗸 𝗺𝗶𝗹 𝗷𝗮𝘆𝗲𝗴𝗮.\n"

	var rows []tg.KeyboardButtonRow

	for _, username := range channels {

		link := fmt.Sprintf("https://t.me/%s", username)

		// 🛑 Yahan se maine msg me link jodne wala code hata diya hai 🛑

		// Button mein safe bold font (𝗝𝗢𝗜𝗡) aur link rahega
		rows = append(rows, tg.KeyboardButtonRow{
			Buttons: []tg.KeyboardButtonClass{
				&tg.KeyboardButtonURL{
					Text: "📢 𝗝𝗢𝗜𝗡 " + username,
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
