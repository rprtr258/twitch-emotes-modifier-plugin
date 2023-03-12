package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/rprtr258/twitch-emotes-modifier-plugin/internal/logic"
)

func run() error {
	if len(os.Args) != 2 {
		return fmt.Errorf("usage: %s <query>", filepath.Base(os.Args[0]))
	}

	id, err := logic.ProcessQuery("01FGH31ADR00047CPSV86K53E0", logic.ParseTokens(logic.TokenRE.FindAllString(os.Args[1], -1)))
	if err != nil {
		return err
	}

	fmt.Println(id)
	return nil
}

func main() {
	log.SetFlags(0)
	if err := run(); err != nil {
		log.Fatal(err.Error())
	}
}
