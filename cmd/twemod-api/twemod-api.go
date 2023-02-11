package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/rprtr258/twitch-emotes-modifier-plugin/internal/logic"
	"go.uber.org/zap"
)

type request struct {
	Channel string `json:"channel"`
	Message string `json:"message"`
}

func postHandler(c echo.Context) error {
	var req request
	if err := c.Bind(&req); err != nil {
		return err
	}

	prompts := [][]string{}
	for _, word := range strings.Fields(req.Message) {
		indexes := logic.TokenRE.FindAllIndex([]byte(word), -1)
		if len(indexes) < 2 ||
			indexes[0][0] > 0 ||
			indexes[len(indexes)-1][1] < len(word)-1 {
			continue
		}

		tokens := make([]string, len(indexes))
		for i, idx := range indexes {
			tokens[i] = word[idx[0]:idx[1]]
		}

		prompts = append(prompts, tokens)
	}
	if len(prompts) == 0 {
		return nil
	}

	for _, prompt := range prompts {
		resultID, err := logic.ProcessQuery(req.Channel, logic.ParseTokens(prompt))
		if err != nil {
			return c.String(
				http.StatusInternalServerError,
				err.Error(),
			)
		}
		zap.L().Info(
			"processed prompt",
			zap.Strings("prompt", prompt),
			zap.String("resultID", string(resultID)),
		)
	}

	return nil
}

func run() error {
	l, err := zap.NewProduction()
	if err != nil {
		return err
	}
	zap.ReplaceGlobals(l)
	defer l.Sync()

	e := echo.New()
	assetHandler := http.FileServer(http.FS(os.DirFS(".")))
	e.GET("/static/*",
		echo.WrapHandler(http.StripPrefix("/static/", assetHandler)),
	)
	e.POST("/", postHandler)

	const addr = ":1323"
	return e.Start(addr)
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err.Error())
	}
}
