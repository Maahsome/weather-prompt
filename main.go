package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type SingleCityWeather struct {
	Coord struct {
		Lon float64 `json:"lon"`
		Lat float64 `json:"lat"`
	} `json:"coord"`
	Weather []struct {
		ID          int    `json:"id"`
		Main        string `json:"main"`
		Description string `json:"description"`
		Icon        string `json:"icon"`
	} `json:"weather"`
	Base string `json:"base"`
	Main struct {
		Temp      float64 `json:"temp"`
		FeelsLike float64 `json:"feels_like"`
		TempMin   float64 `json:"temp_min"`
		TempMax   float64 `json:"temp_max"`
		Pressure  int     `json:"pressure"`
		Humidity  int     `json:"humidity"`
	} `json:"main"`
	Visibility int `json:"visibility"`
	Wind       struct {
		Speed float64 `json:"speed"`
		Deg   int     `json:"deg"`
	} `json:"wind"`
	Clouds struct {
		All int `json:"all"`
	} `json:"clouds"`
	Dt  int `json:"dt"`
	Sys struct {
		Type    int    `json:"type"`
		ID      int    `json:"id"`
		Country string `json:"country"`
		Sunrise int    `json:"sunrise"`
		Sunset  int    `json:"sunset"`
	} `json:"sys"`
	Timezone int    `json:"timezone"`
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Cod      int    `json:"cod"`
}

func main() {
	var cw SingleCityWeather

	initConfig()

	apiKey := viper.GetString("apikey")
	home, err := os.UserHomeDir()
	cobra.CheckErr(err)
	cacheFile := fmt.Sprintf("%s/.config/weather-prompt/cache.json", home)
	cacheIndicator := "*"

	cacheInfo, err := os.Stat(cacheFile)
	if err == nil {
		now := time.Now()
		diff := now.Sub(cacheInfo.ModTime())
		if diff > (time.Duration(600) * time.Second) {
			fetchUri := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?q=Moose Lake,MN,US&appid=%s&units=imperial", apiKey)
			restClient := resty.New()
			resp, resperr := restClient.R().
				Get(fetchUri)

			if resperr != nil {
				logrus.WithError(resperr).Fatal("Oops")
			}

			marshErr := json.Unmarshal(resp.Body(), &cw)
			if marshErr != nil {
				logrus.Fatal("Cannot marshall Weather", marshErr)
			}

			wferr := os.WriteFile(cacheFile, resp.Body(), 0644)
			cobra.CheckErr(wferr)
			cacheIndicator = ""
		} else {
			cacheRead, rerr := os.ReadFile(cacheFile)
			cobra.CheckErr(rerr)

			marshErr := json.Unmarshal(cacheRead, &cw)
			if marshErr != nil {
				logrus.Fatal("Cannot marshall Weather", marshErr)
			}
		}
	} else {
		fetchUri := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?q=Moose Lake,MN,US&appid=%s&units=imperial", apiKey)
		restClient := resty.New()
		resp, resperr := restClient.R().
			Get(fetchUri)

		if resperr != nil {
			logrus.WithError(resperr).Fatal("Oops")
		}

		marshErr := json.Unmarshal(resp.Body(), &cw)
		if marshErr != nil {
			logrus.Fatal("Cannot marshall Weather", marshErr)
		}

		wferr := os.WriteFile(cacheFile, resp.Body(), 0644)
		cobra.CheckErr(wferr)
		cacheIndicator = ""
	}

	indicator := "☀️"
	switch cw.Weather[0].Main {
	case "Thunderstorm":
		indicator = "⛈"
	case "Drizzle":
		indicator = "🌦"
	case "Rain":
		indicator = "🌧"
	case "Snow":
		indicator = "🌨"
	case "Tornado":
		indicator = "🌪"
	case "Fog":
		indicator = "💨"
	case "Clouds":
		if cw.Weather[0].ID == 801 {
			indicator = "🌤"
		}
		if cw.Weather[0].ID == 802 {
			indicator = "⛅️"
		}
		if cw.Weather[0].ID == 803 {
			indicator = "🌥"
		}
		if cw.Weather[0].ID == 804 {
			indicator = "☁️"
		}
	}
	celsiusMainTemp := convertToCelsius(cw.Main.Temp)
	fmt.Printf("%.2fF/%.2fC (%.2fF)%s %s", cw.Main.Temp, celsiusMainTemp, cw.Main.FeelsLike, cacheIndicator, indicator)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// Find home directory.
	home, err := os.UserHomeDir()
	cobra.CheckErr(err)

	workDir := fmt.Sprintf("%s/.config/weather-prompt", home)
	if _, err := os.Stat(workDir); err != nil {
		if os.IsNotExist(err) {
			mkerr := os.MkdirAll(workDir, os.ModePerm)
			if mkerr != nil {
				logrus.Fatal("Error creating ~/.config/weather-prompt directory", mkerr)
			}
		}
	}
	if stat, err := os.Stat(workDir); err == nil && stat.IsDir() {
		configFile := fmt.Sprintf("%s/%s", workDir, "config.yaml")
		createRestrictedConfigFile(configFile)
		viper.SetConfigFile(configFile)
	} else {
		logrus.Info("The ~/.config/weather-prompt path is a file and not a directory, please remove the 'weather-prompt' file.")
		os.Exit(1)
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		logrus.Warn("Failed to read viper config file.")
	}
}

func createRestrictedConfigFile(fileName string) {
	if _, err := os.Stat(fileName); err != nil {
		if os.IsNotExist(err) {
			file, ferr := os.Create(fileName)
			if ferr != nil {
				logrus.Info("Unable to create the configfile.")
				os.Exit(1)
			}
			mode := int(0600)
			if cherr := file.Chmod(os.FileMode(mode)); cherr != nil {
				logrus.Info("Chmod for config file failed, please set the mode to 0600.")
			}
		}
	}
}

func convertToCelsius(value float64) float64 {
	convertedValue := (value - 32) * 5 / 9
	return convertedValue
}

func convertToFahrenheit(value float64) float64 {
	convertedValue := (value * 9 / 5) + 32
	return convertedValue
}
