package webs

import (
	"http"
	"strings"
	"os"
)

import (
	"contents"
	"dic"
	"study"
	"main/config"
)

func homeView(req *http.Request, s *session) interface{} {
	return nil
}

// =========================================================

func dicSearchView(req *http.Request, s *session) interface{} {
	dicfile := req.FormValue("dic")
	japanese := req.FormValue("j")
	meaning := req.FormValue("m")

	if dict, ok := dic.DicMap[dicfile]; ok {
		return giveTplData{
			"Entries": dict.LookUp(japanese, meaning),
			"Dic": dict,
			"Dics": dic.Dics,
		}
	}
	return getDataError{http.StatusNotFound, os.NewError("No such dictionnary.")}
}

// ==================================================

func loginView(req *http.Request, s *session) interface{} {
	rurl := req.FormValue("back")
	if rurl == "" { rurl = "/" }

	if req.FormValue("username") != "" && req.FormValue("password") != "" {
		user := study.GetUser(req.FormValue("username"))
		if user != nil && user.CheckPass(req.FormValue("password")) {
			s.User = user
			s.Admin = false
			if user.GetAttr("admin") == study.DBTRUE {
				if _, ok := config.Conf.AdminUsers[user.Username]; ok {
					s.Admin = true
				} else {
					user.SetAttr("admin", study.DBFALSE)
				}
			}
			return redirectResponse(rurl)
		}
		return giveTplData{
			"ReturnTo": rurl,
			"Error": messages["BadUserOrPass"],
		}
	}
	return giveTplData{
		"ReturnTo": rurl,
	}
}

func registerView(req *http.Request, s *session) interface{} {
	return redirectResponse("/")
}

func logoutView(req *http.Request, s *session) interface{} {
	s.User = nil
	s.Admin = false

	rurl := req.FormValue("back")
	if rurl == "" { rurl = "/" }
	return redirectResponse(rurl)
}

// ==================================================

func browseView(req *http.Request, s *session) interface {} {
	path := strings.Split(req.URL.Path, "/", -1)

	lvl := contents.LevelsMap[contents.Info.DefaultLevel]
	if len(path) >= 3 && path[2] != "" {
		if l, ok := contents.LevelsMap[path[2]]; ok {
			lvl = l
		} else {
			return getDataError{http.StatusNotFound, os.NewError("No such level.")}
		}
	}

	if len(lvl.Lessons) == 0 {
		return giveTplData{
			"Level": lvl,
			"Lesson": nil,
			"Levels": contents.Levels,
		}
	}

	less := lvl.Lessons[0]
	if len(path) >= 4 && path[3] != "" {
		if l, ok := lvl.LessonsMap[path[3]]; ok {
			less = l
		} else {
			return getDataError{http.StatusNotFound, os.NewError("No such lesson.")}
		}
	}

	return giveTplData{
		"Level": lvl,
		"Lesson": less,
		"Levels": contents.Levels,
	}
}

func chunkSummaryView(req *http.Request, s *session) interface{} {
	path := strings.Split(req.URL.Path, "/", -1)

	if len(path) < 5 || path[4] == "" {
		return getDataError{http.StatusNotFound, nil}
	}

	lvl, ok := contents.LevelsMap[path[2]]
	if !ok { return getDataError{http.StatusNotFound, os.NewError(messages["NoSuchLevel"])} }
	less, ok := lvl.LessonsMap[path[3]]
	if !ok { return getDataError{http.StatusNotFound, os.NewError(messages["NoSuchLesson"])} }
	chunk, ok := less.ChunksMap[path[4]]
	if !ok { return getDataError{http.StatusNotFound, os.NewError(messages["NoSuchChunk"])} }

	return giveTplData{
		"Level": lvl,
		"Lesson": less,
		"Chunk": chunk,
	}
}
