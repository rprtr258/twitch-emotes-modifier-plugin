package main

import (
	"log"
)

func run() error {
	// id, err := logic.ProcessQuery(os.Args[1])
	// if err != nil {
	// 	return err
	// }

	// fmt.Println(id)
	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err.Error())
	}
}
