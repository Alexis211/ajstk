package study

import (
	"github.com/kuroneko/gosqlite3"
)

import (
	"contents"
)

const (		//lesson possible statuses
	LS_NOT_AVAILABLE = iota
	LS_AVAILABLE
	LS_STUDYING
	LS_DONE
)

const (		//chunk possible statuses
	CS_NOT_AVAILABLE = iota
	CS_READING
	CS_REPEAT
	CS_DONE
)

type LessonWithStatus struct {
	Lesson *contents.Lesson
	Status int64
}
func (s LessonWithStatus) StatusStr() string {
	return []string{"not_available", "available", "studying", "done"}[s.Status]
}
func (s LessonWithStatus) Available() bool { return s.Status == LS_AVAILABLE }
func (s LessonWithStatus) Studying() bool { return s.Status == LS_STUDYING }
func (s LessonWithStatus) Done() bool { return s.Status == LS_DONE }

type ChunkWithStatus struct {
	Chunk *contents.Chunk
	Status int64
	User *User
}
func (s ChunkWithStatus) StatusStr() string {
	return []string{"not_available", "reading", "repeat", "done"}[s.Status]
}
func (s ChunkWithStatus) Reading() bool { return s.Status == CS_READING }
func (s ChunkWithStatus) Repeat() bool { return s.Status == CS_REPEAT }
func (s ChunkWithStatus) Done() bool { return s.Status == CS_DONE }
func (s ChunkWithStatus) RepeatDone() bool {
	return s.User.IsChunkSRSDone(s.Chunk)
}

func (u *User) checkStudyTables() {
	// lesson_study.id = "Level/Lesson"
	u.DBQuery(`
		CREATE TABLE IF NOT EXISTS 'lesson_study' (
			id VARCHAR(256) PRIMARY KEY,
			status INTEGER
		)`)
	// chunk_study.id = "Level/Lesson/Chunk"
	u.DBQuery(`
		CREATE TABLE IF NOT EXISTS 'chunk_study' (
			id VARCHAR(256) PRIMARY KEY,
			status INTEGER
		)`)
}

func (u *User) loadStudyStatus() {
	u.lessonStudy = make(map[string]int64)
	/* do :
		1. load table lesson_study in u.lessonStudy
		2. make sure every lesson has an entry in there
		3. check lesson with NOT_AVAILABLE status to see if they can be AVAILABLE
	*/
	st := u.DBQuerySt("SELECT id, status FROM lesson_study")
	for st.Step() == sqlite3.ROW {
		row := st.Row()
		u.lessonStudy[row[0].(string)] = row[1].(int64)
	}
	for _, level := range contents.Levels {
		for _, lesson:= range level.Lessons {
			u.GetLessonStudy(lesson)
		}
	}
	u.CheckNewAvailableLessons()

	u.chunkStudy = make(map[string]int64)
	st = u.DBQuerySt("SELECT id, status FROM chunk_study")
	for st.Step() == sqlite3.ROW {
		row := st.Row()
		u.chunkStudy[row[0].(string)] = row[1].(int64)
	}
	u.CheckDoneLessons()
}

// ============================

func (u *User) SetLessonStudy(less *contents.Lesson, status int64) {
	id := less.FullId()
	if v, ok := u.lessonStudy[id]; !ok || v != status {
		if el, _ := u.DBQueryFetchOne("SELECT status FROM lesson_study WHERE id = ?", id); el != nil {
			u.DBQuery("UPDATE lesson_study SET status = ? WHERE id = ?", status, id)
		} else {
			u.DBQuery("INSERT INTO lesson_study(id, status) VALUES(?, ?)", id, status)
		}
	}
	u.lessonStudy[id] = status
}

func (u *User) GetLessonStudy(less *contents.Lesson) int64 {
	if s, ok := u.lessonStudy[less.FullId()]; ok {
		return s
	}
	u.SetLessonStudy(less, LS_NOT_AVAILABLE)
	return LS_NOT_AVAILABLE
}

func (u *User) SetChunkStudy(chunk *contents.Chunk, status int64) {
	id := chunk.FullId()
	if v, ok := u.chunkStudy[id]; !ok || v != status {
		if el, _ := u.DBQueryFetchOne("SELECT status FROM chunk_study WHERE id = ?", id); el != nil {
			u.DBQuery("UPDATE chunk_study SET status = ? WHERE id = ?", status, id)
		} else {
			u.DBQuery("INSERT INTO chunk_study(id, status) VALUES(?, ?)", id, status)
		}
	}
	u.chunkStudy[id] = status
}

func (u *User) GetChunkStudy(chunk *contents.Chunk) int64 {
	if s, ok := u.chunkStudy[chunk.FullId()]; ok {
		return s
	}
	return CS_NOT_AVAILABLE
}

func (u *User) CheckNewAvailableLessons() {
	for _, level := range contents.Levels {
		for _, lesson := range level.Lessons {
			if u.GetLessonStudy(lesson) <= LS_AVAILABLE {
				available := true
				for _, dep := range lesson.Info.Deps {
					if v, ok := u.lessonStudy[level.Id + "/" + dep]; ok && v < LS_DONE {
						available = false
						break
					}
				}
				if available {
					u.SetLessonStudy(lesson, LS_AVAILABLE)
				} else {
					u.SetLessonStudy(lesson, LS_NOT_AVAILABLE)
				}
			}
		}
	}
}

func (u *User) CheckDoneLessons() {
	for _, level := range contents.Levels {
		for _, lesson := range level.Lessons {
			if u.GetLessonStudy(lesson) >= LS_STUDYING {
				finished := true
				for _, chunk := range lesson.Chunks {
					st := u.GetChunkStudy(chunk)
					if st != CS_DONE { finished = false }
					if st == CS_NOT_AVAILABLE { u.SetChunkStudy(chunk, CS_READING) }
				}
				if finished {
					u.SetLessonStudy(lesson, LS_DONE)
				} else {
					u.SetLessonStudy(lesson, LS_STUDYING)
				}
			}
		}
	}
}

// =============================================

func (u *User) GetLessonStatuses(level *contents.Level) []LessonWithStatus {
	ret := make([]LessonWithStatus, 0, len(level.Lessons))
	for _, lesson := range level.Lessons {
		ret = append(ret, LessonWithStatus{lesson, LS_NOT_AVAILABLE})
	}
	if u != nil {
		for id, lws := range ret {
			ret[id].Status = u.GetLessonStudy(lws.Lesson)
		}
	}
	return ret
}

func (u *User) GetChunkStatuses(lesson *contents.Lesson) []ChunkWithStatus {
	ret := make([]ChunkWithStatus, 0, len(lesson.Chunks))
	for _, chunk := range lesson.Chunks {
		ret = append(ret, ChunkWithStatus{chunk, CS_NOT_AVAILABLE, u})
	}
	if u != nil {
		for id, cws := range ret {
			ret[id].Status = u.GetChunkStudy(cws.Chunk)
		}
	}
	return ret
}

func (u *User) StartStudyingLesson(lesson *contents.Lesson) {
	if u.GetLessonStudy(lesson) < LS_STUDYING {
		u.SetLessonStudy(lesson, LS_STUDYING)
		for _, chunk := range lesson.Chunks {
			u.SetChunkStudy(chunk, CS_READING)
		}
		u.CheckNewAvailableLessons()
	}
}

func (u *User) SetChunkStatus(chunk *contents.Chunk, status int64) {
	before := u.GetChunkStudy(chunk)
	if before == CS_NOT_AVAILABLE || status == CS_NOT_AVAILABLE || status == before { return }
	if status == CS_REPEAT {
		//TODO : add SRS items
	} else if before == CS_REPEAT {
		//TODO : remove SRS items
	}
	u.SetChunkStudy(chunk, status)
	// update lesson status
	lesson_done := true
	for _, ch := range chunk.Lesson.Chunks {
		if u.GetChunkStudy(ch) != CS_DONE {
			lesson_done = false
			break
		}
	}
	if lesson_done {
		u.SetLessonStudy(chunk.Lesson, LS_DONE)
	} else {
		u.SetLessonStudy(chunk.Lesson, LS_STUDYING)
	}
	u.CheckNewAvailableLessons()
}

func (u *User) IsChunkSRSDone(chunk *contents.Chunk) bool {
	if u.GetChunkStudy(chunk) != CS_REPEAT { return false }
	//TODO !!!
	return true //means nothing, just for testing
}
