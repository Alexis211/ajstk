package contents

import (
	"log"
	"util"
)

// ***************************** STRUCTURES *********************

type DataInfo struct {
	SRSGroups map[string][]string
	Levels []string
	DefaultLevel string
	Dictionaries map[string]string
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

type Chunk struct {
	Id string
	Lesson *Lesson
	Level *Level
	Title, DescHTML string
	Contents string
	Summary []string
	SRSItems []SRSItem
}

type SRSItem struct {
	Group, Subgroup string
	Meaning, Japanese, Reading, Comment string
	Long bool
	HTMLWithFuri string
}

// *************************** WHERE THE DATA LIES **************

var contentsFolder string

var Info DataInfo
var Levels []*Level
var LevelsMap map[string]*Level

// **************************** LOAD FUNCTIONS *******************

func LoadDataFolder(f string) {
	log.Printf("Loading and parsing data files...")
	contentsFolder = f

	Info, Levels, LevelsMap = LoadData()
}

func LoadData() (DataInfo, []*Level, map[string]*Level) {
	var info DataInfo
	var levels []*Level
	var levelsMap = make(map[string]*Level)

	util.LoadJSONFile(contentsFolder + "/info.json", &info)

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

	util.LoadJSONFile(contentsFolder + "/" + level + "/info.json", &ret.Info)

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

	util.LoadJSONFile(contentsFolder + "/" + level + "/" + lesson + "/info.json", &ret.Info)

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

	lines := util.ReadLines(contentsFolder + "/" + level + "/" + lesson + "/" + chunk)

	return parseChunk(lines)
}
