package webs

import (
	"http"
)

import (
	"study"
	"util"
)


/* LOGIN STATUSES :
	- Not logged in : session.User = nil
	- Standard user : session.User = something, Admin = false
	- Admin user    : session.User = something, Admin = true
*/
type session struct {
	Admin bool
	User *study.User
	Id string
}

type sessionViewHandler interface {
	handle(w http.ResponseWriter, req *http.Request, s *session)
}

// implements http.Handler
type sessionView struct {
	h sessionViewHandler
}

var sessions = make(map[string]*session)

func (v *sessionView) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	sessid := "" 
	for _, cookie := range req.Cookie {
		if cookie.Name == "sessid" {
			sessid = cookie.Value
			break
		}
	}
	if sessid == "" {
		sessid = util.UUIDGen()
		w.Header().Add("Set-Cookie", "sessid=" + sessid + "; path=/")
	}
	sess, ok := sessions[sessid]
	if !ok {
		sess = new(session)
		sess.Id = sessid
		sessions[sessid] = sess
	}
	v.h.handle(w, req, sess)
}
