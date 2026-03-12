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

	// ⚠️ Safe Small Caps Font - No Flags Issue
	msg := "🚀 ʙᴏᴛ ᴋᴀ ɪsᴛᴇᴍᴀᴀʟ ᴋᴀʀɴᴇ ᴋᴇ ʟɪʏᴇ ᴇᴋ ᴄʜʜᴏᴛᴀ sᴀ sᴛᴇᴘ!\n\n" +
		"1. ᴘᴇʜʟᴇ ɴᴇᴇᴄʜᴇ ᴅɪʏᴇ ɢᴀʏᴇ sᴀʙʜɪ ʀᴇǫᴜɪʀᴇᴅ ᴄʜᴀɴɴᴇʟs ᴊᴏɪɴ ᴋᴀʀᴇɪɴ.\n" +
		"2. ᴜsᴋᴇ ʙᴀᴀᴅ ᴀᴘɴɪ ғɪʟᴇ ᴅᴏʙᴀʀᴀ sᴇɴᴅ ʏᴀ ғᴏʀᴡᴀʀᴅ ᴋᴀʀᴇɪɴ.\n\n" +
		"✨ ғɪʀ ᴀᴀᴘᴋᴏ ᴛᴜʀᴀɴᴛ sᴛʀᴇᴀᴍ / ᴅᴏᴡɴʟᴏᴀᴅ ʟɪɴᴋ ᴍɪʟ ᴊᴀʏᴇɢᴀ.\n"

	var rows []tg.KeyboardButtonRow

	// i ka use karke hum 1, 2, 3... count lagayenge
	for i, username := range channels {

		link := fmt.Sprintf("https://t.me/%s", username)

		// Button ka text "📢 ᴊᴏɪɴ ᴄʜᴀɴɴᴇʟ 1", "📢 ᴊᴏɪɴ ᴄʜᴀɴɴᴇʟ 2" wagaira set karna
		buttonText := fmt.Sprintf(" ᴊᴏɪɴ ᴄʜᴀɴɴᴇʟ %d", i+1)

		rows = append(rows, tg.KeyboardButtonRow{
			Buttons: []tg.KeyboardButtonClass{
				&tg.KeyboardButtonURL{
					Text: buttonText,
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
