package main

import (
	"discocheckbot/api"
	"strconv"
	"strings"
	"time"
)

func getSkillMessage(cmd string, chatId int64, color int) api.SendMessage {
	chkColor := int64(color)
	smsg := api.SendMessage{
		ChatID: chatId,
		Text:   "select skill:",
		ReplyMarkup: &api.InlineKeyboardMarkup{
			InlineKeyboard: [][]api.InlineKeyboardButton{
				{{Text: skillNames[intLogic], CallbackData: makeClbk(cmd, chkColor, intLogic)},
					{Text: skillNames[intEncyclopedia], CallbackData: makeClbk(cmd, chkColor, intEncyclopedia)}},
				{{Text: skillNames[intRhetoric], CallbackData: makeClbk(cmd, chkColor, intRhetoric)},
					{Text: skillNames[intDrama], CallbackData: makeClbk(cmd, chkColor, intDrama)}},
				{{Text: skillNames[intConcept], CallbackData: makeClbk(cmd, chkColor, intConcept)},
					{Text: skillNames[intVisual], CallbackData: makeClbk(cmd, chkColor, intEncyclopedia)}},
				{{Text: skillNames[psyVolition], CallbackData: makeClbk(cmd, chkColor, psyVolition)},
					{Text: skillNames[psyInland], CallbackData: makeClbk(cmd, chkColor, psyInland)}},
				{{Text: skillNames[psyEmpathy], CallbackData: makeClbk(cmd, chkColor, psyEmpathy)},
					{Text: skillNames[psyAuthority], CallbackData: makeClbk(cmd, chkColor, psyAuthority)}},
				{{Text: skillNames[psyEsprit], CallbackData: makeClbk(cmd, chkColor, psyEsprit)},
					{Text: skillNames[psySuggestion], CallbackData: makeClbk(cmd, chkColor, psySuggestion)}},
				{{Text: skillNames[phyEndurance], CallbackData: makeClbk(cmd, chkColor, phyEndurance)},
					{Text: skillNames[phyPain], CallbackData: makeClbk(cmd, chkColor, phyPain)}},
				{{Text: skillNames[phyInstrument], CallbackData: makeClbk(cmd, chkColor, phyInstrument)},
					{Text: skillNames[phyElectrochem], CallbackData: makeClbk(cmd, chkColor, phyElectrochem)}},
				{{Text: skillNames[phyShivers], CallbackData: makeClbk(cmd, chkColor, phyShivers)},
					{Text: skillNames[phyHalflight], CallbackData: makeClbk(cmd, chkColor, phyHalflight)}},
				{{Text: skillNames[motCoordintation], CallbackData: makeClbk(cmd, chkColor, motCoordintation)},
					{Text: skillNames[motPerception], CallbackData: makeClbk(cmd, chkColor, motPerception)}},
				{{Text: skillNames[motReaction], CallbackData: makeClbk(cmd, chkColor, motReaction)},
					{Text: skillNames[motSavoir], CallbackData: makeClbk(cmd, chkColor, motSavoir)}},
				{{Text: skillNames[motInterfacing], CallbackData: makeClbk(cmd, chkColor, motInterfacing)},
					{Text: skillNames[motComposure], CallbackData: makeClbk(cmd, chkColor, motComposure)}},
			},
		},
	}
	return smsg
}

func getSkillDifEditMessage(chatId int64, msgId int, clbk string) api.EditMessageText {
	emsg := api.EditMessageText{
		ChatID:    chatId,
		MessageID: msgId,
		Text:      "select check difficulty:",
		ReplyMarkup: &api.InlineKeyboardMarkup{
			InlineKeyboard: [][]api.InlineKeyboardButton{
				{{Text: difficultyNames[difTrivial], CallbackData: makeClbk(clbk, difTrivial)}},
				{{Text: difficultyNames[difEasy], CallbackData: makeClbk(clbk, difEasy)}},
				{{Text: difficultyNames[difMedium], CallbackData: makeClbk(clbk, difMedium)}},
				{{Text: difficultyNames[difChallenging], CallbackData: makeClbk(clbk, difChallenging)}},
				{{Text: difficultyNames[difFormidable], CallbackData: makeClbk(clbk, difFormidable)}},
				{{Text: difficultyNames[difLegendary], CallbackData: makeClbk(clbk, difLegendary)}},
				{{Text: difficultyNames[difHeroic], CallbackData: makeClbk(clbk, difHeroic)}},
				{{Text: difficultyNames[difGodly], CallbackData: makeClbk(clbk, difGodly)}},
				{{Text: difficultyNames[difImpossible], CallbackData: makeClbk(clbk, difImpossible)}},
			},
		},
	}
	return emsg
}

func getListCheckMessage(cmd string, chatId int64, list []check) api.SendMessage {
	var btnList [][]api.InlineKeyboardButton
	var btnRow []api.InlineKeyboardButton
	var listText string
	var newestId, oldestId int64
	if len(list) > 0 {
		newestId = list[0].Id
		oldestId = list[len(list)-1].Id
		btnRow = []api.InlineKeyboardButton{
			{Text: "⬅️ newer", CallbackData: makeClbk(cmd, listCheckPrevious, newestId)},
			{Text: "older ➡️", CallbackData: makeClbk(cmd, listCheckNext, oldestId)},
		}
		btnList = append(btnList, btnRow)
		btnRow = nil
	} else {
		listText = "You have no checks at the moment"
	}
	for i, chk := range list {
		listText = concat(
			listText, strconv.Itoa(i+1), ". ", typeNames[chk.Typ],
			"\n", skillNames[chk.Skill], "/",
			difficultyNames[chk.Difficulty], "\n",
			chk.Description, "\n\n")
		btnRow = append(btnRow, api.InlineKeyboardButton{
			Text:         strconv.Itoa(i + 1),
			CallbackData: makeClbk(cmd, listCheckDetail, chk.Id),
		})
		if (i+1)%3 == 0 || i+1 == len(list) {
			btnList = append(btnList, btnRow)
			btnRow = nil
		}
	}
	smsg := api.SendMessage{
		ChatID: chatId,
		Text:   listText,
		ReplyMarkup: &api.InlineKeyboardMarkup{
			InlineKeyboard: btnList,
		},
	}
	return smsg
}

func getSingleCheckEditMessage(chatId int64, msgId int, clbk string, chk check) api.EditMessageText {

	emsg := api.EditMessageText{}
	return emsg
}

func getSingleCheckMessage(chatId int64, chk check) api.SendMessage {
	smsg := api.SendMessage{
		ChatID: chatId,
		Text: concat(typeNames[chk.Typ], ":\n", skillNames[chk.Skill], "/", difficultyNames[chk.Difficulty],
			"\n", chk.Description, "\n\nCreated at: ", chk.CreatedAt.Format(time.DateTime)),
	}
	return smsg
}

func getSkillTxtEditMessage(chatId int64, msgId int, chk check) api.EditMessageText {
	emsg := api.EditMessageText{
		ChatID:    chatId,
		MessageID: msgId,
		Text:      concat("enter descrption of the check:\n", skillNames[chk.Skill], "/", difficultyNames[chk.Difficulty]),
	}
	return emsg
}

func getErrorMessage(chatId int64, err error) api.SendMessage {
	smsg := api.SendMessage{
		ChatID: chatId,
		Text:   concat("update was not handled due to:\n", err.Error()),
	}
	return smsg
}

func getStartMessage(chatId int64) api.SendMessage {
	smsg := api.SendMessage{
		ChatID: chatId,
		Text: `Welcome!\nYou are able to create new /white, retriable checks, and /red, non-retriable checks.\n
			   Use /top command in order to discover your checks and make an attempt to pass these checks:\n`,
	}
	return smsg
}

func getCbqAnswer(cbqId string, text string) api.AnswerCallbackQuery {
	answer := api.AnswerCallbackQuery{
		CallbackQueryId: cbqId,
		Text:            text,
	}
	return answer
}

func getErrorCbqAnswer(cbqId string, err error) api.AnswerCallbackQuery {
	answer := api.AnswerCallbackQuery{
		CallbackQueryId: cbqId,
		Text:            err.Error(),
		ShowAlert:       true,
	}
	return answer
}

func makeClbk(start string, params ...int64) string {
	var sb strings.Builder
	for i := 0; i < len(params); i++ {
		if i > 0 || start == "" {
			sb.WriteString("/")
		} else {
			sb.WriteString(start)
			sb.WriteString("/")
		}
		sb.WriteString(strconv.FormatInt(params[i], 10))
	}
	return sb.String()
}

func concat(str ...string) string {
	var sb strings.Builder
	for _, s := range str {
		sb.WriteString(s)
	}
	return sb.String()
}
