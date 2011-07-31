package webs

import (
	"http"
	"strings"
	"os"
)

import (
	"contents"
)

func homeView(req *http.Request, s *session) interface{} {
	return nil
}

// =========================================================

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
	if !ok { return getDataError{http.StatusNotFound, os.NewError("No such level.")} }
	less, ok := lvl.LessonsMap[path[3]]
	if !ok { return getDataError{http.StatusNotFound, os.NewError("No such lesson.")} }
	chunk, ok := less.ChunksMap[path[4]]
	if !ok { return getDataError{http.StatusNotFound, os.NewError("No such chunk.")} }

	return giveTplData{
		"Level": lvl,
		"Lesson": less,
		"Chunk": chunk,
	}
}
