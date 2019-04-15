package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
)

type Timer struct {
	Name             string `toml:"name"`
	Interval         string `toml:"interval"`
	Pause            string `toml:"pause"`
	Title            string `toml:"title"`
	Text             string `toml:"text"`
	IntervalDuration time.Duration
	PauseDuration    time.Duration
}

type Config struct {
	Timers []Timer `toml:"intervals"`
}

var wg sync.WaitGroup

func main() {
	// If there are errors at the end outputs them
	// Если есть ошибки в конце выводит их
	var err error
	defer func() {
		if err != nil {
			fmt.Printf("We have some erros here: %e\n", err)
		}
	}()

	// Get current user
	usr, err := user.Current()
	if err != nil {
		return
	}

	// Reads the configuration file by path from the command-line argument, either "./config.yml" by default
	// Читает конфигурационный файл по пути из аргумента командной строки, либо "./config.toml" по умолчанию
	var config Config
	if len(os.Args) > 1 {
		config, err = readConfig(os.Args[1])
	} else {
		config, err = readConfig(usr.HomeDir + "/.config/chronoshift/config.toml")
	}
	if err != nil {
		return
	}

	// Setting up and running each timer
	// Настройка и запуск каждого таймера
	for index, timer := range config.Timers {
		// Gets the duration of the interval
		// Получает длительность интервала
		interval, err := stringToDuration(timer.Interval)
		if err != nil {
			return
		}
		config.Timers[index].IntervalDuration = interval

		// Gets the duration of the pause
		// Получает длительность паузы
		pause, err := stringToDuration(timer.Pause)
		if err != nil {
			return
		}
		config.Timers[index].PauseDuration = pause

		// Each timer runs in a separate goroutine
		// Запускает каждый таймер в отдельной горутине
		wg.Add(1)
		go runTimer(config.Timers[index])
	}

	wg.Wait()

}

// Starts the timer with the received settings
// Запусскает таймер с полученными настройками
func runTimer(timer Timer) {
	defer wg.Done()
	for {
		time.Sleep(timer.IntervalDuration)
		cmd := exec.Command("notify-send",
			"-u", "normal",
			"-t", "20000",
			// "-h", "int:x:500",
			// "-h", "int:y:500",
			// "-i", user.HomeDir+"/Sync/git/GorusEye/eyeofhorus64.png",
			"-a", "ChronoShift",
			timer.Title,
			timer.Text)
		err := cmd.Run()
		if err != nil {
			fmt.Printf("cmd.Run() failed with %s\n", err)
		}
		time.Sleep(timer.PauseDuration)
	}
}

// Reads the configuration file .tom l, returns a structure of type Config
// Читает конфигурационный файл .toml, возвращает структуру типа Config
func readConfig(path string) (Config, error) {
	var config Config
	if _, err := toml.DecodeFile(path, &config); err != nil {
		return Config{}, err
	}
	return config, nil
}

// Converts a text value ("20m") to a time type.Duration
// Преобразует текстовое значение ("20m") в тип time.Duration
func stringToDuration(stringDuration string) (time.Duration, error) {
	if stringDuration == "" {
		stringDuration = "0s"
	}
	multiplier, err := strconv.Atoi(stringDuration[:len(stringDuration)-1])
	if err != nil {
		return 0, err
	}
	var timeUnit time.Duration
	switch stringDuration[len(stringDuration)-1:] {
	case "s":
		timeUnit = time.Second
	case "m":
		timeUnit = time.Minute
	case "h":
		timeUnit = time.Hour
	default:
		return 0, errors.New(`Wrong time format, use something like that: "20s", "5m" or "2h"`)
	}
	return time.Duration(multiplier) * timeUnit, err
}
