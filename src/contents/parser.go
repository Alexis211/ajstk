package contents

/* ************************************* TODO list (finished) : *************
  -	Select a template language to use, write stuff so that the parsed text is not
    just HTML, but a user-info-specific template. >> DONE
  - Write parser for SRS items >> DONE, now connect to user-specific info >> DONE
  - Write parser for inline furigana >> DONE, now connect to user-specific info >> DONE
********************************* */

import (
	"strings"
	"fmt"
	"log"
	"template"
)

func parseChunk(lines []string, lesson *Lesson) *Chunk {
	ret := &Chunk{Title: lines[0], ToC: make([]*ChunkSlide, 0, 10), Lesson: lesson}
	lines = lines[1:]

	ret.DescHTML, lines = parseText(lines, false, nil)
	ret.SRSItems = getSRSItems(lines)

	cont := ""
	partNum := 0
	for len(lines) != 0 && (len(lines) != 1 || lines[0] != "") {
		if lines[0][0:7] != "\\slide:" {
			log.Panicf("Error : expected '\\slide:' directive in chunk file.")
		}
		slideTitle := lines[0][7:]
		lines = lines[1:]
		slideText, nl := parseText(lines, true, ret)
		lines = nl

		partNum++
		cont += fmt.Sprintf(`
<div class="slide" id="part%v" name="part%v">
	<h1>%v</h1>
	%v
</div>
`, partNum, partNum, slideTitle, slideText)
		ret.ToC = append(ret.ToC, &ChunkSlide{partNum, slideTitle})
	}
	// fmt.Printf(cont)		//DEBUG
	ret.Contents = template.MustParse(cont, nil)

	return ret
}

var srsItemNumber = 1

func getSRSItems(lines []string) []*SRSItem {
	ret := make([]*SRSItem, 0)

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
			ret = append(ret, &SRSItem {
				busy[0], busy[1], busy[2], busy[3], busy[4], busy[5], long, makeHtmlWithFuri(busy[3], busy[4], false, nil), srsItemNumber,
			})
			srsItemNumber++
			busy = nil
			long = false
		}
	}
	return ret
}

func makeHtmlWithFuri(japanese, reading string, templated bool, srsItem *SRSItem) string {
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

			if templated {
				if srsItem == nil {
					ret += `<span class="furi"><span {.section User.AFNever}class="hide_furi"{.end}>`
				} else {
					ret += fmt.Sprintf(`
	<span class="furi"><span {.section User.LFNever}class="hide_furi"
		{.end}{.section User.LFNotKnown}{.section Known%v}class="hide_furi"
		{.end}{.end}{.section User.LFNotStudying}{.section Studying%v}class="hide_furi"{.end}{.end}>`, 
					srsItem.Number, srsItem.Number)
				}
			} else {
				ret += `<span class="furi"><span>`
			}
			for i := 0; i < fir; i++ {
				ret += string(rrunes[i])
			}
			ret += "</span>"
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
			} else if len(text) > 7 && text[0:7] == "\\class:" {
				temp.setRootTag(classTag(text[7:]))
			} else if len(text) > 7 && text[0:7] == "\\style:" {
				temp.setRootTag(styleTag(text[7:]))
			} else if text == "\\hidden" {
				temp.setRootTag(hiddenTag(hiddenTagNumber))
				hiddenTagNumber++
			} else {
				break
			}
		} else if len(text) > 2 && text[0:2] == "  " {
			if temp.root_tag == nil || !temp.root_tag.spansTwoSpaces() {
				temp.setRootTag(htmlTag{"blockquote", true})
			}
			temp.push(text[2:])
		} else if text[0] == ':' {
			temp.setRootTag(htmlTag{"table", false})
			temp.parseTableLine(text)
		} else if text[0] == '-' {
			temp.setRootTag(htmlTag{"ul", true})
			temp.startLI()
			temp.push(text[1:])
		} else if text[0] == '#' {
			temp.setRootTag(htmlTag{"ol", true})
			temp.startLI()
			temp.push(text[1:])
		} else if text[0] == '=' {
			n := 1
			for ;text[n] == '='; n++ {}
			temp.setRootTag(htmlTag{fmt.Sprintf("h%v", n+1), false})
			temp.push(text[n:])
			temp.finish()	// titles do not span several lines.
		} else {
			temp.setRootTag(htmlTag{"p", false})
			temp.push(text)
		}
	}
	temp.finish()
	if last == len(lines) - 1 {
		return temp.result, []string{}
	}
	return temp.result, lines[last:]
}

type rootTag interface {
	begin() string
	end() string
	sameAs (rootTag) bool
	spansTwoSpaces() bool
}
type htmlTag struct {
	tag string
	spans bool
}

func (t htmlTag) begin() string { return "<" + t.tag + ">" }
func (t htmlTag) end() string { return "</" + t.tag + ">" }
func (t htmlTag) spansTwoSpaces() bool { return t.spans }
func (t htmlTag) sameAs(t2 rootTag) bool {
	if tt, ok := t2.(htmlTag); ok && tt.tag == t.tag { return true }
	return false
}

type classTag string
func (t classTag) begin() string { return fmt.Sprintf(`<p class="%v">`, string(t)) }
func (t classTag) end() string { return "</p>" }
func (t classTag) spansTwoSpaces() bool { return true }
func (t classTag) sameAs(t2 rootTag) bool { return false }

type styleTag string
func (t styleTag) begin() string { return fmt.Sprintf(`<p style="%v">`, string(t)) }
func (t styleTag) end() string { return "</p>" }
func (t styleTag) spansTwoSpaces() bool { return true }
func (t styleTag) sameAs(t2 rootTag) bool { return false }

type hiddenTag int
var hiddenTagNumber = 1
func (t hiddenTag) begin() string {
	return fmt.Sprintf(
		`<blockquote>
			<a style="font-size: 0.8em" href="#" id="hs%v" 
				onclick="$('hidden%v').show();$('hs%v').hide();$('hh%v').show()">{Hidden}</a>
			<a style="font-size: 0.8em; display: none" href="#" id="hh%v"
				onclick="$('hidden%v').hide();$('hs%v').show();$('hh%v').hide()">{Hideable}</a>
			<div style="display: none" id="hidden%v">`, t, t, t, t, t, t, t, t, t)
}
func (t hiddenTag) end() string { return "</div></blockquote>" }
func (t hiddenTag) spansTwoSpaces() bool { return true }
func (t hiddenTag) sameAs(t2 rootTag) bool { return false }

type parsingText struct {
	chunk *Chunk		// so that we can have info about SRS furigana items
	root bool
	result string
	root_tag rootTag		// either blockquote, table, ul, ol, or p
	temp_txt string
	is_in_li bool
}

func(t *parsingText) finish() {
	t.finishTempTxt()
	if t.root_tag != nil {
		t.result += t.root_tag.end() + "\n"
		t.root_tag = nil
	}
}

func (t *parsingText) setRootTag(tag rootTag) {
	if t.root_tag != nil  && t.root_tag.sameAs(tag) { return }
	t.finish()
	t.root_tag = tag
	t.result += tag.begin()
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
		} else if char == '[' {
			ignoreUpTo = t.parseLink(s, pos)
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
		t.result += makeHtmlWithFuri(parts[0], parts[1], true, nil)
	} else {
		found := false
		for _, f := range(t.chunk.SRSItems) {
			if f.Japanese == parts[0] {
				t.result += makeHtmlWithFuri(f.Japanese, f.Reading, true, f)
				found = true
				break
			}
		}
		if !found {
			for _, c := range(t.chunk.Lesson.Chunks) {
				if c == t.chunk { break }
				for _, f := range c.SRSItems {
					if f.Japanese == parts[0] {
						t.result += makeHtmlWithFuri(f.Japanese, f.Reading, true, f)
						found = true
						break
					}
				}
				if found { break }
			}
		}
		if !found {
			log.Panicf("Error : word %v cannot have automatic furigana.", parts[0])
		}
	}
	return ret
}

func (t *parsingText) parseLink(s string, pos int) int {
	ret := pos
	str := ""
	escape := false
	for p, char := range s {
		if p < pos+1 { continue }
		if escape {
			str += string(char)
			escape = false
		} else if char == '\\' {
			escape = true
		} else if char == ']' {
			ret = p
			break
		} else {
			str += string(char)
		}
	}
	if str[0] == '!' {			// image
		t.result += fmt.Sprintf(`<img src="%v" />`, str[1:])
	} else {
		spl := strings.Split(str, "|", 2)
		if len(spl) == 2 {
			t.result += fmt.Sprintf(`<a href="%v">%v</a>`, spl[0], spl[1])
		} else {
			t.result += fmt.Sprintf(`<a href="%v">%v</a>`, str, str)
		}
	}
	return ret
}

func (t *parsingText) parseSRS(lines []string, start int) int {
	if lines[start][0:5] != "\\srs:" {
		log.Panicf("Fail. should not appear I believe.")
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
		log.Panicf("Fail. Expected continuation to SRS item data.")
	}
	num := 0
	for _, f := range t.chunk.SRSItems {
		if f.Japanese == fields[3] {
			num = f.Number
			break
		}
	}
	t.result += fmt.Sprintf(
`<div class="%v{.section Known%v} siknown{.or}{.section Studying%v} sib{Box%v}{.end}{.end}">
	<h6>%v / %v</h6>
	<strong>%v</strong>
	%v
	<div class="comment">%v</div>
</div>
`, class, num, num, num, fields[0], fields[1], makeHtmlWithFuri(fields[3], fields[4], false, nil), fields[2], fields[5])
	return pos
}
