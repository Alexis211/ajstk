package util

import (
	"os"
	"log"
	"io/ioutil"
	"json"
	"bufio"
	"crypto/sha1"
	"io"
	"fmt"
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

func Exists(filename string) bool {
	_, err := os.Stat(filename)
	if err == nil {
		return true
	}
	return false
}

func StrSHA1(s string) string {
	b := sha1.New()
	io.WriteString(b, s)
	return fmt.Sprintf("%x", b.Sum())
}

func UUIDGen() string {
	f, _ := os.Open("/dev/urandom")
	b := make([]byte, 16)
	f.Read(b)
	f.Close()
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
