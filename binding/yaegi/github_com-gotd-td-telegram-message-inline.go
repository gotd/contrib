// Code generated by 'yaegi extract github.com/gotd/td/telegram/message/inline'. DO NOT EDIT.

package yaegi

import (
	"go/constant"
	"go/token"
	"reflect"

	"github.com/gotd/td/telegram/message/inline"
)

func init() {
	Symbols["github.com/gotd/td/telegram/message/inline"] = map[string]reflect.Value{
		// function, constant and variable definitions
		"Article":           reflect.ValueOf(inline.Article),
		"ArticleType":       reflect.ValueOf(constant.MakeFromLiteral("\"article\"", token.STRING, 0)),
		"Audio":             reflect.ValueOf(inline.Audio),
		"AudioType":         reflect.ValueOf(constant.MakeFromLiteral("\"audio\"", token.STRING, 0)),
		"ContactType":       reflect.ValueOf(constant.MakeFromLiteral("\"contact\"", token.STRING, 0)),
		"Document":          reflect.ValueOf(inline.Document),
		"DocumentType":      reflect.ValueOf(constant.MakeFromLiteral("\"document\"", token.STRING, 0)),
		"File":              reflect.ValueOf(inline.File),
		"GIF":               reflect.ValueOf(inline.GIF),
		"GIFType":           reflect.ValueOf(constant.MakeFromLiteral("\"gif\"", token.STRING, 0)),
		"Game":              reflect.ValueOf(inline.Game),
		"GameType":          reflect.ValueOf(constant.MakeFromLiteral("\"game\"", token.STRING, 0)),
		"LocationType":      reflect.ValueOf(constant.MakeFromLiteral("\"location\"", token.STRING, 0)),
		"MPEG4GIF":          reflect.ValueOf(inline.MPEG4GIF),
		"MPEG4GIFType":      reflect.ValueOf(constant.MakeFromLiteral("\"mpeg4_gif\"", token.STRING, 0)),
		"MediaAuto":         reflect.ValueOf(inline.MediaAuto),
		"MediaAutoStyled":   reflect.ValueOf(inline.MediaAutoStyled),
		"MessageGame":       reflect.ValueOf(inline.MessageGame),
		"MessageGeo":        reflect.ValueOf(inline.MessageGeo),
		"MessageStyledText": reflect.ValueOf(inline.MessageStyledText),
		"MessageText":       reflect.ValueOf(inline.MessageText),
		"New":               reflect.ValueOf(inline.New),
		"Photo":             reflect.ValueOf(inline.Photo),
		"PhotoType":         reflect.ValueOf(constant.MakeFromLiteral("\"photo\"", token.STRING, 0)),
		"ResultMessage":     reflect.ValueOf(inline.ResultMessage),
		"Sticker":           reflect.ValueOf(inline.Sticker),
		"StickerType":       reflect.ValueOf(constant.MakeFromLiteral("\"sticker\"", token.STRING, 0)),
		"VenueType":         reflect.ValueOf(constant.MakeFromLiteral("\"venue\"", token.STRING, 0)),
		"Video":             reflect.ValueOf(inline.Video),
		"VideoType":         reflect.ValueOf(constant.MakeFromLiteral("\"video\"", token.STRING, 0)),
		"Voice":             reflect.ValueOf(inline.Voice),
		"VoiceType":         reflect.ValueOf(constant.MakeFromLiteral("\"voice\"", token.STRING, 0)),

		// type definitions
		"ArticleResultBuilder":    reflect.ValueOf((*inline.ArticleResultBuilder)(nil)),
		"DocumentResultBuilder":   reflect.ValueOf((*inline.DocumentResultBuilder)(nil)),
		"GameResultBuilder":       reflect.ValueOf((*inline.GameResultBuilder)(nil)),
		"MessageGameBuilder":      reflect.ValueOf((*inline.MessageGameBuilder)(nil)),
		"MessageMediaAutoBuilder": reflect.ValueOf((*inline.MessageMediaAutoBuilder)(nil)),
		"MessageMediaGeoBuilder":  reflect.ValueOf((*inline.MessageMediaGeoBuilder)(nil)),
		"MessageOption":           reflect.ValueOf((*inline.MessageOption)(nil)),
		"MessageTextBuilder":      reflect.ValueOf((*inline.MessageTextBuilder)(nil)),
		"PhotoResultBuilder":      reflect.ValueOf((*inline.PhotoResultBuilder)(nil)),
		"ResultBuilder":           reflect.ValueOf((*inline.ResultBuilder)(nil)),
		"ResultOption":            reflect.ValueOf((*inline.ResultOption)(nil)),

		// interface wrapper definitions
		"_MessageOption": reflect.ValueOf((*_github_com_gotd_td_telegram_message_inline_MessageOption)(nil)),
		"_ResultOption":  reflect.ValueOf((*_github_com_gotd_td_telegram_message_inline_ResultOption)(nil)),
	}
}

// _github_com_gotd_td_telegram_message_inline_MessageOption is an interface wrapper for MessageOption type
type _github_com_gotd_td_telegram_message_inline_MessageOption struct {
}

// _github_com_gotd_td_telegram_message_inline_ResultOption is an interface wrapper for ResultOption type
type _github_com_gotd_td_telegram_message_inline_ResultOption struct {
}
