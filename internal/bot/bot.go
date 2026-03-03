// Package bot contains telegram bot
package bot

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/biisal/fast-stream-bot/config"
	botutils "github.com/biisal/fast-stream-bot/internal/bot/bot-utils"
	"github.com/biisal/fast-stream-bot/internal/bot/commands"
	repo "github.com/biisal/fast-stream-bot/internal/database/psql/sqlc"
	"github.com/biisal/fast-stream-bot/internal/service/user"
	"github.com/biisal/fast-stream-bot/internal/types"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/markup"
	"github.com/gotd/td/tg"
)

type Bot struct {
	WorkingPressure int
	Default         bool
	BotUserName     string
	Dispatcher      *tg.UpdateDispatcher
	Ctx             context.Context
	Client          *telegram.Client
	Sender          *message.Sender
	Cfg             *config.Config
	userService     user.Service
}

func NewBot(ctx context.Context, cfg *config.Config,
	client *telegram.Client, dispatcher *tg.UpdateDispatcher,
	userService user.Service, isDefault bool,
) *Bot {
	api := tg.NewClient(client)
	sender := message.NewSender(api)
	return &Bot{
		Default:     isDefault,
		Ctx:         ctx,
		Client:      client,
		Dispatcher:  dispatcher,
		Cfg:         cfg,
		Sender:      sender,
		userService: userService,
	}
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
			case val == "/broadcast":
				_, err = bc.HandleBroadcast(b.Cfg.ADMIN_ID)

			case val == "/help":
				_, err = bc.HandleHelp(b.Cfg.ADMIN_ID)

			case val == "/stat":
				_, err = bc.HandleStat(b.Cfg.ADMIN_ID)

			case strings.HasPrefix(val, "/unban"):
				_, err = bc.HandleToggleBan(b.Cfg.ADMIN_ID, false)

			case strings.HasPrefix(val, "/ban"):
				_, err = bc.HandleToggleBan(b.Cfg.ADMIN_ID, true)

			case strings.HasPrefix(val, "/report"):
				_, err = bc.HandleReport(b.Cfg.ADMIN_ID)

			default:

				// Unknown command except /start
				if strings.HasPrefix(val, "/") && !strings.HasPrefix(val, "/start") {
					_, err = bc.HandleSendCommandList(b.Cfg.ADMIN_ID)
					return err
				}

				// 🔥 FORCE SUBSCRIBE CHECK
				if len(b.Cfg.FORCE_SUB_CHANNELS) > 0 {

					userID := userInfo.ID

					joined := IsUserJoined(
						ctx,
						b.Client.API(),
						userID,
						b.Cfg.FORCE_SUB_CHANNELS,
					)

					if !joined {
						return SendForceSubscribeMessage(
							ctx,
							b.Client.API(),
							userID,
							b.Cfg.FORCE_SUB_CHANNELS,
						)
					}
				}

				// If joined → normal start
				_, err = bc.HandleStart()
			}

			return err
		}
	})
}
