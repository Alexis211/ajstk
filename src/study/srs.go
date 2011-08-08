package study

import (
	"contents"
)

const (				// Setting : lesson furigana display conditions
	LF_ALWAYS = iota
	LF_NOT_KNOWN
	LF_NOT_STUDYING
	LF_NEVER
)
const (				// Setting : any furigana display conditions
	AF_ALWAYS = iota
	AF_NEVER
)

func (u *User) checkFuriCfg() {
	if u.GetAttr("lessfuri") == "" { u.SetLessFuri(LF_NOT_STUDYING) }
	if u.GetAttr("anyfuri") == "" { u.SetAnyFuri(AF_ALWAYS) }
}

func (u *User) SetLessFuri(status int) {
	u.SetAttr("lessfuri",
		[]string{"always", "not_known", "not_studying", "never"}[status])
}
func (u *User) GetLessFuri() int {
	if u == nil { return LF_ALWAYS }
	return map[string]int{
		"always": LF_ALWAYS, "not_known": LF_NOT_KNOWN,
		"not_studying": LF_NOT_STUDYING, "never": LF_NEVER,
	}[u.GetAttr("lessfuri")]
}
func (u *User) SetAnyFuri(status int) {
	u.SetAttr("anyfuri", []string{"always", "never"}[status])
}
func (u *User) GetAnyFuri() int {
	if u == nil { return AF_ALWAYS }
	return map[string]int{
		"always": AF_ALWAYS, "never": AF_NEVER,
	}[u.GetAttr("anyfuri")]
}

func (u *User) LFAlways() bool { return u.GetLessFuri() == LF_ALWAYS }
func (u *User) LFNotKnown() bool { return u.GetLessFuri() == LF_NOT_KNOWN }
func (u *User) LFNotStudying() bool { return u.GetLessFuri() == LF_NOT_STUDYING }
func (u *User) LFNever() bool { return u.GetLessFuri() == LF_NEVER }
func (u *User) AFAlways() bool { return u.GetAnyFuri() == AF_ALWAYS }
func (u *User) AFNever() bool { return u.GetAnyFuri() == AF_NEVER }

// =================================================

const doneMinBox = 5

const (			//srs item possible statuses
	SS_NOT_STUDYING = iota
	SS_LEARNING
	SS_REPEATING
	SS_DONE
)

type SRSItemWithStatus struct {
	Item *contents.SRSItem
	Status int64
	Box int64
}
func (i SRSItemWithStatus) Studying() bool { return i.Status >= SS_LEARNING }
func (i SRSItemWithStatus) Known() bool { return i.Status == SS_DONE }

func (u *User) checkSRSTables() {
	u.DBQuery(`
		CREATE TABLE IF NOT EXISTS 'srs_study' (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			chunk_id VARCHAR(256),
			groupe VARCHAR(256), subgroup VARCHAR(256),
			meaning VARCHAR(256), japanese VARCHAR(256),
				reading VARCHAR(256), comment VARCHAR(256),
			box INTEGER, next_review VARCHAR(42),
			noteToSelf VARCHAR(256),
			UNIQUE (groupe, subgroup, japanese),
			UNIQUE (groupe, subgroup, meaning)
		)`)
	u.DBQuery(`CREATE INDEX IF NOT EXISTS 'srs_chunk_index'
		ON 'srs_study' (chunk_id)`)
	u.DBQuery(`CREATE INDEX IF NOT EXISTS 'srs_group_index'
		ON 'srs_study' (groupe)`)
	u.DBQuery(`CREATE INDEX IF NOT EXISTS 'srs_subgroup_index'
		ON 'srs_study' (groupe, subgroup)`)
	u.DBQuery(`CREATE INDEX IF NOT EXISTS 'srs_subgroup_review_date_index'
		ON 'srs_study' (next_review)`)
	u.DBQuery(`CREATE INDEX IF NOT EXISTS 'srs_box_index'
		ON 'srs_study' (box)`)
}

func (u *User) GetSRSItemStatuses(chunk *contents.Chunk) []SRSItemWithStatus {
	ret := make([]SRSItemWithStatus, 0, len(chunk.SRSItems))
	for _, srsItem := range chunk.SRSItems {
		ret = append(ret, SRSItemWithStatus{srsItem, SS_NOT_STUDYING, 0})
	}
	if u != nil {
		std := u.GetChunkStudy(chunk)
		if std == CS_READING {
			for id := range ret {
				ret[id].Status = SS_LEARNING
			}
		} else if std == CS_REPEAT {
			for _, dbs := range u.DBQueryFetchAll(
				`SELECT japanese, box FROM 'srs_study' WHERE chunk_id = ?`, chunk.FullId()) {
				for id, _ := range ret {
					if ret[id].Item.Japanese == dbs[0].(string) {
						ret[id].Box = dbs[1].(int64)
						if ret[id].Box >= doneMinBox {
							ret[id].Status = SS_DONE
						} else if ret[id].Box == 0 {
							ret[id].Status = SS_LEARNING
						} else {
							ret[id].Status = SS_REPEATING
						}
					}
				}
			}
		} else if std == CS_DONE {
			for id := range ret {
				ret[id].Status = SS_DONE
				ret[id].Box = 10
			}
		}
	}
	return ret
}

func (u *User) IsChunkSRSDone(chunk *contents.Chunk) bool {
	if u.GetChunkStudy(chunk) != CS_REPEAT { return false }

	e, _ := u.DBQueryFetchOne(
		`SELECT chunk_id FROM 'srs_study' WHERE chunk_id = ? AND box < ?`,
		chunk.FullId(), doneMinBox)
	return e == nil
}

// ===================================

func (u *User) addSRSItems(chunk *contents.Chunk) {
	for _, item := range chunk.SRSItems {
		//Delete already existing items
		u.DBQuery(
			"DELETE FROM 'srs_study' WHERE groupe = ? AND subgroup = ? AND japanese = ?",
			item.Group, item.Subgroup, item.Japanese)
		u.DBQuery(
			"DELETE FROM 'srs_study' WHERE groupe = ? AND subgroup = ? AND meaning = ?",
			item.Group, item.Subgroup, item.Meaning)
		//Add item
		u.DBQuery(
			`INSERT INTO 'srs_study' (
				chunk_id, groupe, subgroup,
				meaning, japanese, reading, comment,
				box, next_review, noteToSelf)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, date('now', '+1 day'), null)`,
			chunk.FullId(), item.Group, item.Subgroup,
			item.Meaning, item.Japanese, item.Reading, item.Comment,
			0)
	}
}

func (u *User) removeSRSItems(chunk *contents.Chunk) {
	u.DBQuery("DELETE FROM 'srs_study' WHERE chunk_id = ?", chunk.FullId())
}

