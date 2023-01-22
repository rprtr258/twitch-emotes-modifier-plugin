package main

import (
	"net/http"
	"os"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/rprtr258/twitch-emotes-modifier-plugin/internal/logic"
)

const _indexPage = `<!DOCTYPE html><html>
<head><style>
body {
	background: black;
	display: flex;
	align-items: center;
	height: 90vh;
	justify-content: center;
	flex-direction: column;
}
</style></head>
<body>
	<form action="/" method="POST">
		<textarea name="q">%</textarea>
		<button>send</button>
	</form>
	<img src="/static/%.webp" />
</body>
</html>`

func getFileSystem() http.FileSystem {
	return http.FS(os.DirFS("."))
}

func main() {
	e := echo.New()
	assetHandler := http.FileServer(getFileSystem())
	e.GET(
		"/static/*",
		echo.WrapHandler(http.StripPrefix("/static/", assetHandler)),
	)
	e.GET(
		"/",
		func(c echo.Context) error {
			return c.HTML(http.StatusOK, _indexPage)
		},
	)
	e.POST(
		"/",
		func(c echo.Context) error {
			query := c.FormValue("q")
			e.Logger.Error(query)
			resID, err := logic.ProcessQuery(query)
			if err != nil {
				return c.String(
					http.StatusInternalServerError,
					err.Error(),
				)
			}

			page := strings.Replace(_indexPage, "%", resID, -1)
			return c.HTML(http.StatusOK, page)
		},
	)
	e.Logger.Fatal(e.Start(":1323"))
}
