package main

import (
	"html/template"
	"net/http"
	"os"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/rprtr258/twitch-emotes-modifier-plugin/internal/logic"
)

var _indexPage = template.Must(template.New("").Parse(`<!DOCTYPE html><html>
<head><style>
body {
	background: #040404;
	color: white;
	display: flex;
	align-items: center;
	justify-content: center;
	flex-direction: column;
}
form {
	width: 90%;
}
textarea {
	width: 90%;
}
button {
	width: 9%;
}
</style></head>
<body>
	<form action="/" method="POST">
		<textarea name="q">{{.Query}}</textarea>
		<button>send</button>
	</form>
	{{range .Imgs}}
		<span>{{.}}</span>
		<img src="/static/{{.}}.webp" />
	{{end}}
</body>
</html>`))

type indexPage struct {
	Query string
	Imgs  []string
}

func getFileSystem() http.FileSystem {
	return http.FS(os.DirFS("."))
}

func renderIndexPage(query string) (string, error) {
	entries, err := os.ReadDir(".")
	if err != nil {
		return "", err
	}

	imgs := make([]string, 0, len(entries))
	for _, entry := range entries {
		name, _, ok := strings.Cut(entry.Name(), ".webp")
		if !ok {
			continue
		}

		imgs = append(imgs, name)
	}

	sb := strings.Builder{}
	if err := _indexPage.Execute(&sb, indexPage{
		Query: query,
		Imgs:  imgs,
	}); err != nil {
		return "", err
	}

	return sb.String(), nil
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
			page, err := renderIndexPage("")
			if err != nil {
				return c.String(
					http.StatusInternalServerError,
					err.Error(),
				)
			}

			return c.HTML(http.StatusOK, page)
		},
	)
	e.POST(
		"/",
		func(c echo.Context) error {
			query := c.FormValue("q")
			e.Logger.Error(query)

			_, err := logic.ProcessQuery(query)
			if err != nil {
				return c.String(
					http.StatusInternalServerError,
					err.Error(),
				)
			}

			page, err := renderIndexPage("")
			if err != nil {
				return c.String(
					http.StatusInternalServerError,
					err.Error(),
				)
			}

			return c.HTML(http.StatusOK, page)
		},
	)
	e.Logger.Fatal(e.Start(":1323"))
}
