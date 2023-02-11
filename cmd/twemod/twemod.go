package main

import (
	"fmt"
	"log"
	"os"

	"github.com/rprtr258/twitch-emotes-modifier-plugin/internal/logic"
)

func run() error {
	id, err := logic.ProcessQuery("61506ec343b2d9da0d328dc0", logic.ParseTokens(logic.TokenRE.FindAllString(os.Args[1], -1)))
	if err != nil {
		return err
	}

	fmt.Println(id)
	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err.Error())
	}
}
