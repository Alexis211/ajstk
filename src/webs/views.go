package webs

import (
	"http"
	"strings"
	"os"
	"fmt"
	"json"
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
	return getDataError{http.StatusNotFound, os.NewError(messages["NoSuchDictionary"])}
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

func settingsView(req *http.Request, s *session) interface{} {
	if s.User == nil {
		return redirectResponse("/login?back=/settings")
	}

	ret := giveTplData{
		"Email": s.User.GetAttr("email"),
		"FullName": s.User.GetAttr("fullname"),
	}

	if req.FormValue("email") != "" && req.FormValue("fullname") != "" {
		ret["Email"] = req.FormValue("email")
		ret["FullName"] = req.FormValue("fullname")
		s.User.SetAttr("email", req.FormValue("email"))
		s.User.SetAttr("fullname", req.FormValue("fullname"))
		ret["Message"] = messages["InfoUpdated"]
	}

	apw, pw1, pw2 := req.FormValue("apw"), req.FormValue("pw1"), req.FormValue("pw2")
	if apw != "" && pw1 != "" && pw2 != "" {
		if !s.User.CheckPass(apw) {
			ret["Error"] = messages["BadActualPassword"]
		} else if pw1 != pw2 {
			ret["Error"] = messages["MustEnterSamePasswordTwice"]
		} else {
			s.User.SetAttr("password", util.StrSHA1(pw1))
			ret["Message"] = messages["InfoUpdated"]
		}
	}

	lessfuri, ok1 := map[string]int{
		"always": study.LF_ALWAYS,
		"notknown": study.LF_NOT_KNOWN,
		"notstudying": study.LF_NOT_STUDYING,
		"never": study.LF_NEVER,
	}[req.FormValue("lessfuri")]
	anyfuri, ok2 := map[string]int {
		"always": study.AF_ALWAYS,
		"never": study.AF_NEVER,
	}[req.FormValue("anyfuri")]
	if ok1 && ok2 {
		s.User.SetLessFuri(lessfuri)
		s.User.SetAnyFuri(anyfuri)
		ret["Message"] = messages["InfoUpdated"]
	}

	return ret
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
		panic(getDataError{http.StatusNotFound, os.NewError(messages["NoSuchLevel"])})
	}

	less, ok := lvl.LessonsMap[path[3]]
	if !ok {
		panic(getDataError{http.StatusNotFound, os.NewError(messages["NoSuchLesson"])})
	}

	s.User.StartStudyingLesson(less)
	s.User.SetAttr("current_study_lvl", lvl.Id)
	s.User.SetAttr("current_study_less", less.Id)

	return "/study_home"
}

func studyHomeView(req *http.Request, s *session) interface{} {
	if s.User == nil {
		return redirectResponse("/login?back=/study_home")
	}

	ret := giveTplData{
		"Levels": contents.Levels,
		"SRSData": s.User.GetSRSStats(),
	}

	lvl, ok := contents.LevelsMap[s.User.GetAttr("current_study_lvl")]
	if !ok { return ret }
	less, ok := lvl.LessonsMap[s.User.GetAttr("current_study_less")]
	if !ok { return ret }

	ret["Level"] = lvl
	ret["Lesson"] = less
	ret["LessonsWS"] = s.User.GetLessonStatuses(lvl)
	chunksWS := s.User.GetChunkStatuses(less)
	ret["ChunksWS"] = chunksWS
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
			return getDataError{http.StatusNotFound, os.NewError(messages["NoSuchLevel"])}
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
			return getDataError{http.StatusNotFound, os.NewError(messages["NoSuchLesson"])}
		}
	}

	if s.User != nil && lvl.Id == s.User.GetAttr("current_study_lvl") && less.Id == s.User.GetAttr("current_study_less") {
		return redirectResponse("/study_home")
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

func chunkViewParseUrl(req *http.Request) (chunk *contents.Chunk, err getDataError) {
	path := strings.Split(req.URL.Path, "/", -1)

	if len(path) < 5 || path[4] == "" {
		return nil, getDataError{http.StatusNotFound, nil}
	}

	lvl, ok := contents.LevelsMap[path[2]]
	if !ok { return nil, getDataError{http.StatusNotFound, os.NewError(messages["NoSuchLevel"])} }
	less, ok := lvl.LessonsMap[path[3]]
	if !ok { return nil, getDataError{http.StatusNotFound, os.NewError(messages["NoSuchLesson"])} }
	chunk, ok = less.ChunksMap[path[4]]
	if !ok { return nil, getDataError{http.StatusNotFound, os.NewError(messages["NoSuchChunk"])} }
	return chunk, getDataError{http.StatusOK, nil}
}

func chunkSummaryView(req *http.Request, s *session) interface{} {
	chunk, err := chunkViewParseUrl(req)
	if err.Code != http.StatusOK { return err }

	d, e := json.Marshal(chunk.ToC)
	if e != nil { return getDataError{http.StatusInternalServerError, e} }
	ret := giveTplData{
		"Level": chunk.Level,
		"Lesson": chunk.Lesson,
		"Chunk": chunk,
		"ToCJSON": string(d),
	}
	if s.User != nil && s.User.GetChunkStudy(chunk) != study.CS_NOT_AVAILABLE {
		ret["Study"] = study.ChunkWithStatus{chunk, s.User.GetChunkStudy(chunk), s.User}
	}

	giveTextTpl := giveTplData{
		"User": s.User,
		"Hidden": messages["Hidden"],
		"Hideable": messages["Hideable"],
	}

	srsItemsWS := s.User.GetSRSItemStatuses(chunk)
	ret["SRSItemsWS"] = srsItemsWS
	for _, ss := range srsItemsWS {
		if ss.Studying() {
			giveTextTpl[fmt.Sprintf("Studying%v", ss.Item.Number)] = true
		}
		if ss.Known() {
			giveTextTpl[fmt.Sprintf("Known%v", ss.Item.Number)] = true
		}
		giveTextTpl[fmt.Sprintf("Box%v", ss.Item.Number)] = ss.Box
	}
	if s.User != nil {		// add SRS data from the rest of the lesson
		for _, c := range chunk.Lesson.Chunks {
			if c == chunk { break }
			sl := s.User.GetSRSItemStatuses(c)
			for _, ss := range sl {
				if ss.Studying() {
					giveTextTpl[fmt.Sprintf("Studying%v", ss.Item.Number)] = true
				}
				if ss.Known() {
					giveTextTpl[fmt.Sprintf("Known%v", ss.Item.Number)] = true
				}
			}
		}
	}
	text := util.StringWriter{""}
	chunk.Contents.Execute(&text, giveTextTpl)
	ret["ChunkText"] = text.Str

	return ret
}

func goChunkView(req *http.Request, s *session) string {
	path := strings.Split(req.URL.Path, "/", -1)
	if len(path) < 6 {
		panic(getDataError{http.StatusInternalServerError,
			os.NewError("Malformed request.")})
	}

	if s.User == nil {
		panic(getDataError{http.StatusForbidden, os.NewError("Reserved to logged in users.")})
	}

	chunk, err := chunkViewParseUrl(req)
	if err.Code != http.StatusOK { panic(err) }

	if path[5] == "reading" { s.User.SetChunkStatus(chunk, study.CS_READING) }
	if path[5] == "repeat" { s.User.SetChunkStatus(chunk, study.CS_REPEAT) }
	if path[5] == "done" { s.User.SetChunkStatus(chunk, study.CS_DONE) }
	return "/chunk_summary/" + chunk.FullId()
}

// ==============================================================

func srsViewsSplitPath(req *http.Request, needSubgroup bool) (group, subgroup string) {
	path := strings.Split(req.URL.Path, "/", -1)
	if needSubgroup {
		if len(path) < 4 || path[3] == "" {
			panic(getDataError{http.StatusNotFound, os.NewError("Malformed request")})
		}
		return path[2], path[3]
	}
	if len(path) < 3 || path[2] == "" {
		panic(getDataError{http.StatusNotFound, os.NewError("Malformed request")})
	}
	if len(path) >= 4 {
		return path[2], path[3]
	}
	return path[2], ""
}

func srsHomeView(req *http.Request, s *session) interface{} {
	if s.User == nil {
		return redirectResponse("/login?back=/srs_home")
	}

	srsStats := s.User.GetSRSStats()
	return giveTplData{
		"SRSData": srsStats,
	}
}

func srsActivateView(req *http.Request, s *session) string {
	if s.User == nil {
		return "/login?back=/srs_home"
	}

	group, subgroup := srsViewsSplitPath(req, true)
	s.User.SRSActivateSubgroup(group, subgroup)
	return "/srs_home"
}

func srsDeactivateView(req *http.Request, s *session) string {
	if s.User == nil {
		return "/login?back=/srs_home"
	}

	group, subgroup := srsViewsSplitPath(req, true)
	s.User.SRSDeactivateSubgroup(group, subgroup)
	return "/srs_home"
}

func srsReviewView(req *http.Request, s *session) interface{} {
	if s.User == nil {
		return redirectResponse("/login?back=/srs_home")
	}

	group, subgroup := srsViewsSplitPath(req, false)

	items := s.User.GetSRSReviewItems(group, subgroup)
	if len(items) == 0 {
		return getDataError{http.StatusNotFound, os.NewError(messages["NothingToStudy"])}
	}
	data, e := json.Marshal(items)
	if e != nil { return getDataError{http.StatusInternalServerError, e} }
	ret := giveTplData{
		"Data": string(data),
		"Group": strings.ToLower(group),
	}
	if subgroup != "" { ret["SubGroup"] = strings.ToLower(subgroup) }
	return ret
}

func srsTomorrowView(req *http.Request, s *session) interface{} {
	if s.User == nil {
		return redirectResponse("/login?back=/srs_home")
	}

	group, subgroup := srsViewsSplitPath(req, false)

	items := s.User.GetSRSTomorrowItems(group, subgroup)
	if len(items) == 0 {
		return getDataError{http.StatusNotFound, os.NewError(messages["NothingToStudy"])}
	}
	data, e := json.Marshal(items)
	if e != nil { return getDataError{http.StatusInternalServerError, e} }
	ret := giveTplData{
		"Data": string(data),
		"IsDrill": true,
		"Group": strings.ToLower(group),
	}
	if subgroup != "" { ret["SubGroup"] = strings.ToLower(subgroup) }
	return ret
}

func srsChunkDrillView(req *http.Request, s *session) interface{} {
	if s.User == nil {
		return redirectResponse("/login?back=/srs_home")
	}

	chunk, err := chunkViewParseUrl(req)
	if err.Code != http.StatusOK { return err }

	items := s.User.GetSRSChunkItemsDrill(chunk)
	if len(items) == 0 {
		return getDataError{http.StatusNotFound, os.NewError(messages["NothingToStudy"])}
	}
	data, e := json.Marshal(items)
	if e != nil { return getDataError{http.StatusInternalServerError, e} }
	return giveTplData{
		"Data": string(data),
		"IsDrill": true,
		"Chunk": chunk,
	}
}

func srsLessonDrillView(req *http.Request, s *session) interface{} {
	if s.User == nil {
		return redirectResponse("/login?back=/srs_home")
	}

	path := strings.Split(req.URL.Path, "/", -1)

	if len(path) < 4 || path[3] == "" {
		return getDataError{http.StatusNotFound, nil}
	}

	lvl, ok := contents.LevelsMap[path[2]]
	if !ok { return getDataError{http.StatusNotFound, os.NewError(messages["NoSuchLevel"])} }
	less, ok := lvl.LessonsMap[path[3]]
	if !ok { return getDataError{http.StatusNotFound, os.NewError(messages["NoSuchLesson"])} }

	items := s.User.GetSRSLessonItemsDrill(less)
	if len(items) == 0 {
		return getDataError{http.StatusNotFound, os.NewError(messages["NothingToStudy"])}
	}
	data, e := json.Marshal(items)
	if e != nil { return getDataError{http.StatusInternalServerError, e} }
	return giveTplData{
		"Data": string(data),
		"IsDrill": true,
		"Lesson": less,
	}
}

func srsSaveView(req *http.Request, s *session) interface{} {
	if s.User == nil {
		return getDataError{http.StatusForbidden, os.NewError("Reserved to logged in users.")}
	}

	var success, fail []int64
	e := json.Unmarshal([]byte(req.FormValue("success")), &success)
	if e != nil { panic(e) }
	e = json.Unmarshal([]byte(req.FormValue("fail")), &fail)
	if e != nil { panic(e) }

	s.User.UpdateSRSItemStatuses(success, fail)

	return giveTplData{"Status": "Data saved", "Success": success, "Fail": fail}
}
