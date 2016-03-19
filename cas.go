package main

import (
	"github.com/patrickmn/go-cache"
	"github.com/satori/go.uuid"
	"time"
)

type CAS struct {
	tickets  *cache.Cache
	backends []Backend
}

func (cas *CAS) GenerateLoginTicket() string {
	u := "LT-" + uuid.NewV4().String()
	cas.tickets.Set(u, true, 15*time.Minute)
	return u
}

func (cas *CAS) Authenticate(username, password string) *AuthenticatedUser {
	for _, b := range cas.backends {
		if au := b.Authenticate(username, password); au != nil {
			return au
		}
	}
	return nil
}

func (cas *CAS) IsValidLoginTicket(ticket string) bool {
	if ticket[0:2] != "LT" {
		return false
	}
	_, ok := cas.tickets.Get(ticket)
	if ok {
		cas.tickets.Delete(ticket)
	}
	return ok
}

func (cas *CAS) GenerateServiceTicket(service string) string {
	u := "ST-" + uuid.NewV4().String()
	cas.tickets.Set(u, service, 5*time.Minute)
	return u
}

func (cas *CAS) IsValidServiceTicket(ticket, service string) bool {
	if ticket[0:2] != "ST" {
		return false
	}
	tkt, ok := cas.tickets.Get(ticket)
	if ok {
		cas.tickets.Delete(ticket)
	} else {
		return false
	}
	return tkt.(string) == service
}

func NewCAS() *CAS {
	cas := CAS{}
	cas.tickets = cache.New(cache.NoExpiration, 1*time.Minute)
	cas.backends = []Backend{}
	return &cas
}
