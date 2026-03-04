package bot

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/biisal/fast-stream-bot/config"
	"github.com/biisal/fast-stream-bot/internal/bot/commands"
	repo "github.com/biisal/fast-stream-bot/internal/database/psql/sqlc"
	"github.com/biisal/fast-stream-bot/internal/service/user"
	"github.com/biisal/fast-stream-bot/internal/types"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/tg"
)

type Bot struct {
	Default         bool
	BotUserName     string
	Dispatcher      *tg.UpdateDispatcher
	Ctx             context.Context
	Client          *telegram.Client
	Sender          *message.Sender
	Cfg             *config.Config
	userService     user.Service

	WorkingPressure int // ✅ FIXED (int instead of int32)
}

func NewBot(
	ctx context.Context,
	cfg *config.Config,
	client *telegram.Client,
	dispatcher *tg.UpdateDispatcher,
	userService user.Service,
	isDefault bool,
) *Bot {

	api := tg.NewClient(client)
	sender := message.NewSender(api)

	return &Bot{
		Default:         isDefault,
		Ctx:             ctx,
		Client:          client,
		Dispatcher:      dispatcher,
		Cfg:             cfg,
		Sender:          sender,
		userService:     userService,
		WorkingPressure: 0,
	}
}

func (b *Bot) validateAndGetUser(
	ctx context.Context,
	m *tg.Message,
	e tg.Entities,
	builder *message.Builder,
) (*user.TgUser, *repo.User, error) {

	userInfo := b.userService.GetUserInfo(ctx, m, e)

	dbUser, err := b.userService.GetUserByTgID(ctx, userInfo.ID)
	if err != nil {

		if !errors.Is(err, types.ErrorNotFound) {
			slog.Error("Failed to get user", "error", err)
			return nil, nil, fmt.Errorf("internal server error")
		}

		dbUser, err = b.userService.CreateUser(ctx, repo.CreateUserParams{
			ID:     userInfo.ID,
			Credit: b.Cfg.INITIAL_CREDITS,
		})

		if err != nil && !errors.Is(err, types.ErrorDuplicate) {
			slog.Error("Failed to create user", "error", err)
			return nil, nil, fmt.Errorf("internal server error")
		}
	}

	if dbUser.IsBanned {
		builder.Text(ctx, "🚫 You are banned to use this bot.")
		return nil, nil, fmt.Errorf("user banned")
	}

	return userInfo, dbUser, nil
}

func (b *Bot) SetUpOnMessage() {

	b.Dispatcher.OnNewMessage(func(ctx context.Context, e tg.Entities, update *tg.UpdateNewMessage) error {

		m, ok := update.Message.(*tg.Message)
		if !ok || m.Out {
			return nil
		}

		builder := b.Sender.Reply(e, update)

		userInfo, dbUser, err := b.validateAndGetUser(ctx, m, e, builder)
		if err != nil {
			return err
		}

		// ✅ FORCE SUBSCRIBE CHECK
		if len(b.Cfg.FORCE_SUB_CHANNELS) > 0 {

			joined := IsUserJoined(
				ctx,
				b.Client.API(),
				userInfo.ID,
				b.Cfg.FORCE_SUB_CHANNELS,
			)

			if !joined {
				return SendForceSubscribeMessage(
					ctx,
					b.Client.API(),
					userInfo.ID,
					b.Cfg.FORCE_SUB_CHANNELS,
				)
			}
		}

		bc := commands.NewContext(
			ctx,
			m,
			e,
			builder,
			b.Client,
			b.Sender,
			userInfo,
			dbUser,
			b.userService,
			b.Cfg,
		)

		switch m.Media.(type) {

		case *tg.MessageMediaDocument, *tg.MessageMediaPhoto:

			_, err = bc.MediaForwarding(commands.MediaForwardParams{
				Cfg:    b.Cfg,
				Update: update,
				Client: b.Client,
			})

			if err != nil {
				slog.Error("Failed to forward message", "error", err)
			}
			return err

		default:

			val := strings.TrimSpace(m.Message)

			switch {

			case val == "/help":
				_, err = bc.HandleHelp(b.Cfg.ADMIN_ID)

			case val == "/stat":
				_, err = bc.HandleStat(b.Cfg.ADMIN_ID)

			case strings.HasPrefix(val, "/ban"):
				_, err = bc.HandleToggleBan(b.Cfg.ADMIN_ID, true)

			case strings.HasPrefix(val, "/unban"):
				_, err = bc.HandleToggleBan(b.Cfg.ADMIN_ID, false)

			default:
				_, err = bc.HandleStart()
			}

			return err
		}
	})
}
