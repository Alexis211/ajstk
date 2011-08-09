package webs

import (
	"http"
	"template"
	"log"
)

import (
	"main/config"
	"util"
	"contents"
)

/* ***************************************************
		WEB SERVER MODULE FOR AJSTK - MAIN SOURCE FILE
		********************************************************* */


// ************************* TEMPLATES && OTHER DATA *********************

var tpl = map[string]*template.Template {
	"error": nil,

	"home": nil,
	"login": nil,
	"register": nil,
	"settings": nil,

	"study_home": nil,
	"browse": nil,
	"chunk_summary": nil,
	"chunk_read": nil,

	"dic_results": nil,
/*
	"srs_home": nil,
	"srs_review_drill": nil,*/
}

var Info struct {
	ExtraPages []string
	Messages map[string]string
}
var messages = make(map[string]string)

// ************************** IMPORTANT FUNCTIONS ****************

func Serve() {
	log.Printf("Starting web server at %v...", config.Conf.HTTPServeAddr)

	http.Handle("/", &sessionView{&tplView{"/", "home", homeView}} )
	for _, page := range Info.ExtraPages {
		http.Handle("/" + page, &sessionView{&tplView{"/" + page, page, homeView}} )
	}

	http.Handle("/login", &sessionView{&tplView{"/login", "login", loginView}} )
	http.Handle("/register", &sessionView{&tplView{"/register", "register", registerView}} )
	http.Handle("/logout", &sessionView{&redirectView{"/logout", logoutView}} )
	http.Handle("/settings", &sessionView{&tplView{"/settings", "settings", settingsView}} )

	http.Handle("/go_study/", &sessionView{&redirectView{"", goStudyView}} )
	http.Handle("/study_home", &sessionView{&tplView{"/study_home", "study_home", studyHomeView}})
	http.Handle("/browse/", &sessionView{&tplView{"", "browse", browseView}} )
	http.Handle("/chunk_summary/", &sessionView{&tplView{"", "chunk_summary", chunkSummaryView}} )
	http.Handle("/chunk_read/", &sessionView{&tplView{"", "chunk_read", chunkSummaryView}} )
	http.Handle("/go_chunk/", &sessionView{&redirectView{"", goChunkView}} )

	http.Handle("/dic", &sessionView{&tplView{"/dic", "dic_results", dicSearchView}})

	http.Handle("/image/", http.FileServer(config.Conf.WebFolder, ""))
	http.Handle("/style/", http.FileServer(config.Conf.WebFolder, ""))
	http.Handle("/js/", http.FileServer(config.Conf.WebFolder, ""))

	http.Handle("/reload_tpl", &sessionView{&redirectView{"/reload_tpl",
		func(req *http.Request, s *session) string {
			LoadWebFiles()
			return req.FormValue("back")
		}}})
	http.Handle("/reload_data", &sessionView{&redirectView{"/reload_data",
		func(req *http.Request, s *session) string {
			contents.Info, contents.Levels, contents.LevelsMap = contents.LoadData()
			return req.FormValue("back")
		}}})

	err := http.ListenAndServe(config.Conf.HTTPServeAddr, nil)
	if err != nil { log.Fatalf("Error while starting HTTP server : %v", err) }
}

func LoadWebFilesDir() {
	defer func() {
		if err := recover(); err != nil {
			log.Fatalf("Error while loading web files : %v", err)
		}
	}()

	LoadWebFiles()
}

func LoadWebFiles() {
	log.Printf("Loading web server templates...")
	util.LoadJSONFile(config.Conf.WebFolder + "/info.json", &Info)
	messages = Info.Messages
	for _, page := range Info.ExtraPages {
		tpl[page] = nil
	}

	for name, _ := range tpl {
		log.Printf("Loading template : %v...", name)
		tpl[name] = loadExtendedTemplate(config.Conf.WebFolder + "/tpl", name)
	}
}

type extendedTemplate struct {
	baseDir string
	chunks map[string][]string
	files map[string]bool
}

func loadExtendedTemplate(dir, name string) *template.Template {
	t := &extendedTemplate{ dir, make(map[string][]string), make(map[string]bool) }
	t.load(name)
	contents := t.resolve("main")
	return template.MustParse(contents, nil)
}

func (t *extendedTemplate) load(file string) {
	if _, ok := t.files[file]; ok { return }
	t.files[file] = true

	lines := util.ReadLines(t.baseDir + "/" + file + ".html")
	current := "main"
	if _, ok := t.chunks["main"]; !ok { t.chunks["main"] = make([]string, 0) }

	for _, l := range lines {
		if len(l) > 7 && l[0:7] == "~~LOAD " {
			t.load(l[7:])
		} else if len(l) > 9 && l[0:9] == "~~DEFINE " {
			current = l[9:]
			t.chunks[current] = make([]string, 0)
		} else {
			t.chunks[current] = append(t.chunks[current], l)
		}
	}
}

func (t *extendedTemplate) resolve(chunk string) string {
	data, ok := t.chunks[chunk]
	if !ok {
		log.Printf("Warning : call to unknown or recursively called chunk %v", chunk)
		return ""
	}
	t.chunks[chunk] = nil, false

	ret := ""
	for _, l := range data {
		if len(l) > 7 && l[0:7] == "~~CALL " {
			ret += t.resolve(l[7:])
		} else {
			ret += l + "\n"
		}
	}
	return ret
}
