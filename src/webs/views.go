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
	"util"
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
	rurl := req.FormValue("back")
	if rurl == "" { rurl = "/" }

	ret := giveTplData{
		"ReturnTo": rurl,
	}

	if req.FormValue("username") != "" {
		if req.FormValue("password") == "" {
			ret["Error"] = messages["MustEnterPassword"]
		} else if req.FormValue("password2") != req.FormValue("password") {
			ret["Error"] = messages["MustEnterSamePasswordTwice"]
		} else if req.FormValue("email") == "" {
			ret["Error"] = messages["MustEnterEmail"]
		} else if study.GetUser(req.FormValue("username")) != nil {
			ret["Error"] = messages["UsernameTaken"]
		} else {
			u := study.CreateUser(req.FormValue("username"))
			u.SetAttr("admin", study.DBFALSE)
			u.SetAttr("password", util.StrSHA1(req.FormValue("password")))
			u.SetAttr("email", req.FormValue("email"))
			u.SetAttr("fullname", req.FormValue("fullname"))
			return redirectResponse("/login?back=" + rurl)
		}
	}

	return ret
}

func logoutView(req *http.Request, s *session) string {
	s.User = nil
	s.Admin = false

	rurl := req.FormValue("back")
	if rurl == "" { rurl = "/" }
	return rurl
}

// ==================================================

func goStudyView(req *http.Request, s *session) string {
	if s.User == nil {
		return "/login?back=" + req.URL.Path
	}

	path := strings.Split(req.URL.Path, "/", -1)
	if len(path) < 4 {
		panic(getDataError{http.StatusNotFound, os.NewError("Malformed request.")})
	}

	lvl, ok := contents.LevelsMap[path[2]]
	if !ok {
		panic(getDataError{http.StatusNotFound, os.NewError("No such level.")})
	}

	less, ok := lvl.LessonsMap[path[3]]
	if !ok {
		panic(getDataError{http.StatusNotFound, os.NewError("No such lesson.")})
	}

	s.User.StartStudyingLesson(less)
	s.User.SetAttr("current_study_lvl", lvl.Id)
	s.User.SetAttr("current_study_less", less.Id)

	return "/study_home"
}

func studyHomeView(req *http.Request, s *session) interface{} {
	if s.User == nil {
		return redirectResponse("/login?back=study_home")
	}

	//TODO here : get SRS status
	ret := giveTplData{
	}

	lvl, ok := contents.LevelsMap[s.User.GetAttr("current_study_lvl")]
	if !ok { return ret }
	less, ok := lvl.LessonsMap[s.User.GetAttr("current_study_less")]
	if !ok { return ret }

	ret["Level"] = lvl
	ret["Lesson"] = less
	ret["LessonsWS"] = s.User.GetLessonStatuses(lvl)
	ret["ChunksWS"] = s.User.GetChunkStatuses(less)
	ret["LessonStudy"] = study.LessonWithStatus{less, s.User.GetLessonStudy(less)}
	return ret
}

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

	ret := giveTplData{
		"Levels": contents.Levels,
		"Level": lvl,
		"Lesson": less,
		"LessonsWS": s.User.GetLessonStatuses(lvl),
		"ChunksWS": s.User.GetChunkStatuses(less),
	}
	if s.User != nil {
		ret["LessonStudy"] = study.LessonWithStatus{less, s.User.GetLessonStudy(less)}
	}
	return ret
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

	ret := giveTplData{
		"Level": lvl,
		"Lesson": less,
		"Chunk": chunk,
		"Hi": "hi",
	}
	if s.User != nil && s.User.GetChunkStudy(chunk) != study.CS_NOT_AVAILABLE {
		ret["Study"] = study.ChunkWithStatus{chunk, s.User.GetChunkStudy(chunk), s.User}
	}
	return ret
}

func goChunkView(req *http.Request, s *session) string {
	path := strings.Split(req.URL.Path, "/", -1)

	if len(path) < 4 || path[5] == "" {
		panic(getDataError{http.StatusNotFound, os.NewError("Malformed request")})
	}

	lvl, ok := contents.LevelsMap[path[2]]
	if !ok { panic(getDataError{http.StatusNotFound, os.NewError(messages["NoSuchLevel"])}) }
	less, ok := lvl.LessonsMap[path[3]]
	if !ok { panic(getDataError{http.StatusNotFound, os.NewError(messages["NoSuchLesson"])}) }
	chunk, ok := less.ChunksMap[path[4]]
	if !ok { panic(getDataError{http.StatusNotFound, os.NewError(messages["NoSuchChunk"])}) }

	if s.User == nil {
		panic(getDataError{http.StatusForbidden, os.NewError("Reserved to logged in users.")})
	}
	if path[5] == "reading" { s.User.SetChunkStatus(chunk, study.CS_READING) }
	if path[5] == "repeat" { s.User.SetChunkStatus(chunk, study.CS_REPEAT) }
	if path[5] == "done" { s.User.SetChunkStatus(chunk, study.CS_DONE) }
	return "/chunk_summary/" + chunk.FullId()
}
