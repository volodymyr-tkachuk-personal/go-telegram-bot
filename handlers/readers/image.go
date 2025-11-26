package readers

import (
	"context"
	"time"

	"github.com/ufy-it/go-telegram-bot/handlers/buttons"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// UserImageAndDataReply contains image input from a user
type UserImageAndDataReply struct {
	Exit  bool
	Data  string
	Image []tgbotapi.PhotoSize
}

type UserImagesAndDataReply struct {
	Exit   bool
	Data   string
	Images [][]tgbotapi.PhotoSize
}

// GetImage asks a user to send image
func GetImage(ctx context.Context, conversation BotConversation, text string, navigation buttons.ButtonSet, textOnIncorrect string) (UserImageAndDataReply, error) {
	msg := conversation.NewMessage(text)
	validator := func(update *tgbotapi.Update) (bool, string) {
		if update != nil && update.Message != nil && update.Message.Photo != nil && len(update.Message.Photo) > 0 {
			return true, ""
		}
		return false, textOnIncorrect
	}
	reply, exit, err := AskGenericMessageReplyWithValidation(ctx, conversation, msg, navigation, validator, true)
	result := UserImageAndDataReply{Exit: exit}
	if reply != nil && reply.CallbackQuery != nil {
		result.Data = reply.CallbackQuery.Data
	}
	if reply != nil && reply.Message != nil && reply.Message.Photo != nil && len(reply.Message.Photo) > 0 {
		result.Image = reply.Message.Photo
	}
	return result, err
}

const MediaGroupAdditionalPhotoWaitTimeout = 2 * time.Second

// GetImages asks a user to send multiple images in one media group
func GetImages(ctx context.Context, conversation BotConversation, text string, navigation buttons.ButtonSet, textOnIncorrect string) (UserImagesAndDataReply, error) {
	msg := conversation.NewMessage(text)
	validator := func(update *tgbotapi.Update) (bool, string) {
		if update != nil && update.Message != nil && update.Message.MediaGroupID != "" && update.Message.Photo != nil && len(update.Message.Photo) > 0 {
			return true, ""
		}
		return false, textOnIncorrect
	}
	reply, exit, err := AskGenericMessageReplyWithValidation(ctx, conversation, msg, navigation, validator, true)
	result := UserImagesAndDataReply{
		Exit:   exit,
		Images: make([][]tgbotapi.PhotoSize, 0),
	}
	if err != nil {
		return result, err
	}
	if reply != nil && reply.CallbackQuery != nil {
		result.Data = reply.CallbackQuery.Data
	}
	if reply != nil && reply.Message != nil && len(reply.Message.Photo) > 0 {
		result.Images = append(result.Images, reply.Message.Photo)
		mediaGroupID := reply.Message.MediaGroupID
		for {
			update, isExit, isTimeout := conversation.GetUpdateFromUserWithTimeout(ctx, MediaGroupAdditionalPhotoWaitTimeout)
			if isExit {
				result.Exit = true
				break
			}
			if isTimeout {
				break
			}
			if update != nil && update.Message != nil && update.Message.MediaGroupID == mediaGroupID && update.Message.Photo != nil && len(update.Message.Photo) > 0 {
				result.Images = append(result.Images, update.Message.Photo)
			} else {
				// no part of a media group, will be lost for now
				break
			}
		}
	}
	return result, err
}
