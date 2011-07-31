package webs

import (
	"http"
)

// various levels of authentication
const (
	SA_GUEST = iota
	SA_USER
	SA_ADMIN
)

type session struct {
	auth int	// one of SA_* defined above
	username string
}

type sessionViewHandler interface {
	handle(w http.ResponseWriter, req *http.Request, s *session)
}

// implements http.Handler
type sessionView struct {
	h sessionViewHandler
}

func (v *sessionView) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// TODO : get session
	v.h.handle(w, req, nil)
}
