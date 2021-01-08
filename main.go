package main

import (
	"fmt"

	"github.com/poodlenoodle42/Discord_Calender_Bot/config"
)

func main() {
	config := config.ReadConfigFile("config.yaml")
	fmt.Println(config.Token)
}
