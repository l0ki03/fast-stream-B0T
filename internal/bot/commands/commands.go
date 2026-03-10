// Package commands contains commands of bot
package commands

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/biisal/fast-stream-bot/config"
	botutils "github.com/biisal/fast-stream-bot/internal/bot/bot-utils"
	repo "github.com/biisal/fast-stream-bot/internal/database/psql/sqlc"
	"github.com/biisal/fast-stream-bot/internal/service/user"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/markup"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/tg"
)

type Context struct {
	ctx         context.Context
	msg         *tg.Message
	entities    tg.Entities
	builder     *message.Builder
	userInfo    *user.TgUser
	dbUser      *repo.User
	userService user.Service
	sender      *message.Sender
	client      *telegram.Client
	cfg         *config.Config
}

func (bc *Context) Reply(msg string) (tg.UpdatesClass, error) {
	return bc.builder.Text(bc.ctx, msg)
}

func NewContext(ctx context.Context, msg *tg.Message,
	entities tg.Entities, builder *message.Builder,
	client *telegram.Client, sender *message.Sender,
	userInfo *user.TgUser, dbUser *repo.User, userService user.Service, cfg *config.Config,
) *Context {
	return &Context{
		ctx, msg, entities, builder, userInfo,
		dbUser, userService, sender, client, cfg,
	}
}

const helpMsg = `Using this bot is easy!
	
1. Send me a file.
2. I’ll create a link for it.
3. Open the link in your browser to stream or download the file instantly.`

func (bc *Context) HandleSendCommandList(adminID int64) (tg.UpdatesClass, error) {
	msg := `Available commands:
/start - Start the bot
/help - Get help
/stat - Get your statistics
/report - Replay a message to report to admin`

	if bc.userInfo.ID == adminID {
		msg += `/broadcast - Broadcast a message to all users

/ban - Ban a user
/unban - Unban a user`
	}

	return bc.Reply(msg)
}

// ⚠️ FIXED: Start Message with working Share Link
func (bc *Context) HandleStart() (tg.UpdatesClass, error) {
	username := "@" + bc.userInfo.Username
	if bc.userInfo.Username == "" {
		username = strings.TrimSpace(bc.userInfo.FirstName + " " + bc.userInfo.LastName)
	}
	msg := fmt.Sprintf(`ʜᴇʏ %s! 🔥 

⚡️ ꜰᴀꜱᴛ ꜱᴛʀᴇᴀᴍ ʙᴏᴛ ᴍᴇɪɴ ᴡᴇʟᴄᴏᴍᴇ! 

📂 ꜰɪʟᴇ ʙʜᴇᴊᴏ… 
🚀 ɪɴꜱᴛᴀɴᴛ ʟɪɴᴋ ʟᴏ… 
🎬 ꜱᴛʀᴇᴀᴍ ᴋᴀʀᴏ ʏᴀ 
📥 ᴅᴏᴡɴʟᴏᴀᴅ ᴋᴀʀᴏ — ꜰᴜʟʟ ꜱᴘᴇᴇᴅ!  

💡 ꜱɪᴍᴘʟᴇ. ꜰᴀꜱᴛ. ᴘᴏᴡᴇʀꜰᴜʟ. 

🚫 ᴀᴅᴜʟᴛ ʏᴀ ɪʟʟᴇɢᴀʟ ᴄᴏɴᴛᴇɴᴛ = ᴅɪʀᴇᴄᴛ ʙᴀɴ! ❌

`, username)

	// Telegram Native Share URL with text (Fixed URL and encoding)
	botUrl := "https://t.me/hmmediafiletolinkbot" 
	shareText := "Try%20this%20bot!%20Quickly%20stream%20or%20download%20your%20files%20with%20security%20and%20reliability."
	shareLink := fmt.Sprintf("https://t.me/share/url?url=%s&text=%s", botUrl, shareText)
	
	keyboard := markup.InlineKeyboard(
		markup.Row(
			markup.URL("ᴍᴏᴠɪᴇ ɢʀᴏᴜᴘ", "https://t.me/HMmedia_movie_group"),
			markup.URL("ᴊᴏɪɴ ᴜᴘᴅᴀᴛᴇ ᴄʜᴀɴɴᴇʟ", "https://t.me/HMmedia_Movie"),
		),
		markup.Row(
			markup.URL("× ꜱʜᴀʀᴇ ×", shareLink),
		),
	)
	
	return bc.builder.Markup(keyboard).Text(bc.ctx, msg)
}

func (bc *Context) HandleHelp(adminID int64) (tg.UpdatesClass, error) {
	commands := `Available commands:
/start - Start the bot
/help - Get help 
/stat - Get your statistics
`
	if bc.userInfo.ID == adminID {
		commands += `/broadcast - Broadcast a message to all users
/ban - Ban a user 
/unban - Unban a user`
	}

	fullMsg := fmt.Sprintf("%s\n\n%s", helpMsg, commands)
	return bc.Reply(fullMsg)
}

func (b *Context) SendLogMessage(msg string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, inputPeer, err := botutils.GetChannelPeer(b.client.API(), ctx, b.cfg.LOG_CHANNEL_ID)
	if err != nil {
		return err
	}
	b.sender.To(inputPeer.InputPeer()).Text(ctx, msg)
	return nil
}

func (b *Context) SendMainChannrlInviteLink(ctx context.Context, builder *message.Builder) (tg.UpdatesClass, error) {
	inviteLink, err := botutils.GetMainChannelInviteLink(ctx, b.client.API(), b.cfg)
	if err != nil {
		return nil, err
	}
	msg := `Due to server overload only our channel users can use this bot!
Please join our channel and continue using this bot :)`
	keyboard := markup.InlineKeyboard(
		markup.Row(
			markup.URL("Join Channel", inviteLink),
		),
	)

	return builder.Markup(keyboard).Text(ctx, msg)
}

func (bc *Context) HandleStat(adminID int64) (tg.UpdatesClass, error) {
	statMsg := fmt.Sprintf("Your statistics:\n\nTotal links: %d\nTotal credits: %d", bc.dbUser.TotalLinks,
		bc.dbUser.Credit)
	if bc.userInfo.ID == adminID {
		totalUserCount, err := bc.userService.GetUsersCount(bc.ctx)
		if err != nil {
			slog.Error("Failed to get total users count", "error", err)
			statMsg += fmt.Sprintf("\nFailed to get total users count! Err : %s\n", err.Error())
		}
		statMsg += fmt.Sprintf("\n\nTotal users: %d\n", totalUserCount)
	}
	return bc.Reply(statMsg)
}

func (bc *Context) ForwardMsgToLogChannel(replyedMessageID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, inputPeer, err := botutils.GetChannelPeer(bc.client.API(), ctx, bc.cfg.LOG_CHANNEL_ID)
	if err != nil {
		return err
	}

	fromPeer := &tg.InputPeerUser{UserID: bc.userInfo.ID, AccessHash: bc.userInfo.AccessHash}
	_, err = bc.sender.To(inputPeer.InputPeer()).ForwardIDs(fromPeer, replyedMessageID).Send(bc.ctx)
	return err
}

func (bc *Context) HandleReport(adminId int64) (tg.UpdatesClass, error) {
	if bc.userInfo.ID == adminId {
		return bc.Reply("not for admin")
	}
	var replyedMessageId int
	switch replay := bc.msg.ReplyTo.(type) {
	case *tg.MessageReplyHeader:
		replyedMessageId = replay.ReplyToMsgID
	case *tg.MessageReplyStoryHeader:
		return bc.Reply("can't report story")
	default:
		return bc.Reply("Reply a valid message to report")
	}

	fromPeer := &tg.InputPeerUser{UserID: bc.userInfo.ID, AccessHash: bc.userInfo.AccessHash}
	_, adminPeer, err := botutils.GetUserPeer(bc.client.API(), bc.ctx, adminId)
	if err != nil {
		if err = bc.ForwardMsgToLogChannel(replyedMessageId); err != nil {
			slog.Error("Failed to forward report message to log channel", "error", err)
			return bc.Reply("Failed to report your messsage! please try again later")
		}
		return bc.Reply("Reported your message to admin")
	}

	randomID := time.Now().UnixNano()
	_, err = bc.client.API().MessagesForwardMessages(bc.ctx, &tg.MessagesForwardMessagesRequest{
		FromPeer:   fromPeer,
		ToPeer:     adminPeer.InputPeer(),
		ID:         []int{replyedMessageId},
		RandomID:   []int64{randomID},
		DropAuthor: false,
	})
	if err != nil {
		if err = bc.ForwardMsgToLogChannel(replyedMessageId); err != nil {
			slog.Error("Failed to forward report message to log channel", "error", err)
			return bc.Reply("Failed to report your messsage! please try again later")
		}
		return bc.Reply("Reported your message to admin")
	}
	return bc.Reply("Reported your message to admin")
}

// ⚠️ FIXED: Unlimited (No Referrals) & Attractive Message Style
func (bc *Context) MediaForwarding(params MediaForwardParams) (tg.UpdatesClass, error) {

	// Safe Message Cast
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

	// Forward to DB Channel
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

	// Safe Message ID Extraction
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

	// Watch Link
	streamLink := fmt.Sprintf(
		"%s/watch/%d?hash=%s",
		params.Cfg.FQDN,
		messageId,
		msgHash,
	)

	// Download Link
	downloadLink := fmt.Sprintf(
		"%s/stream/%d/%d/%s?d=1",
		params.Cfg.FQDN,
		params.Cfg.DB_CHANNEL_ID,
		messageId,
		msgHash,
	)

	// User Message (Attractive Design)
	var textOpts []styling.StyledTextOption

	textOpts = append(textOpts,
		styling.Bold("► YOUR LINK GENERATED ! 😎\n\n"),
		styling.Bold("► FILE NAME : "), styling.Italic(file.FileName), styling.Plain("\n"),
		styling.Bold("► FILE SIZE : "), styling.Bold(botutils.MakeSizeReadable(file.Size)), styling.Plain("\n\n"),
		styling.Plain("► "), styling.TextURL("Support Us", "https://t.me/biisalbot"),
	)

	// Inline Buttons
	btn := markup.InlineKeyboard(
		markup.Row(
			markup.URL("STREAM 🔺", streamLink),
			markup.URL("DOWNLOAD 🔻", downloadLink),
		),
	)

	if _, err = bc.builder.Markup(btn).StyledText(bc.ctx, textOpts...); err != nil {
		slog.Error("Send message failed", "error", err)
		return nil, err
	}

	// Update total links count
	bc.dbUser, _ = bc.userService.
		IncrementTotalLinkCount(bc.ctx, bc.dbUser.ID)

	// Delete original message (if not admin)
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

	// Log in DB channel
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
