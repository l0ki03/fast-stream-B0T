func (bc *Context) MediaForwarding(params MediaForwardParams) (tg.UpdatesClass, error) {
	if params.Cfg.REF {
		if bc.dbUser.Credit < params.Cfg.MIN_CREDITS_REQUIRED {
			referUrl := botutils.GetReferLink(bc.userInfo.Username, bc.userInfo.ID)
			now := time.Now()
			nextMidnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())

			btn := markup.InlineKeyboard(
				markup.Row(
					markup.URL("Get Credits By Refer", referUrl),
				),
			)

			msg := fmt.Sprintf(
				"You're out of credits! 😢\nYou need %d more credits to use this bot.\n\nWait for %s to get new credits or Refer one user to earn %d credits.",
				params.Cfg.MIN_CREDITS_REQUIRED-bc.dbUser.Credit,
				nextMidnight.Sub(now).Round(time.Second).String(),
				params.Cfg.INCREMENT_CREDITS,
			)

			return bc.builder.Markup(btn).Text(bc.ctx, msg)
		}
	}

	m := params.Update.Message.(*tg.Message)
	fromPeer := &tg.InputPeerUser{UserID: bc.userInfo.ID, AccessHash: bc.userInfo.AccessHash}

	file, err := botutils.GetMediaFromMessage(m)
	if err != nil {
		slog.Error("Failed to get media from message", "error", err)
		return nil, err
	}

	msgHash := botutils.MakeHashByFileInfo(file)

	_, channelInputPeer, err := botutils.GetChannelPeer(params.Client.API(), bc.ctx, params.Cfg.DB_CHANNEL_ID)
	if err != nil {
		slog.Error("Failed to get channel peer", "error", err)
		return nil, err
	}

	fUpdate, err := bc.sender.To(channelInputPeer.InputPeer()).ForwardIDs(fromPeer, m.ID).Send(bc.ctx)
	if err != nil {
		slog.Error("Failed to forward message", "error", err)
		return nil, err
	}

	messageId := fUpdate.(*tg.Updates).Updates[0].(*tg.UpdateMessageID).ID

	// ✅ Generate both links
	streamLink := fmt.Sprintf("%s/watch/%d?hash=%s", params.Cfg.FQDN, messageId, msgHash)
	downloadLink := fmt.Sprintf("%s/download/%d?hash=%s", params.Cfg.FQDN, messageId, msgHash)

	fileMsg := fmt.Sprintf(
		"File Name: %s\nFile Size: %s",
		file.FileName,
		botutils.MakeSizeReadable(file.Size),
	)

	bc.dbUser, err = bc.userService.DecrementCredits(bc.ctx, bc.userInfo.ID, params.Cfg.DECREMENT_CREDITS)
	if err != nil {
		slog.Error("Failed to decrement credit", "error", err)
		return nil, err
	}

	msg := fmt.Sprintf(
		"Your file is ready! 🎉\n\n%s\n\nChoose an option below:",
		fileMsg,
	)

	if params.Cfg.REF {
		msg += fmt.Sprintf("\n\nYou have %d credits remaining 😊", bc.dbUser.Credit)
	}

	// ✅ 2 Buttons
	btn := markup.InlineKeyboard(
		markup.Row(
			markup.URL("▶ Watch Now", streamLink),
			markup.URL("⬇ Download Now", downloadLink),
		),
	)

	if _, err = bc.builder.Markup(btn).Text(bc.ctx, msg); err != nil {
		slog.Error("Failed to send message", "error", err)

		if _, err := bc.userService.IncrementCredits(bc.ctx, bc.userInfo.ID, params.Cfg.DECREMENT_CREDITS, false); err != nil {
			slog.Error("Failed to increment credit", "error", err)
		}

		return nil, err
	}

	if bc.dbUser, err = bc.userService.IncrementTotalLinkCount(bc.ctx, bc.dbUser.ID); err != nil {
		slog.Error("Failed to update user", "error", err)
	}

	// Log message
	fileMsg = fmt.Sprintf("User: %s\nUserId: %d\n\nFile Name: %s\nFile Size: %s",
		bc.userInfo.Username,
		bc.userInfo.ID,
		file.FileName,
		botutils.MakeSizeReadable(file.Size),
	)

	if bc.userInfo.ID != params.Cfg.ADMIN_ID {
		_, err = params.Client.API().MessagesDeleteMessages(bc.ctx, &tg.MessagesDeleteMessagesRequest{
			Revoke: true,
			ID:     []int{m.ID},
		})
		if err != nil {
			slog.Error("Failed to delete user message", "error", err)
		}
	}

	return bc.sender.To(channelInputPeer.InputPeer()).Reply(messageId).Text(bc.ctx, fileMsg)
}
