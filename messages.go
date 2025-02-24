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
	var nextId int64
	if len(list) > 0 {
		if len(list) < maxChecksAtListPage {
			nextId = 0
		} else {
			nextId = list[len(list)-1].Id
		}
		btnRow = []api.InlineKeyboardButton{
			{Text: "⬅️ newer", CallbackData: makeClbk(cmd, listCheckNext, 0)},
			{Text: "older ➡️", CallbackData: makeClbk(cmd, listCheckNext, nextId, 0)},
		}
		btnList = append(btnList, btnRow)
		btnRow = nil
	} else {
		listText = "You have no checks at the moment"
	}
	for i, chk := range list {
		res := 0
		if chk.closed() {
			res = chk.Attempts[len(chk.Attempts)-1].Result
		}
		listText = concat(
			listText, strconv.Itoa(i+1), ". ", typeNames[chk.Typ], " ", resultNames[res],
			"\n", skillNames[chk.Skill], "/",
			difficultyNames[chk.Difficulty],
			"\n", chk.Description, "\n\n")
		btnRow = append(btnRow, api.InlineKeyboardButton{
			Text:         strconv.Itoa(i + 1),
			CallbackData: makeClbk(cmd, listCheckDetail, chk.Id),
		})
		if (i+1)%3 == 0 || i+1 == len(list) { //todo constant max buttons
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

func getListCheckEditMessage(chatId int64, msgId int, list []check, clbkPar []string) api.EditMessageText {
	baseMsg := getListCheckMessage(seeTop, chatId, list)
	var reqId, prevId, nextId int64
	prevId, _ = strconv.ParseInt(clbkPar[len(clbkPar)-1], 10, 64)
	reqId, _ = strconv.ParseInt(clbkPar[2], 10, 64)

	if len(list) < maxChecksAtListPage {
		nextId = reqId
		reqId = prevId
	} else {
		nextId = list[len(list)-1].Id
	}
	baseMsg.ReplyMarkup.InlineKeyboard[0][0].CallbackData = makeClbk(seeTop, listCheckNext, prevId)
	baseMsg.ReplyMarkup.InlineKeyboard[0][1].CallbackData = makeClbk(seeTop, listCheckNext, nextId, reqId)

	emsg := api.EditMessageText{
		ChatID:      chatId,
		MessageID:   msgId,
		Text:        baseMsg.Text,
		ReplyMarkup: baseMsg.ReplyMarkup,
	}
	return emsg
}

func getSingleCheckEditMessage(chatId int64, msgId int, chk check) api.EditMessageText {
	emsg := api.EditMessageText{
		ChatID:    chatId,
		MessageID: msgId,
		Text: concat(typeNames[chk.Typ], ":\n", skillNames[chk.Skill], "/", difficultyNames[chk.Difficulty],
			"\n", chk.Description, "\n\nCreated at: ", chk.CreatedAt.Format(time.DateTime), "\n\n"),
	}
	for _, attempt := range chk.Attempts {
		emsg.Text = concat(emsg.Text, "Attempt at ", attempt.CreatedAt.Format(time.DateTime), "\nResult: ",
			resultNames[attempt.Result], "\n")
	}
	if !chk.closed() {
		emsg.ReplyMarkup = &api.InlineKeyboardMarkup{
			InlineKeyboard: [][]api.InlineKeyboardButton{
				{{Text: resultNames[resSuccess], CallbackData: makeClbk(seeTop, listCheckAction, chk.Id, resSuccess)}},
				{{Text: resultNames[resFailure], CallbackData: makeClbk(seeTop, listCheckAction, chk.Id, resFailure)}},
				{{Text: resultNames[resCanceled], CallbackData: makeClbk(seeTop, listCheckAction, chk.Id, resCanceled)}},
				{{Text: "back", CallbackData: makeClbk(seeTop, listCheckNext, 0)}},
			},
		}
	} else {
		emsg.ReplyMarkup = &api.InlineKeyboardMarkup{
			InlineKeyboard: [][]api.InlineKeyboardButton{
				{{Text: "back", CallbackData: makeClbk(seeTop, listCheckNext, 0)}},
			},
		}
	}
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

//type myStringsBuilder strings.Builder

// func (this *myStringsBuilder) concat(str ...string) {
// 	for _, s := range str {
// 		this.WriteString(s)
// 	}
// }

// func (this *myStringsBuilder) string() string {
// 	return this.String()
// }

func concat(str ...string) string {
	var sb strings.Builder
	for _, s := range str {
		sb.WriteString(s)
	}
	return sb.String()
}
