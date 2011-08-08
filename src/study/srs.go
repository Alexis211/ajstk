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

const (			//srs item possible statuses
	SS_NOT_STUDYING = iota
	SS_LEARNING
	SS_REPEATING
	SS_DONE
)

type SRSItemWithStatus struct {
	Item *contents.SRSItem
	Status int64
	Box int
}
func (i SRSItemWithStatus) Studying() bool { return i.Status >= SS_LEARNING }
func (i SRSItemWithStatus) Known() bool { return i.Status == SS_DONE }

func (u *User) checkSRSTables() {
	//TODO !!!
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
			//TODO : take into account user progress. THE FOLLOWING 4 LINES ARE BAD
			for id := range ret {
				ret[id].Status = SS_DONE
				ret[id].Box = 10
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
	 //TODO !!!
	return true //means nothing, just for testing
}

// ===================================

func (u *User) SRSAddItems(chunk *contents.Chunk) {
	// TODO
}

func (u *User) SRSRemoveItems(chunk *contents.Chunk) {
	// TODO
}

