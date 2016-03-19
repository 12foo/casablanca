package main

import (
	"github.com/GeertJohan/go.rice"
	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	// "gopkg.in/yaml.v2"
	"html/template"
	"io"
	"net/http"
)

type Template struct {
	tpl *template.Template
}

func (t Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.tpl.ExecuteTemplate(w, name, data)
}

func requestCredentials(c echo.Context, cas *CAS) error {
	vals := map[string]interface{}{
		"lt": cas.GenerateLoginTicket(),
	}
	if service := c.Query("service"); service != "" {
		vals["service"] = service
	}
	return c.Render(http.StatusOK, "login", vals)
}

func tryLogin(cas *CAS) echo.HandlerFunc {
	return func(c echo.Context) error {
		return requestCredentials(c, cas)
	}
}

func submitCredentials(cas *CAS) echo.HandlerFunc {
	return func(c echo.Context) error {
		return nil
	}
}

func createServer(cas *CAS) *echo.Echo {
	staticBox := rice.MustFindBox("static")
	staticServer := http.StripPrefix("/static/", http.FileServer(staticBox.HTTPBox()))

	tplBox := rice.MustFindBox("templates")
	tpl := Template{}
	tpl.tpl = template.Must(template.New("login").Parse(tplBox.MustString("login.html")))

	e := echo.New()
	e.SetRenderer(tpl)
	e.Get("/login", tryLogin(cas))
	e.Post("/login", submitCredentials(cas))
	e.Get("/static/*", standard.WrapHandler(staticServer))
	return e
}

func main() {
	cas := NewCAS()
	s := createServer(cas)
	s.Run(standard.New(":3000"))
}
