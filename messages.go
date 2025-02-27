package main

import (
	"discocheckbot/api"
	"strconv"
	"strings"
	"unicode/utf16"
)

func getSkillMessage(cmd string, chatId int64, color int) api.SendMessage {
	chkColor := int64(color)
	smsg := api.SendMessage{
		ChatID: chatId,
		Text:   "Select skill:",
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
		Text:      "Select check difficulty:",
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
	var markup *api.InlineKeyboardMarkup
	var format []api.MessageEntity
	var nextId int64
	var msgText myStringsBuilder
	if len(list) > 0 {
		if len(list) < maxChecksAtListPage {
			nextId = 0
		} else {
			nextId = list[len(list)-1].Id
		}
		btnRow = []api.InlineKeyboardButton{
			{Text: "⬅️ Newer", CallbackData: makeClbk(cmd, listCheckBackward, 0)},
			{Text: "Older ➡️", CallbackData: makeClbk(cmd, listCheckForward, nextId)},
		}
		btnList = append(btnList, btnRow)
		btnRow = nil
	} else {
		msgText.sb.WriteString("You have no checks at the moment")
	}
	for i, chk := range list {
		crossBegin := 0
		crossEnd := 0
		boldBegin := 0
		boldEnd := 0
		res := 0
		sep := " "
		if chk.closed() {
			sep = " - "
			res = chk.Attempts[len(chk.Attempts)-1].Result
			crossBegin = len(utf16.Encode([]rune(msgText.sb.String())))
			if crossBegin > 0 {
				crossBegin--
			}
		}
		msgText.concat(strconv.Itoa(i+1), ". ", typeNames[chk.Typ], sep, resultNames[res], "\n")
		boldBegin = len(utf16.Encode([]rune(msgText.sb.String()))) - 1
		msgText.concat(skillNames[chk.Skill], " - ", difficultyNames[chk.Difficulty], "\n")
		boldEnd = len(utf16.Encode([]rune(msgText.sb.String()))) - 1
		msgText.sb.WriteString(chk.Description)
		if chk.closed() {
			crossEnd = len(utf16.Encode([]rune(msgText.sb.String()))) - 1
			format = append(format, api.MessageEntity{
				Type:   api.CrossedEntity,
				Offset: crossBegin,
				Length: crossEnd - crossBegin})
		}
		format = append(format, api.MessageEntity{
			Type:   api.BoldEntity,
			Offset: boldBegin,
			Length: boldEnd - boldBegin,
		})
		msgText.sb.WriteString("\n\n")
		btnRow = append(btnRow, api.InlineKeyboardButton{
			Text:         strconv.Itoa(i + 1),
			CallbackData: makeClbk(cmd, listCheckDetail, chk.Id),
		})
		if (i+1)%maxCheckBtnInRow == 0 || i+1 == len(list) {
			btnList = append(btnList, btnRow)
			btnRow = nil
		}
	}
	if len(btnList) > 0 {
		markup = &api.InlineKeyboardMarkup{InlineKeyboard: btnList}
	}
	smsg := api.SendMessage{
		ChatID:      chatId,
		Text:        msgText.sb.String(),
		ReplyMarkup: markup,
		Entities:    format,
	}
	return smsg
}

func getListCheckEditMessage(chatId int64, msgId int, list []check) api.EditMessageText {
	baseMsg := getListCheckMessage(seeTop, chatId, list)
	var prevId, nextId int64
	prevId = list[0].Id
	nextId = list[len(list)-1].Id
	baseMsg.ReplyMarkup.InlineKeyboard[0][0].CallbackData = makeClbk(seeTop, listCheckBackward, prevId)
	baseMsg.ReplyMarkup.InlineKeyboard[0][1].CallbackData = makeClbk(seeTop, listCheckForward, nextId)

	emsg := api.EditMessageText{
		ChatID:      chatId,
		MessageID:   msgId,
		Text:        baseMsg.Text,
		ReplyMarkup: baseMsg.ReplyMarkup,
		Entities:    baseMsg.Entities,
	}
	return emsg
}

func getSingleCheckEditMessage(chatId int64, msgId int, chk check) api.EditMessageText {
	var msgText myStringsBuilder
	msgText.concat(typeNames[chk.Typ], ":\n")
	boldBegin := len(utf16.Encode([]rune(msgText.sb.String()))) - 1
	msgText.concat(skillNames[chk.Skill], " - ", difficultyNames[chk.Difficulty], "\n")
	boldEnd := len(utf16.Encode([]rune(msgText.sb.String()))) - 1
	msgText.concat(chk.Description, "\n\n")

	emsg := api.EditMessageText{
		ChatID:    chatId,
		MessageID: msgId,
		Entities:  []api.MessageEntity{{Type: api.BoldEntity, Offset: boldBegin, Length: boldEnd - boldBegin}},
	}
	for _, attempt := range chk.Attempts {
		msgText.concat("Attempt at: ", attempt.CreatedAt.Format("2.01.2006 15:04"), "\nResult: ",
			resultNames[attempt.Result], "\n")
	}
	emsg.Text = msgText.sb.String()
	if !chk.closed() {
		emsg.ReplyMarkup = &api.InlineKeyboardMarkup{
			InlineKeyboard: [][]api.InlineKeyboardButton{
				{{Text: resultNames[resSuccess], CallbackData: makeClbk(seeTop, listCheckAction, chk.Id, resSuccess)}},
				{{Text: resultNames[resFailure], CallbackData: makeClbk(seeTop, listCheckAction, chk.Id, resFailure)}},
				{{Text: resultNames[resCanceled], CallbackData: makeClbk(seeTop, listCheckAction, chk.Id, resCanceled)}},
				{{Text: "Back", CallbackData: makeClbk(seeTop, listCheckForward, 0)}},
			},
		}
	} else {
		emsg.ReplyMarkup = &api.InlineKeyboardMarkup{
			InlineKeyboard: [][]api.InlineKeyboardButton{
				{{Text: "Back", CallbackData: makeClbk(seeTop, listCheckForward, 0)}},
			},
		}
	}
	return emsg
}

func getSingleCheckMessage(chatId int64, chk check) api.SendMessage {
	var msgText myStringsBuilder
	msgText.concat(typeNames[chk.Typ], ":\n")
	boldBegin := len(utf16.Encode([]rune(msgText.sb.String()))) - 1
	msgText.concat(skillNames[chk.Skill], " - ", difficultyNames[chk.Difficulty], "\n")
	boldEnd := len(utf16.Encode([]rune(msgText.sb.String()))) - 1
	msgText.concat(chk.Description, "\n\nCreated at: ", chk.CreatedAt.Format("2.01.2006 15:04"))
	smsg := api.SendMessage{
		ChatID:   chatId,
		Text:     msgText.sb.String(),
		Entities: []api.MessageEntity{{Type: api.BoldEntity, Offset: boldBegin, Length: boldEnd - boldBegin}},
	}
	return smsg
}

func getSkillTxtEditMessage(chatId int64, msgId int, chk check) api.EditMessageText {
	var msgText myStringsBuilder
	msgText.concat("Enter descrption of the check:\n")
	boldBegin := len(utf16.Encode([]rune(msgText.sb.String()))) - 1
	msgText.concat(skillNames[chk.Skill], " - ", difficultyNames[chk.Difficulty], "\n")
	boldEnd := len(utf16.Encode([]rune(msgText.sb.String()))) - 1
	emsg := api.EditMessageText{
		ChatID:    chatId,
		MessageID: msgId,
		Text:      msgText.sb.String(),
		Entities:  []api.MessageEntity{{Type: api.BoldEntity, Offset: boldBegin, Length: boldEnd - boldBegin}},
	}
	return emsg
}

func getErrorMessage(chatId int64, err error) api.SendMessage {
	var msgText myStringsBuilder
	msgText.concat("Request was not handled due to error:\n", err.Error())
	smsg := api.SendMessage{
		ChatID: chatId,
		Text:   msgText.sb.String(),
	}
	return smsg
}

func getStartMessage(chatId int64) api.SendMessage {
	smsg := api.SendMessage{
		ChatID: chatId,
		Text: `Welcome!
				You are able to create new /white, retriable checks, and /red, non-retriable checks.
			    Use /top command in order to discover your checks and make an attempt to pass them`,
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
	var msgText myStringsBuilder
	msgText.concat("Request was not handled due to error:\n", err.Error())
	answer := api.AnswerCallbackQuery{
		CallbackQueryId: cbqId,
		Text:            msgText.sb.String(),
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

// just the same as strings.Builder, but with multiple WriteString arguments
type myStringsBuilder struct {
	sb strings.Builder
}

func (this *myStringsBuilder) concat(str ...string) int {
	var totalLen int
	for _, s := range str {
		slen, _ := this.sb.WriteString(s)
		totalLen += slen
	}
	return totalLen
}
