package contents

import (
	"log"
	"util"
	"main/config"
	"template"
)

// ***************************** STRUCTURES *********************

type DicInfo struct {
	Filename, Name, Format string
}
type DataInfo struct {
	SRSGroups map[string][]string
	Levels []string
	DefaultLevel string
	Dictionaries []DicInfo
}

type Level struct {
	Id string
	Info struct {
		Name, Description string
		Lessons []string
	}
	Lessons []*Lesson
	LessonsMap map[string]*Lesson
}

type Lesson struct {
	Id string
	Level *Level
	Info struct {
		Title, Description string
		Deps []string
		Chunks []string
	}
	Chunks []*Chunk
	ChunksMap map[string]*Chunk
}
func (l *Lesson) FullId() string {
	return l.Level.Id + "/" + l.Id
}

type Chunk struct {
	Id string
	Lesson *Lesson
	Level *Level
	Title, DescHTML string
	Contents *template.Template
	Summary []string
	SRSItems []*SRSItem
}
func (c *Chunk) FullId() string {
	return c.Level.Id + "/" + c.Lesson.Id + "/" + c.Id
}
func (c *Chunk) HasSRS() bool {
	return len(c.SRSItems) > 0
}

type SRSItem struct {
	Group, Subgroup string
	Meaning, Japanese, Reading, Comment string
	Long bool
	HTMLWithFuri string
	Number int
}

// *************************** WHERE THE DATA LIES **************

var Info DataInfo
var Levels []*Level
var LevelsMap map[string]*Level

// **************************** LOAD FUNCTIONS *******************

func LoadDataFolder() {
	log.Printf("Loading and parsing data files...")

	Info, Levels, LevelsMap = LoadData()
}

func LoadData() (DataInfo, []*Level, map[string]*Level) {
	var info DataInfo
	var levels []*Level
	var levelsMap = make(map[string]*Level)

	util.LoadJSONFile(config.Conf.ContentsFolder + "/info.json", &info)

	levels = make([]*Level, len(info.Levels))
	for id, name := range info.Levels {
		levels[id] = loadLevel(name)
		levelsMap[name] = levels[id]
	}
	return info, levels, levelsMap
}

func loadLevel(level string) *Level {
	log.Printf("Loading level : %v...", level)
	ret := &Level{Id: level, LessonsMap: make(map[string]*Lesson)}

	util.LoadJSONFile(config.Conf.ContentsFolder + "/" + level + "/info.json", &ret.Info)

	ret.Lessons = make([]*Lesson, len(ret.Info.Lessons))
	for id, name := range ret.Info.Lessons {
		ret.Lessons[id] = loadLesson(level, name, ret)
		ret.LessonsMap[name] = ret.Lessons[id]
	}
	return ret
}

func loadLesson(level, lesson string, lp *Level) *Lesson {
	log.Printf("Loading lesson : %v / %v...", level, lesson)
	ret := &Lesson{Level: lp, Id: lesson, ChunksMap: make(map[string]*Chunk)}

	util.LoadJSONFile(config.Conf.ContentsFolder + "/" + level + "/" + lesson + "/info.json", &ret.Info)

	ret.Chunks = make([]*Chunk, len(ret.Info.Chunks))
	for id, name := range ret.Info.Chunks {
		ret.Chunks[id] = loadChunk(level, lesson, name)
		ret.ChunksMap[name] = ret.Chunks[id]
		ret.Chunks[id].Id = name
		ret.Chunks[id].Lesson = ret
		ret.Chunks[id].Level = lp
	}
	return ret
}

func loadChunk(level, lesson, chunk string) *Chunk {
	log.Printf("Loading chunk : %v / %v / %v...", level, lesson, chunk)

	lines := util.ReadLines(config.Conf.ContentsFolder + "/" + level + "/" + lesson + "/" + chunk)

	return parseChunk(lines)
}
