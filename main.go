package main

import (
	"encoding/json"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"strings"
)

type WeatherApiConfig struct {
	WeatherApiKey string
}

type TelegramApiConfig struct {
	TelegramApiKey string
}

type Response struct {
	Weather []map[string]interface{}
	Main    map[string]interface{}
	Wind    map[string]interface{}
	Name    string
}

func getWindCharacteristic(wind_speed int) string {
	if wind_speed <= 5 {
		return "слабый"
	} else if wind_speed <= 14 {
		return "умеренный"
	} else if wind_speed <= 24 {
		return "сильный"
	} else if wind_speed <= 32 {
		return "очень сильный"
	}
	return "ураганный"
}

func degreesToString(degree float64) string {
	if degree > 337.5 {
		return "северный"
	}
	if degree > 292.5 {
		return "северо-западный"
	}
	if degree > 247.5 {
		return "западный"
	}
	if degree > 202.5 {
		return "юго-западный"
	}
	if degree > 157.5 {
		return "южный"
	}
	if degree > 122.5 {
		return "юго-восточный"
	}
	if degree > 67.5 {
		return "восточный"
	}
	if degree > 22.5 {
		return "северо-воточный"
	}
	return "северный"
}

func createWeatherMessage(weatherConfig *WeatherApiConfig, channel chan string) {
	res, err := http.Get(fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?q=moscow&appid=%s&lang=ru&units=metric&mode=json", weatherConfig.WeatherApiKey))
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)

	//fmt.Printf("%v", string(body))
	if err != nil {
		fmt.Printf("Error occured: %s", err.Error())
		channel <- ""
	}

	var response Response

	err = json.Unmarshal(body, &response)
	if err != nil {
		fmt.Printf("\nCan't unmarshal file: %s", err.Error())
		channel <- ""
	}

	temp_min := response.Main["temp_min"]
	temp_max := response.Main["temp_max"]
	pressure := response.Main["pressure"]
	humidity := response.Main["humidity"]
	wind_speed := response.Wind["speed"]
	wind_deg := response.Wind["deg"]

	weather_msg := fmt.Sprintf("Погода: %s\n"+
		"Температура: %v-%v°C\n"+
		"Давление: %v мм рт. ст., влажность %v%%\n"+
		"Ветер: %s, %v м/с, %s", response.Weather[0]["description"],
		math.Round(temp_min.(float64)),
		math.Round(temp_max.(float64)),
		math.Round(pressure.(float64)*0.750062),
		humidity,
		getWindCharacteristic(int(wind_speed.(float64))),
		int(wind_speed.(float64)),
		degreesToString(wind_deg.(float64)))

	channel <- weather_msg

}

func sendMessage(weatherConfig *WeatherApiConfig, update *tgbotapi.Update, bot *tgbotapi.BotAPI) {
	channel := make(chan string)
	go createWeatherMessage(weatherConfig, channel)

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, <-channel)
	//msg.ReplyToMessageID = update.Message.MessageID

	bot.Send(msg)
}

func main() {

	var weatherConfig WeatherApiConfig
	var telegramConfig TelegramApiConfig

	dat, err := os.ReadFile("api_config.json")
	if err != nil {
		fmt.Printf("Can't read file: %s\nYou need to fill the file \"api_config.json\" with the variable \"weatherApiKey\" and put as its value your api key from openweathermap.org", err.Error())
		return
	}
	strJson := string(dat)

	err = json.Unmarshal([]byte(strJson), &weatherConfig)
	if err != nil {
		fmt.Printf("Can't unmarshal file: %s", err.Error())
		return
	}

	err = json.Unmarshal([]byte(strJson), &telegramConfig)
	if err != nil {
		fmt.Printf("Can't unmarshal file: %s", err.Error())
		return
	}

	bot, err := tgbotapi.NewBotAPI(telegramConfig.TelegramApiKey)
	if err != nil {
		fmt.Printf("Error occured: %s", err.Error())
		return
	}

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			if strings.ToLower(update.Message.Text) == "погода" {
				//fmt.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
				log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
				go sendMessage(&weatherConfig, &update, bot)
			}
		}
	}
}
