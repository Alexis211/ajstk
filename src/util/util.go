package util

import (
	"os"
	"log"
	"io/ioutil"
	"json"
	"bufio"
)

func LoadJSONFile(f string, w interface{}) {
	data, e := ioutil.ReadFile(f)
	if e != nil { log.Panicf("Error loading JSON file %v : %v", f, e) }
	e = json.Unmarshal(data, w)
	if e != nil { log.Panicf("JSON error in %v : %v", f, e) }
}

func ReadLines(f string) []string {
	ret := make([]string, 0, 10)

	file, e := os.Open(f)
	if e != nil { log.Panicf("Error opening file %v : %v", f, e) }

	rd := bufio.NewReader(file)
	for {
		line, e := rd.ReadString('\n')
		if e != nil { break }
		ret = append(ret, TrimCRLF(line))
	}

	return ret
}

func TrimCRLF(s string) string {
	for len(s) != 0 && (s[len(s)-1] == '\r' || s[len(s)-1] == '\n') {
		s = s[0:len(s)-1]
	}
	return s
}
