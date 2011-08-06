package dic

import (
	"log"
	"strings"
)

import (
	"util"
	"main/config"
	"contents"
)

// **************** STRUCTURES ****************

type DicEntry struct {
	Japanese string
	Reading string
	Meaning []string
}

type dicIndex map[string][]*DicEntry

type Dictionary struct {
	Id, Name, Format string
	fromJapanese, fromMeaning,fromMeaningExtended dicIndex
}

// **** WHERE THE DATA IS

var DicMap = make(map[string]*Dictionary)
var Dics []*Dictionary

// **************** METHODS *******************

func LoadDictionaries() {
	// Load from config.Conf.ContentsFolder/dic
	// In contents.Info.Dictionaries

	for _, d := range contents.Info.Dictionaries {
		dic := &Dictionary{d.Filename, d.Name, d.Format, make(dicIndex), make(dicIndex), make(dicIndex)}
		dic.load()
		DicMap[d.Filename] = dic
		Dics = append(Dics, dic)
	}
	log.Printf("Finished loading dictionaries.")
}

func (d *Dictionary) load() {
	log.Printf("Loading dictionary : %v (%v)...", d.Id, d.Name)

	in := make(chan *DicEntry)
	go parseEdict(config.Conf.ContentsFolder + "/dic/" + d.Id, in)

	for {
		e := <-in
		if e == nil { break }
		d.fromJapanese.add(e.Japanese, e)
		if e.Japanese != e.Reading {
			d.fromJapanese.add(e.Reading, e)
		}
		for _, m := range e.Meaning {
			s := strings.Split(m, " ", -1)
			for len(s) > 0 && len(s[0]) > 0 && s[0][0] == '(' {
				s = s[1:]
			}
			for _, kk := range s {
				d.fromMeaningExtended.add(kk, e)
			}
			d.fromMeaning.add(strings.Join(s, " "), e)
		}
	}
}

func (i dicIndex) add(key string, e *DicEntry) {
	if key == "" { return }
	if _, ok := i[key]; ok {
		i[key] = append(i[key], e)
	} else {
		i[key] = []*DicEntry{e}
	}
}

// ========== EDICT FILE FORMAT PARSER

func parseEdict(filename string, c chan *DicEntry) {
	lines := util.ReadLines(filename)
	for _, line := range lines {
		e := parseEdictEntry(line)
		if e != nil {
			c <- e
		}
	}
	c <- nil
}

func parseEdictEntry(s string) *DicEntry {
	split := strings.Split(s, " ", 3)
	if len(split) != 3 { return nil }
	if split[1][0] == '/' {
		split[2] = split[1] + " " + split[2]
		split[1] = split[0]
	} else if split[1][0] == '[' && split[1][len(split[1])-1] == ']' {
		split[1] = split[1][1:len(split[1])-1]
	}
	meaning := strings.Split(split[2], "/", -1)
	for meaning[0] == "" {
		meaning = meaning[1:]
	}
	for meaning[len(meaning)-1] == "" {
		meaning = meaning[:len(meaning)-1]
	}
	return &DicEntry{split[0], split[1], meaning}
}

// ========== LOOKUP

func (d *Dictionary) LookUp(japanese, meaning string) []*DicEntry {
	if japanese == "" && meaning == "" {
		return nil
	} else if japanese == "" {
		return d.fromMeaning[meaning]
	} else if meaning == "" {
		return d.fromJapanese[japanese]
	}
	// lookup with both japanese and meaning
	els := make(map[*DicEntry]bool)
	for _, e := range append(d.fromMeaning[meaning], d.fromMeaningExtended[meaning]...) {
		if e.matchJ(japanese) {
			els[e] = true
		}
	}
	for _, e := range d.fromJapanese[japanese] {
		if e.matchM(meaning) {
			els[e] = true
		}
	}
	ret := make([]*DicEntry, 0, len(els))
	for e, _ := range els {
		ret = append(ret, e)
	}
	return ret
}

func (e *DicEntry) matchJ(s string) bool {
	if strings.Index(e.Japanese, s) != -1 { return true }
	if strings.Index(e.Reading, s) != -1 { return true }
	return false
}

func (e *DicEntry) matchM(s string) bool {
	for _, m := range e.Meaning {
		if strings.Index(m, s) != -1 { return true }
	}
	return false
}
