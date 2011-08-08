package study

import (
	"os"
	"time"
	"log"
	"github.com/kuroneko/gosqlite3"
)

import (
	"util"
	"main/config"
)

const dbTimeout = 30	//seconds

const (
	DBFALSE = "false"
	DBTRUE = "true"
)

// ========================================

type User struct {
	Username string
	db *sqlite3.Database
	dbLastUsed int64

	lessonStudy, chunkStudy map[string]int64
}

var users = make(map[string]*User)

// ========================================

func GetUser(username string) *User {
	if u, ok := users[username]; ok {
		return u
	}
	u := &User{Username: username}
	if util.Exists(u.dbFile()) {
		users[username] = u
		u.loadUp()
		return u
	}
	return nil
}

func CreateUser(username string) *User {
	u := GetUser(username)
	if u != nil { return u }
	u = &User{Username: username}
	users[username] = u
	u.loadUp()
	return u
}

func Startup() {
	log.Printf("Starting user DB closing goroutine...")
	go closeUnusedDBs()
	log.Printf("Checking admin users...")
	checkAdminUsers()
}

func checkAdminUsers() {
	for name, info := range config.Conf.AdminUsers {
		log.Printf("Admin user : %v", name)
		if info.ResetDBAtStartup {
			os.Remove((&User{Username: name}).dbFile())
		}
		user := CreateUser(name)
		user.SetAttr("admin", DBTRUE)
		user.SetAttr("password", util.StrSHA1(info.Password))
		user.SetAttr("email", info.Email)
		user.SetAttr("fullname", info.FullName)
	}
}

func closeUnusedDBs() {
	defer func() {		//will be called when program exits
		for _, user := range users {
			if user.db != nil {
				user.db.Close()
			}
		}
	}()
	t := time.NewTicker(20000000)
	for {
		<-t.C
		for _, user := range users {
			if user.dbLastUsed < time.Seconds() - dbTimeout && user.db != nil {
				user.db.Close()
				user.db = nil
			}
		}
	}
}

// =================== USER BASIC FUNCTIONS

func (u *User) loadUp() {
	u.checkTables()
	u.loadStudyStatus()
	u.checkFuriCfg()
}

func (u *User) checkTables() {
	u.DBQuery(`
		CREATE TABLE IF NOT EXISTS 'attributes' (
			id VARCHAR(42) PRIMARY KEY,
			value VARCHAR(100)
		)
	`)
	u.checkStudyTables()
	u.checkSRSTables()
}

func (u *User) SetAttr(name string, value string) {
	if e, _ := u.DBQueryFetchOne("SELECT id FROM 'attributes' WHERE id = ?", name); e != nil {
		u.DBQuery("UPDATE 'attributes' SET value = ? WHERE id = ?", value, name)
	} else {
		u.DBQuery("INSERT INTO 'attributes' VALUES(?, ?)", name, value)
	}
}

func (u *User) GetAttr(name string) string {
	if e, _ := u.DBQueryFetchOne("SELECT value FROM 'attributes' WHERE id = ?", name); e != nil {
		return e[0].(string)
	}
	return ""
}

func (u *User) CheckPass(pass string) bool {
	return u.GetAttr("password") == util.StrSHA1(pass)
}

// ===================== DB HELPERS (use these !)

func (u *User) dbFile() string {
	return config.Conf.UserDataFolder + "/" + u.Username + ".db3"
}

func (u *User) openDB() {
	u.dbLastUsed = time.Seconds()
	if u.db != nil { return }
	db, e := sqlite3.Open(u.dbFile())
	if e != nil { log.Panicf("Unable to open user DB file %v : %v", u.dbFile(), e) }
	u.db = db
}

func (u *User) DBQuery(sql string, v ...interface{}) os.Error {
	st := u.DBQuerySt(sql, v...)
	e := st.Step()
	st.Finalize()
	return e
}

func (u *User) DBQueryFetchOne(sql string, v ...interface{}) ([]interface{}, os.Error) {
	st := u.DBQuerySt(sql, v...)
	e := st.Step()
	if e == sqlite3.ROW {
		ret := st.Row()
		st.Finalize()
		return ret, nil
	}
	st.Finalize()
	return nil, e
}

func (u *User) DBQueryFetchAll(sql string, v ...interface{}) [][]interface{} {
	st := u.DBQuerySt(sql, v...)
	rows := make([][]interface{}, 0, 2)
	for st.Step() == sqlite3.ROW {
		rows = append(rows, st.Row())
	}
	return rows
}

func (u *User) DBQuerySt(sql string, v ...interface{}) *sqlite3.Statement {
	u.openDB()
	st, e := u.db.Prepare(sql, v...)
	if e != nil { log.Panicf("SQL error : %v ; query : %v", e, sql) }
	return st
}
