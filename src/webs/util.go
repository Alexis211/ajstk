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
type redirectResponse string

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
	// fmt.Printf("%#v\n", d) // DEBUG

	if ee, ok := d.(getDataError); ok {
		w.WriteHeader(ee.Code)
		e := tpl["error"].Execute(w, giveTplData{"Sess": s, "Error": ee, "Request": req } )
		if e != nil { w.Write([]byte(e.String())) }
	} else if rurl, ok := d.(redirectResponse); ok {
		w.Header().Add("Location", string(rurl))
		w.WriteHeader(http.StatusFound)
	} else {
		e := tpl[v.tpl].Execute(w, giveTplData{"Sess": s, "Data": d, "Request": req } )
		if e != nil { w.Write([]byte(e.String())) }
	}
}

type redirectView struct {
	checkUniqueURL string
	process func(req *http.Request, s *session) string
}

func (v *redirectView) handle(w http.ResponseWriter, req *http.Request, s *session) {
	if v.checkUniqueURL != "" && req.URL.Path != v.checkUniqueURL {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "<pre>404 page not found</pre>")
		return
	}

	defer func() {
		if e := recover(); e != nil {
			fmt.Fprintf(w, "%v", e)
			log.Printf("Error in handling %v : %v", req.RawURL, e)
			gde, ok := e.(getDataError)
			if !ok {
				if ose, ok := e.(os.Error); ok {
					gde = getDataError{http.StatusInternalServerError, ose}
				} else {
					gde = getDataError{http.StatusInternalServerError, os.NewError(fmt.Sprintf("%v", e))}
				}
			}
			w.WriteHeader(gde.Code)
			e2 := tpl["error"].Execute(w, giveTplData{"Sess": s, "Error": gde, "Request": req } )
			if e2 != nil { w.Write([]byte(e2.String())) }
		}
	}()

	rurl := v.process(req, s)
	if rurl == "" { rurl = "/" }
	w.Header().Add("Location", rurl)
	w.WriteHeader(http.StatusFound)
}
