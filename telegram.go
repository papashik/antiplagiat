package main

import (
	databaseapi "antiplagiat/db"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"os"
	"strconv"
	"strings"
)

var (
	TG_BOT_API_KEY = os.Getenv("TG_BOT_API_KEY")
)

var statusMap map[int64]int = make(map[int64]int)
var textFromUsers map[int64]string = make(map[int64]string)
var headers map[int64][]string = make(map[int64][]string)
var langKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("Go", "go"),
		tgbotapi.NewInlineKeyboardButtonData("C++", "cpp"),
		tgbotapi.NewInlineKeyboardButtonData("Java", "java"),
		tgbotapi.NewInlineKeyboardButtonData("Python", "python"),
	),
)

type courseStruct struct {
	name      string
	long_name string
	amount    int
}

func getNamesOfCourses() []tgbotapi.InlineKeyboardButton {
	rows, err := databaseapi.GetCourseRows()
	if err != nil {
		log.Println(err)
	}
	courses := make([]tgbotapi.InlineKeyboardButton, 0)
	for rows.Next() {
		var course courseStruct
		if err := rows.Scan(&course.name, &course.long_name, &course.amount); err != nil {
			log.Println(err)
		}
		button := tgbotapi.NewInlineKeyboardButtonData(course.long_name, course.name)
		courses = append(courses, button)
	}
	return courses
}

func searchForErrors(Message *tgbotapi.Message, bot *tgbotapi.BotAPI) {
	msg := tgbotapi.NewMessage(Message.Chat.ID, "")
	statusMap[Message.From.ID] = 0
	head, ok := headers[Message.From.ID]

	if ok { // заголовок есть
		text, good := textFromUsers[Message.From.ID]
		if good { //текст лабораторной присутствует
			lab, _ := strconv.Atoi(head[2])
			variant, _ := strconv.Atoi(head[3])
			err := databaseapi.AddDelivery(head[0], head[1], head[5], lab, variant, text, head[4])
			if err != nil {
				log.Println(err)
				msg.Text = "Произошла ошибка при загрузке лабораторной работы в базу данных. \n/help - возможные причины проблем при отправке\n/send - повторить отправку\n\nОшибка: \n" + err.Error()
			} else {
				msg.Text = "Загрузка лабораторной работы произошла успешно"
				fmt.Println(text)
			}
		} else {
			msg.Text = "Отсутствует текст лабораторной работы. Пожалуйста, повторите загрузку с помощью команды /send."
		}
	} else {
		msg.Text = "Отсутствует информация о лабораторной работе. Пожалуйста, повторите загрузку с помощью команды /send."
	}
	if _, err := bot.Send(msg); err != nil {
		log.Println(err)
	}
	delete(textFromUsers, Message.From.ID)
	delete(headers, Message.From.ID)
}
func main() {
	bot, err := tgbotapi.NewBotAPI(TG_BOT_API_KEY)
	if err != nil {
		log.Println(err)
	}

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)
	var msg tgbotapi.MessageConfig
	for update := range updates {
		if update.Message == nil && update.CallbackQuery != nil { // обработка сообщений от inline клавиатуры
			status, ok := statusMap[update.CallbackQuery.From.ID]
			msg = tgbotapi.NewMessage(update.CallbackQuery.From.ID, "")
			if ok {
				switch status {
				case 2:
					headers[update.CallbackQuery.From.ID] = append(headers[update.CallbackQuery.From.ID], update.CallbackQuery.Data) //записываем язык в заголовок
					statusMap[update.CallbackQuery.From.ID] = 3
					courseKeyboard := tgbotapi.NewInlineKeyboardMarkup(getNamesOfCourses())
					//обновляем информацию о курсах и создаем клавиатуру с доступными
					msg.ReplyMarkup = courseKeyboard
					msg.Text = "Выберите курс"
				case 3:
					headers[update.CallbackQuery.From.ID] = append(headers[update.CallbackQuery.From.ID], update.CallbackQuery.Data) //записываем название курса
					statusMap[update.CallbackQuery.From.ID] = 4
					msg.Text = "Начните отправку кода программ с помощью сообщений. Используйте команду /stop после последнего сообщения с кодом"
				}
			}

		} else if !update.Message.IsCommand() { // сообщение без команды - либо код, либо заголовок, либо не несущее информации сообщение
			msg = tgbotapi.NewMessage(update.Message.Chat.ID, "")
			status, ok := statusMap[update.Message.From.ID]
			if ok {
				switch status {
				case 1:
					info := strings.Split(update.Message.Text, "_")

					if len(info) != 4 { // не хватает элементов в заголовке
						msg.Text = "Заголовок лабораторной работы введен некорректно, введите еще раз"
					} else {
						_, err1 := strconv.Atoi(info[2])
						_, err2 := strconv.Atoi(info[3])
						if err1 != nil {
							msg.Text = "Номер лабораторной работы введен некорректно, отправьте заголовок еще раз"
						} else if err2 != nil {
							msg.Text = "Номер варианта введен некорректно, отправьте заголовок еще раз"
						} else {
							headers[update.Message.From.ID] = info
							statusMap[update.Message.From.ID] = 2
							msg.ReplyMarkup = langKeyboard
							msg.Text = "Выберите язык"
						}
					}
				case 4:
					msg.Text = "Получено сообщение с кодом. Используйте команду /stop для остановки отправки после последнего сообщения"
					textFromUsers[update.Message.From.ID] = textFromUsers[update.Message.From.ID] + "\n" + update.Message.Text
				case 0:
					msg.Text = "Сообщение без команды"
					fmt.Printf("%s", update.Message.Text)
				}
			} else {
				msg.Text = "Сообщение без команды"

			}
		} else {
			msg = tgbotapi.NewMessage(update.Message.Chat.ID, "")
			switch update.Message.Command() {
			case "help":
				msg.Text = "Почему лабораторная работа может не загружаться: \n1. Текст программы содержит синтаксические ошибки \n2. В тексте программ встретились комбинации Telegram по форматированию текста (например, часть текста скрыта за спойлером). В таком случае отправьте код заново в моноширинном формате (форматирование доступно по нажатию правой кнопки мыши)"
			case "start":
				msg.Text = "Бот для приема лабораторных работ.\nДля отправки лабораторной работы используйте команду /send"
			case "send":
				msg.Text = "Введите информацию о лабораторной работе в формате Фамилия_Имя_НомерЛабораторнойРаботы_Вариант, например Иванов_Иван_3_4. Важно заполнять поля в нужном порядке!"
				statusMap[update.Message.From.ID] = 1
				// статус 1 - ожидание заголовка, 2 - ожидание выбора языка программирования? 3 - ожидание выбора курса, 4 - ожидание сообщения с кодом, 0 - нейтральный
			case "stop":
				go searchForErrors(update.Message, bot)
				msg.Text = "Происходит загрузка лабораторной работы, ожидайте"

			default:
				msg.Text = "Неизвестная команда"

			}
		}
		if _, err := bot.Send(msg); err != nil {
			log.Println(err)
		}

	}
}
