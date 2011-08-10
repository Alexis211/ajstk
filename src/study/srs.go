package study

import (
	"fmt"
	"github.com/kuroneko/gosqlite3"
)

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
	// TODO : add user comment
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
			subgroup_active INTEGER,
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
		isActive := 0
		if u.SRSIsSubgroupActivated(item.Group, item.Subgroup) { isActive = 1 }
		u.DBQuery(
			`INSERT INTO 'srs_study' (
				chunk_id, groupe, subgroup,
				meaning, japanese, reading, comment,
				box, next_review, subgroup_active)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, date('now', '+1 day'), ?)`,
			chunk.FullId(), item.Group, item.Subgroup,
			item.Meaning, item.Japanese, item.Reading, item.Comment,
			0, isActive)
	}
}

func (u *User) removeSRSItems(chunk *contents.Chunk) {
	u.DBQuery("DELETE FROM 'srs_study' WHERE chunk_id = ?", chunk.FullId())
}

// =========================================

func (u *User) SRSIsSubgroupActivated(group, subgroup string) bool {
	return u.GetAttr("srs_" + group + "__" + subgroup) != "deactivated"
}
func (u *User) SRSActivateSubgroup(group, subgroup string) {
	u.SetAttr("srs_" + group + "__" + subgroup, "")
	u.DBQuery(`UPDATE 'srs_study' SET subgroup_active = 1
		WHERE groupe = ? AND subgroup = ?`, group, subgroup)
}
func (u *User) SRSDeactivateSubgroup(group, subgroup string) {
	u.SetAttr("srs_" + group + "__" + subgroup, "deactivated")
	u.DBQuery(`UPDATE 'srs_study' SET subgroup_active = 0
		WHERE groupe = ? AND subgroup = ?`, group, subgroup)
}



type SRSGroupStat struct {
	Name, Path string
	BoxCardCount map[int64]int64
	SubGroups map[string]*SRSGroupStat
	ToStudy, ToReview int64
	Disabled bool
}
func (g SRSGroupStat) HasStudy() bool { return g.ToStudy != 0 }
func (g SRSGroupStat) HasReview() bool { return g.ToReview != 0 }
func (g SRSGroupStat) BarHTML() string {
	if len(g.BoxCardCount) == 0 || g.Disabled {
		return `<div class="bar"><div class="di" style="width: 200px"></div></div>`
	}
	var barWidth, totalItems, maxBox int64 = 200, 0, 0
	ret := `<div class="bar">`
	for i, k := range g.BoxCardCount {
		totalItems += k
		if i > maxBox { maxBox = i }
	}
	for i := int64(0); i <= maxBox; i++ {
		if count, ok := g.BoxCardCount[i]; ok {
			style := "rknown"
			if i < doneMinBox { style = fmt.Sprintf("r%v", i) }
			ret += fmt.Sprintf(`<div class="%v" style="width: %vpx"></div>`, style, count * barWidth / totalItems)
		}
	}
	ret += `</div>`
	return ret
}
func (u *User) GetSRSStats() map[string]*SRSGroupStat {
	subgroupBoxes := u.DBQueryFetchAll(
		`SELECT COUNT(id) AS count, groupe, subgroup, box FROM 'srs_study'
			GROUP BY groupe, subgroup, box`)
	ret := make(map[string]*SRSGroupStat)
	for _, sgb := range subgroupBoxes {
		count := sgb[0].(int64)
		group := sgb[1].(string)
		subgroup := sgb[2].(string)
		box := sgb[3].(int64)
		if _, ok := ret[group]; !ok {
			ret[group] = &SRSGroupStat{
				Name: group, Path: group,
				BoxCardCount: map[int64]int64{},
				SubGroups: map[string]*SRSGroupStat{},
				ToStudy: 0, ToReview: 0, Disabled: false,
			}
		}
		if _, ok := ret[group].SubGroups[subgroup]; !ok {
			ret[group].SubGroups[subgroup] = &SRSGroupStat{
				Name: subgroup, Path: group + "/" + subgroup,
				BoxCardCount: map[int64]int64{},
				SubGroups: nil, ToStudy: 0, ToReview: 0,
				Disabled: !u.SRSIsSubgroupActivated(group, subgroup),
			}
		}
		if !ret[group].SubGroups[subgroup].Disabled {
			if c, ok := ret[group].BoxCardCount[box]; ok {
				ret[group].BoxCardCount[box] = c + count
			} else {
				ret[group].BoxCardCount[box] = count
			}
			if c, ok := ret[group].SubGroups[subgroup].BoxCardCount[box]; ok {
				ret[group].SubGroups[subgroup].BoxCardCount[box] = c + count
			} else {
				ret[group].SubGroups[subgroup].BoxCardCount[box] = count
			}
		}
	}
	toStudyl := u.DBQueryFetchAll(
		`SELECT COUNT(id) AS count, groupe, subgroup FROM 'srs_study'
			WHERE next_review > DATE('now') AND box == 0 AND subgroup_active = 1
			GROUP BY groupe, subgroup`)
	for _, e := range toStudyl {
		count := e[0].(int64)
		group := e[1].(string)
		subgroup := e[2].(string)
		ret[group].ToStudy += count
		ret[group].SubGroups[subgroup].ToStudy += count
	}
	toReviewl := u.DBQueryFetchAll(
		`SELECT COUNT(id) AS count, groupe, subgroup FROM 'srs_study'
			WHERE next_review <= DATE('now') AND subgroup_active = 1
			GROUP BY groupe, subgroup`)
	for _, e := range toReviewl {
		count := e[0].(int64)
		group := e[1].(string)
		subgroup := e[2].(string)
		ret[group].ToReview += count
		ret[group].SubGroups[subgroup].ToReview += count
	}
	return ret
}


type SRSStudyItem struct {
	Group, Subgroup string
	Meaning, Japanese, Reading, Comment string
	Id, Box int64
	// TODO : add user comment
}
func (u *User) getSRSStudyQuery(sql string, v ...interface{}) []*SRSStudyItem {
	st := u.DBQuerySt(sql, v...)
	ret := make([]*SRSStudyItem, 0, 5)
	for st.Step() == sqlite3.ROW {
		r := st.Row()
		item := &SRSStudyItem{
			Group: r[0].(string), Subgroup: r[1].(string),
			Meaning: r[2].(string), Japanese: r[3].(string),
			Reading: r[4].(string), Comment: r[5].(string),
			Id: r[6].(int64), Box: r[7].(int64),
		}
		ret = append(ret, item)
	}
	return ret
}
func (u *User) GetSRSReviewItems(group, subgroup string) []*SRSStudyItem {
	if subgroup == "" {
		return u.getSRSStudyQuery(
			`SELECT groupe, subgroup, meaning, japanese, reading, comment, id, box
			FROM 'srs_study' WHERE next_review <= DATE('now') AND groupe = ?`,
			group)
	}
	return u.getSRSStudyQuery(
		`SELECT groupe, subgroup, meaning, japanese, reading, comment, id, box
		FROM 'srs_study' WHERE next_review <= DATE('now')
		AND groupe = ? AND subgroup = ?`,
		group, subgroup)
}
func (u *User) GetSRSTomorrowItems(group, subgroup string) []*SRSStudyItem {
	if subgroup == "" {
		return u.getSRSStudyQuery(
			`SELECT groupe, subgroup, meaning, japanese, reading, comment, id, box
			FROM 'srs_study' WHERE next_review > DATE('now') AND box = 0
			AND groupe = ?`,
			group)
	}
	return u.getSRSStudyQuery(
		`SELECT groupe, subgroup, meaning, japanese, reading, comment, id, box
		FROM 'srs_study' WHERE next_review > DATE('now') AND box = 0
		AND groupe = ? AND subgroup = ?`,
		group, subgroup)
}

func (u *User) GetSRSChunkItemsDrill(chunk *contents.Chunk) []*SRSStudyItem {
	ret := make([]*SRSStudyItem, len(chunk.SRSItems))
	for id, item := range chunk.SRSItems {
		ret[id] = &SRSStudyItem{
			Group: item.Group, Subgroup: item.Subgroup,
			Japanese: item.Japanese, Meaning: item.Meaning,
			Reading: item.Reading, Comment: item.Comment,
			Box: 0, Id: 0,
		}
	}
	return ret
}
func (u *User) GetSRSLessonItemsDrill(lesson *contents.Lesson) []*SRSStudyItem {
	ret := make([]*SRSStudyItem, 0)
	for _, ch := range lesson.Chunks {
		ret = append(ret, u.GetSRSChunkItemsDrill(ch)...)
	}
	return ret
}

// ===============================

var boxTimeIntervals = []int{1, 2, 4, 7, 11, 18, 30, 50, 80, 130, 210, 340, 550, 1000}

func (u *User) UpdateSRSItemStatuses(success, fail []int64) {
	boxMap := make(map[int64]int64)
	st := u.DBQuerySt(`SELECT id, box FROM 'srs_study'
		WHERE next_review <= DATE('now')`)
	for st.Step() == sqlite3.ROW {
		r := st.Row()
		boxMap[r[0].(int64)] = r[1].(int64)
	}
	for _, id := range success {
		if box, ok := boxMap[id]; ok {
			u.DBQuery(fmt.Sprintf(
				`UPDATE 'srs_study' SET box = ?, next_review = DATE('now', '+%v days')
				WHERE id = ?`, boxTimeIntervals[box]), box + 1, id)
		}
	}
	for _, id := range fail {
		if _, ok := boxMap[id]; ok {
			u.DBQuery(
				`UPDATE 'srs_study' SET box = 0, next_review = DATE('now', '+1 days')
				WHERE id = ?`, id)
		}
	}
}
