package bot

import (
	"context"
	"fmt"
)

type ForceSubscribe struct {
	API BotAPI
}

func (f *ForceSubscribe) IsJoined(userID int64, channels []int64) bool {

	for _, channelID := range channels {

		member, err := f.API.GetChatMember(context.Background(), channelID, userID)
		if err != nil {
			return false
		}

		status := member.GetStatus()

		if status != "member" &&
			status != "administrator" &&
			status != "creator" {
			return false
		}
	}

	return true
}

func (f *ForceSubscribe) SendJoinMessage(userID int64, channels []int64) {

	text := "🚨 You must join all required channels to continue:\n\n"

	var keyboard [][]InlineKeyboardButton

	for _, ch := range channels {

		chat, err := f.API.GetChat(context.Background(), ch)
		if err != nil {
			continue
		}

		link := fmt.Sprintf("https://t.me/%s", chat.Username)

		btn := InlineKeyboardButton{
			Text: "Join " + chat.Title,
			URL:  link,
		}

		keyboard = append(keyboard, []InlineKeyboardButton{btn})
	}

	checkBtn := InlineKeyboardButton{
		Text: "✅ I Joined – Check Again",
		Data: "force_check",
	}

	keyboard = append(keyboard, []InlineKeyboardButton{checkBtn})

	f.API.SendMessage(userID, text, keyboard)
}
