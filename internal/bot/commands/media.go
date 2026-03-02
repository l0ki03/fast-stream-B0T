package commands

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/biisal/fast-stream-bot/config"
	botutils "github.com/biisal/fast-stream-bot/internal/bot/bot-utils"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/message/markup"
	"github.com/gotd/td/tg"
)

type MediaForwardParams struct {
	Cfg    *config.Config
	Update *tg.UpdateNewMessage
	Client *telegram.Client
}

func (bc *Context) MediaForwarding(params MediaForwardParams) (tg.UpdatesClass, error) {

	// 🔹 Credit Check
	if params.Cfg.REF {
		if bc.dbUser.Credit < params.Cfg.MIN_CREDITS_REQUIRED {

			referUrl := botutils.GetReferLink(bc.userInfo.Username, bc.userInfo.ID)

			now := time.Now()
			nextMidnight := time.Date(
				now.Year(),
				now.Month(),
				now.Day()+1,
				0, 0, 0, 0,
				now.Location(),
			)

			btn := markup.InlineKeyboard(
				markup.Row(
					markup.URL("Get Credits By Refer", referUrl),
				),
			)

			msg := fmt.Sprintf(
				"You're out of credits! 😢\n\nYou need %d more credits.\n\nWait %s or refer a user to earn %d credits.",
				params.Cfg.MIN_CREDITS_REQUIRED-bc.dbUser.Credit,
				nextMidnight.Sub(now).Round(time.Second).String(),
				params.Cfg.INCREMENT_CREDITS,
			)

			return bc.builder.Markup(btn).Text(bc.ctx, msg)
		}
	}

	// 🔹 Safe Message Cast
	if params.Update == nil || params.Update.Message == nil {
		return nil, fmt.Errorf("invalid update or message")
	}

	msgObj, ok := params.Update.Message.(*tg.Message)
	if !ok {
		return nil, fmt.Errorf("not a valid tg.Message")
	}
	m := msgObj

	fromPeer := &tg.InputPeerUser{
		UserID:     bc.userInfo.ID,
		AccessHash: bc.userInfo.AccessHash,
	}

	file, err := botutils.GetMediaFromMessage(m)
	if err != nil {
		slog.Error("Failed to get media", "error", err)
		return nil, err
	}

	msgHash := botutils.MakeHashByFileInfo(file)

	// 🔹 Forward to DB Channel
	_, channelInputPeer, err := botutils.GetChannelPeer(
		params.Client.API(),
		bc.ctx,
		params.Cfg.DB_CHANNEL_ID,
	)
	if err != nil {
		slog.Error("Channel peer error", "error", err)
		return nil, err
	}

	fUpdate, err := bc.sender.
		To(channelInputPeer.InputPeer()).
		ForwardIDs(fromPeer, m.ID).
		Send(bc.ctx)
	if err != nil {
		slog.Error("Forward failed", "error", err)
		return nil, err
	}

	// 🔹 Safe Message ID Extraction
	updates, ok := fUpdate.(*tg.Updates)
	if !ok {
		return nil, fmt.Errorf("invalid update type")
	}

	var messageId int
	for _, u := range updates.Updates {
		if msgID, ok := u.(*tg.UpdateMessageID); ok {
			messageId = msgID.ID
			break
		}
	}

	if messageId == 0 {
		return nil, fmt.Errorf("message ID not found")
	}

	// ✅ Watch Link
	streamLink := fmt.Sprintf(
		"%s/watch/%d?hash=%s",
		params.Cfg.FQDN,
		messageId,
		msgHash,
	)

	// ✅ Download Link
	downloadLink := fmt.Sprintf(
		"%s/stream/%d/%d/%s?d=1",
		params.Cfg.FQDN,
		params.Cfg.DB_CHANNEL_ID,
		messageId,
		msgHash,
	)

	// 🔹 Deduct Credit
	bc.dbUser, err = bc.userService.DecrementCredits(
		bc.ctx,
		bc.userInfo.ID,
		params.Cfg.DECREMENT_CREDITS,
	)
	if err != nil {
		slog.Error("Credit decrement failed", "error", err)
		return nil, err
	}

	// 🔹 User Message
	msg := fmt.Sprintf(
		"🎉 Your file is ready!\n\n📂 Name: %s\n📦 Size: %s\n\nChoose below:",
		file.FileName,
		botutils.MakeSizeReadable(file.Size),
	)

	if params.Cfg.REF {
		msg += fmt.Sprintf("\n\n💳 Credits left: %d", bc.dbUser.Credit)
	}

	btn := markup.InlineKeyboard(
		markup.Row(
			markup.URL("▶ Watch Now", streamLink),
		),
		markup.Row(
			markup.URL("⬇ Download Now", downloadLink),
		),
	)

	if _, err = bc.builder.Markup(btn).Text(bc.ctx, msg); err != nil {

		slog.Error("Send message failed", "error", err)

		// rollback credit
		_, _ = bc.userService.IncrementCredits(
			bc.ctx,
			bc.userInfo.ID,
			params.Cfg.DECREMENT_CREDITS,
			false,
		)

		return nil, err
	}

	// 🔹 Update total links
	bc.dbUser, _ = bc.userService.
		IncrementTotalLinkCount(bc.ctx, bc.dbUser.ID)

	// 🔹 Delete original message (if not admin)
	if bc.userInfo.ID != params.Cfg.ADMIN_ID {
		_, _ = params.Client.API().
			MessagesDeleteMessages(
				bc.ctx,
				&tg.MessagesDeleteMessagesRequest{
					Revoke: true,
					ID:     []int{m.ID},
				},
			)
	}

	// 🔹 Log in DB channel
	logMsg := fmt.Sprintf(
		"User: %s\nUserID: %d\n\nFile: %s\nSize: %s",
		bc.userInfo.Username,
		bc.userInfo.ID,
		file.FileName,
		botutils.MakeSizeReadable(file.Size),
	)

	return bc.sender.
		To(channelInputPeer.InputPeer()).
		Reply(messageId).
		Text(bc.ctx, logMsg)
}
