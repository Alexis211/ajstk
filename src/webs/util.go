package webs

import (
	"http"
	"os"
	"log"
	"fmt"
)

// VARIOUS UTILITIES FOR MAKING WRITING VIEWS EASIER

type giveTplData map[string]interface{}

type tplView struct {
	checkUniqueURL string
	tpl string
	getData func(req *http.Request, s *session) interface{}
}

type getDataError struct {
	Code int
	Error os.Error
}

func (v *tplView) handle(w http.ResponseWriter, req *http.Request, s *session) {
	if v.checkUniqueURL != "" && req.URL.Path != v.checkUniqueURL {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "<pre>404 page not found</pre>")
		return
	}

	defer func() {
		if e := recover(); e != nil {
			fmt.Fprintf(w, "%v", e)
			log.Printf("Error in handling %v : %v", req.RawURL, e)
		}
	}()
	d := v.getData(req, s)
	fmt.Printf("%#v\n", d)

	if ee, ok := d.(getDataError); ok {
		w.WriteHeader(ee.Code)
		e := tpl["error"].Execute(w, giveTplData{"Sess": s, "Error": ee, "Request": req } )
		if e != nil { w.Write([]byte(e.String())) }
	} else {
		e := tpl[v.tpl].Execute(w, giveTplData{"Sess": s, "Data": d, "Request": req } )
		if e != nil { w.Write([]byte(e.String())) }
	}
}


type redirectPrevView struct {
	checkUniqueURL string
	fct func(req *http.Request, s *session)
}

func (v *redirectPrevView) handle(w http.ResponseWriter, req *http.Request, s *session) {
	if v.checkUniqueURL != "" && req.URL.RawPath[0:len(v.checkUniqueURL)] != v.checkUniqueURL {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "<pre>404 page not found</pre>")
		return
	}

	defer func() {
		if e := recover(); e != nil {
			fmt.Fprintf(w, "%v", e)
			log.Printf("Error in handling %v : %v", req.RawURL, e)
		}
	}()

	v.fct(req, s)
	url := req.URL.RawPath[len(v.checkUniqueURL):]
	if url == "" { url = "/" }
	url2 := req.Header.Get("Http-Referer")
	if url2 != "" { url = url2 }
	w.Header().Add("Location", url)
	w.WriteHeader(http.StatusMovedPermanently)
}
