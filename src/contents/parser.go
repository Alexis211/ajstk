package contents

/* ************************************* TODO list : *************
  -	Select a template language to use, write stuff so that the parsed text is not
    just HTML, but a user-info-specific template.
  - Write parser for SRS items >> DONE, now connect to user-specific info
  - Write parser for inline furigana >> DONE, now connect to user-specific info
********************************* */

import (
	"strings"
	"fmt"
	"log"
)

func parseChunk(lines []string) *Chunk {
	ret := &Chunk{Title: lines[0], Summary: make([]string, 0, 10)}
	lines = lines[1:]

	ret.DescHTML, lines = parseText(lines, false, nil)
	ret.SRSItems = getSRSItems(lines)

	ret.Contents = ""
	for len(lines) != 0 && (len(lines) != 1 || lines[0] != "") {
		if lines[0][0:7] != "\\slide:" {
			log.Panicf("Error : expected '\\slide:' directive in chunk file.")
		}
		slideTitle := lines[0][7:]
		lines = lines[1:]
		slideText, nl := parseText(lines, true, ret)
		lines = nl

		ret.Contents += fmt.Sprintf(`
<div class="slide">
	<h1>%v</h1>
	%v
</div>
`, slideTitle, slideText)
		ret.Summary = append(ret.Summary, slideTitle)
	}

	return ret
}

func getSRSItems(lines []string) []SRSItem {
	ret := make([]SRSItem, 0)

	var busy []string = nil
	var long = false
	for _, text := range(lines) {
		if busy != nil {
			if len(text) > 2 && text[0:2] == "  " {
				busy = append(busy, text[2:])
			} else {
				log.Panicf("Error : expected continuation to srs item %v", busy)
			}
		} else if len(text) > 5 && text[0:5] == "\\srs:" {
			busy = strings.Split(text, ":", -1)
			busy = busy[1:]
			if len(busy) < 3 { long = true }
		}
		if busy != nil && len(busy) == 6 {
			ret = append(ret, SRSItem {
				busy[0], busy[1], busy[2], busy[3], busy[4], busy[5], long, makeHtmlWithFuri(busy[3], busy[4]),
			})
			busy = nil
			long = false
		}
	}
	return ret
}

func makeHtmlWithFuri(japanese, reading string) string {
	jrunes := make([]int, 0, len(japanese))
	rrunes := make([]int, 0, len(reading))
	for _, r := range japanese {
		jrunes = append(jrunes, r)
	}
	for _, r := range reading {
		rrunes = append(rrunes, r)
	}

	ret := ""
	for len(jrunes) > 0 && len(rrunes) > 0 {
		if jrunes[0] == rrunes[0] {
			if jrunes[0] != ' ' {	//we don't want no spaces in japanese
				ret += string(jrunes[0])
			}
			jrunes = jrunes[1:]
			rrunes = rrunes[1:]
		} else {
			fij := 0
			fir := 0
			boucle2:
			for ; fij < len(jrunes); fij++ {
				for fir = 0; fir < len(rrunes); fir++ {
					if jrunes[fij] == rrunes[fir] {
						break boucle2
					}
				}
			}

			ret += "<span class=\"furi\"><div>"
			for i := 0; i < fir; i++ {
				ret += string(rrunes[i])
			}
			ret += "</div>"
			for i := 0; i < fij; i++ {
				ret += string(jrunes[i])
			}
			ret += "</span>"

			jrunes = jrunes[fij:]
			rrunes = rrunes[fir:]
		}
	}

	for len(jrunes) > 0 {
		ret += string(jrunes[0])
		jrunes = jrunes[1:]
	}

	return ret
}

// ************* PARSING

func parseText(lines []string, allowSRSItems bool, c *Chunk) (html string, nextlines []string) {
	temp := &parsingText{chunk: c, root: true}

	last := 0
	ignoreUpTo := -1
	for id, text := range lines {
		last = id
		if id <= ignoreUpTo { continue }
		if text == "" {
			temp.finish()
		} else if text[0] == '\\' {
			temp.finish()
			if len(text) > 5 && text[0:5] == "\\srs:" {
				if !allowSRSItems {
					log.Panicf("Error : cannot place SRS item here.")
				} else {
					ignoreUpTo = temp.parseSRS(lines, id)
				}
			} else {
				break
			}
		} else if len(text) > 2 && text[0:2] == "  " {
			if temp.root_tag != "ol" && temp.root_tag != "ul" {
				temp.setRootTag("blockquote")
			}
			temp.push(text[2:])
		} else if text[0] == ':' {
			temp.setRootTag("table")
			temp.parseTableLine(text)
		} else if text[0] == '-' {
			temp.setRootTag("ul")
			temp.startLI()
			temp.push(text[1:])
		} else if text[0] == '#' {
			temp.setRootTag("ol")
			temp.startLI()
			temp.push(text[1:])
		} else if text[0] == '=' {
			n := 1
			for ;text[n] == '='; n++ {}
			temp.setRootTag(fmt.Sprintf("h%v", n+1))
			temp.push(text[n:])
			temp.finish()	// titles do not span several lines.
		} else {
			temp.setRootTag("p")
			temp.push(text)
		}
	}
	temp.finish()
	if last == len(lines) - 1 {
		return temp.result, []string{}
	}
	return temp.result, lines[last:]
}

type parsingText struct {
	chunk *Chunk		// so that we can have info about SRS furigana items
	root bool
	result string
	root_tag string		// either blockquote, table, ul, ol, or p
	temp_txt string
	is_in_li bool
}

func(t *parsingText) finish() {
	t.finishTempTxt()
	if t.root_tag != "" {
		t.result += "</" + t.root_tag + ">\n"
		t.root_tag = ""
	}
}

func (t *parsingText) setRootTag(tag string) {
	if t.root_tag == tag { return }
	t.finish()
	t.root_tag = tag
	t.result += "<" + tag + ">"
}

func (t *parsingText) finishTempTxt() {
	t.temp_txt = t.parseForSpan(t.temp_txt, false)
	if t.is_in_li {
		t.result += "</li>"
		t.is_in_li = false
	}
}

func (t *parsingText) startLI() {
	t.finishTempTxt()
	t.result += "<li>"
	t.is_in_li = true
}

func (t *parsingText) push(s string) {
	if len(s) == 0 { return }
	if s[len(s)-1] == '~' {
		t.temp_txt += s[:len(s)-1] + "<br \\/>\n"
	} else {
		t.temp_txt += s + "\n"
	}
}

func (t *parsingText) parseTableLine(s string) {
	t.result += "<tr>"
	for s != "" && s != ":" {
		s = s[1:]
		is_th := false
		if s[0] == '#' {
			s = s[1:]
			is_th = true
			t.result += "<th>"
		} else {
			t.result += "<td>"
		}
		s = t.parseForSpan(s, true)
		if is_th {
			if s[0] != '#' {
				log.Panicf("Error : expecting '#' on both sides of <th> cell.")
			}
			s = s[1:]
			if len(s) > 0 && s[0] != ':' {
				log.Panicf("Error : expecting ':' after end of <th> delimited by '#'.")
			}
			t.result += "</th>"
		} else {
			t.result += "</td>"
		}
	}
	t.result += "</tr>\n"
}

func (t *parsingText) parseForSpan(s string, isTable bool) string {
	ret := ""
	escape := false
	ignoreUpTo := -1

	stack := make([]string, 0)

	for pos, char := range s {
		if pos <= ignoreUpTo { continue }
		if escape {
			t.result += string(char)
			escape = false
		} else if char == '\\' {
			escape = true
		} else if isTable && (char == '#' || char == ':') {
			ret = s[pos:]
			break
		} else if char == '*' {
			stack = t.moarStack("strong", stack)
		} else if char == '/' {
			stack = t.moarStack("em", stack)
		} else if char == '{' || char == '｛' {
			ignoreUpTo = t.parseFurigana(s, pos)
		} else {
			t.result += string(char)
		}
	}
	if len(stack) != 0 {
		log.Panicf("Unfinished span elements : %v", stack)
	}
	return ret
}

func (t *parsingText) moarStack(s string, stack []string) []string {
	newstack := make([]string, 0, len(stack) + 1)

	pos := -1
	for p, e := range stack {
		if e == s {
			pos = p
			break
		} else {
			newstack = append(newstack, e)
		}
	}

	if pos == -1 {
		newstack = append(newstack, s)
		t.result += "<" + s + ">"
	} else {
		for i := len(stack) - 1; i > pos; i-- {
			t.result += "</" + stack[i] + ">"
		}
		t.result += "</" + s + ">"
		for i := pos+1; i < len(stack); i++ {
			t.result += "<" + stack[i] + ">"
			newstack = append(newstack, stack[i])
		}
	}
	return newstack
}

func (t *parsingText) parseFurigana(s string, pos int) int {
	ret := pos
	parts := []string{""}
	escape := false
	for p, char := range s {
		if p < pos+1 { continue }

		if escape {
			parts[len(parts)-1] += string(char)
			escape = false
		} else if char == '\\' {
			escape = true
		} else if char == ':' || char == '：' {
			if len(parts) > 1 {
				log.Panicf("Error : only one ':' is expected in furigana syntax.")
			}
			parts = append(parts, "")
		} else if char == '}' || char == '｝' {
			ret = p
			break
		} else {
			parts[len(parts)-1] += string(char)
		}
	}
	if len(parts) == 2 {
		t.result += makeHtmlWithFuri(parts[0], parts[1])
	} else {
		found := false
		for _, f := range(t.chunk.SRSItems) {
			if f.Japanese == parts[0] {
				t.result += f.HTMLWithFuri
				found = true
			}
		}
		if !found {
			log.Panicf("Error : word %v cannot have automatic furigana.", parts[0])
		}
	}
	return ret
}

func (t *parsingText) parseSRS(lines []string, start int) int {
	if lines[start][0:5] != "\\srs:" {
		fmt.Errorf("Fail. should not appear I believe.")
	}
	fields := strings.Split(lines[start][5:], ":", -1)
	class := "srs_item"
	if len(fields) < 3 { class = "srs_item_big" }
	pos := start + 1
	for {
		if pos < len(lines) && len(lines[pos]) > 2 && lines[pos][0:2] == "  " {
			fields = append(fields, lines[pos][2:])
			if len(fields) == 6 { break }
		} else {
			pos--
			break
		}
		pos++
	}
	if len(fields) != 6 {
		fmt.Errorf("Fail. Expected continuation to SRS item data.")
	}
	t.result += fmt.Sprintf(
`<div class="%v">
	<h6>%v / %v</h6>
	<strong>%v</strong>
	%v
	<div class="comment">%v</div>
</div>
`, class, fields[0], fields[1], makeHtmlWithFuri(fields[3], fields[4]), fields[2], fields[5])
	return pos
}
