package main

import (
	"fmt"
	"github.com/spf13/viper"
)

func main() {

	viper.SetConfigName("ggg")
	viper.SetConfigType("toml")
	viper.AddConfigPath("$HOME/.config/")

	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println("Error reading config file:", err)
	} else {
		fmt.Println("Config file loaded successfully")
	}
}
