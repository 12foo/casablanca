package main

import (
	"github.com/GeertJohan/go.rice"
	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	// "gopkg.in/yaml.v2"
	"html/template"
	"io"
	"net/http"
	"net/url"
)

type Template struct {
	tpl *template.Template
}

func (t Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.tpl.ExecuteTemplate(w, name, data)
}

func requestCredentials(vals map[string]interface{}, c echo.Context, cas *CAS) {
	vals["lt"] = cas.GenerateLoginTicket()
	if service := c.Query("service"); service != "" {
		vals["service"] = service
	}
}

func tryLogin(cas *CAS) echo.HandlerFunc {
	return func(c echo.Context) error {
		vals := map[string]interface{}{}
		requestCredentials(vals, c, cas)
		return c.Render(http.StatusOK, "login", vals)
	}
}

func submitCredentials(cas *CAS) echo.HandlerFunc {
	return func(c echo.Context) error {
		username := c.Form("username")
		password := c.Form("password")
		vals := map[string]interface{}{}
		if username != "" && password != "" {
			if au := cas.Authenticate(username, password); au != nil {
				if service := c.Form("service"); service != "" {
					// probably store this as a service ticket?
					st := cas.GenerateServiceTicket(service)
					if u, err := url.Parse(service); err != nil {
						u.Query().Set("ticket", st)
						return c.Redirect(302, u.String())
					}
				} else {
					// show "you are logged in" message
				}
			} else {
				vals["error"] = "Wrong username or password."
			}
		}
		requestCredentials(vals, c, cas)
		return c.Render(http.StatusUnauthorized, "login", vals)
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
