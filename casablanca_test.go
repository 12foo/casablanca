package main

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/labstack/echo"
	"github.com/labstack/echo/engine"
	"github.com/labstack/echo/test"
	"github.com/stretchr/testify/assert"
	"net/url"
	"os"
	"strings"
	"testing"
)

var server *echo.Echo

var testUser string = "testUser"
var testPW string = "testPW"

func scrapeLoginTicket(path string) (string, *goquery.Document) {
	req := test.NewRequest(echo.GET, path, nil)
	res := test.NewResponseRecorder()
	server.ServeHTTP(req, res)
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", doc
	}
	return doc.Find("input[name=lt]").AttrOr("value", ""), doc
}

func performLogin(ticket, service, username, password string) engine.Response {
	form := url.Values{
		"lt":       []string{ticket},
		"username": []string{username},
		"password": []string{password},
	}
	if service != "" {
		form.Set("service", service)
	}
	req := test.NewRequest(echo.POST, "/login", strings.NewReader(form.Encode()))
	req.Header().Set("Content-Type", "application/x-www-form-urlencoded")
	res := test.NewResponseRecorder()
	server.ServeHTTP(req, res)
	return res.Response
}

func TestLoginRoutine(t *testing.T) {
	service := "https://myservice.com/auth/"
	// user is sent to CAS from webapp
	ticket, doc := scrapeLoginTicket("/login?service=" + service)
	assert.NotEmpty(t, ticket)
	assert.Equal(t, service, doc.Find("input[name=service]").AttrOr("value", ""))
	// user logs in successfully
	res := performLogin(ticket, service, testUser, testPW)
	assert.Equal(t, 302, res.Status())
	assert.Contains(t, res.Header(), "Location")
	// CAS authenticates to webapp
	redirect, err := url.Parse(res.Header().Get("Location"))
	assert.NoError(t, err)
	assert.NotEmpty(t, redirect.Query().Get("ticket"))
}

func TestLoginRequestBase(t *testing.T) {
	// if service is not specified and session does not exist, SHOULD request credentials
	req := test.NewRequest(echo.GET, "/login", nil)
	res := test.NewResponseRecorder()
	server.ServeHTTP(req, res)
	// if service is not specified and session exists, SHOULD display "already logged in"
}

func TestLoginAccept(t *testing.T) {
	// fail: return to login as credential requestor

	// success (service specified): redirect to service with ticket in GET request
	ticket, _ := scrapeLoginTicket("/login")
	assert.NotEmpty(t, ticket)
	// res := performLogin(ticket, testUser, testPW)
	// assert.Equal(t, 302, )

	// success (service not specified): display "successfully logged in" message
}

func TestLoginRequestRenew(t *testing.T) {
}

func TestMain(m *testing.M) {
	cas := NewCAS()
	createTestData("/tmp/casablanca-test.sqlite3", testUser, testPW, "testuser@email.test")
	backend, err := NewDatabaseBackend(map[string]interface{}{
		"driver":       "sqlite3",
		"connection":   "/tmp/casablanca-test.sqlite3",
		"table":        "users",
		"username_col": "username",
		"password_col": "password",
		"extra": map[string]interface{}{
			"email": "email",
		},
	})
	if err != nil {
		panic(err)
	}
	cas.backends = append(cas.backends, backend)
	server = createServer(cas)
	server.SetDebug(true)
	result := m.Run()
	os.Remove("/tmp/casablanca-test.sqlite3")
	os.Exit(result)
}
