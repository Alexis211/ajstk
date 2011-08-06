package main

import (
	"fmt"
	"github.com/kuroneko/gosqlite3"
)

var FOO *sqlite3.Table

func main() {
	db, e := sqlite3.Open("test.db")
	if e != nil { panic(e.String()) }
	defer db.Close()

	db.Execute("DROP INDEX IF EXISTS 'test'")
	db.Execute("DROP TABLE IF EXISTS 'foo'")
	db.Execute("CREATE TABLE IF NOT EXISTS 'foo' (id INTEGER PRIMARY KEY AUTOINCREMENT, name VARCHAR(20), owner INTEGER)")
	db.Execute("CREATE UNIQUE INDEX IF NOT EXISTS 'test' on 'foo' (name)")
	db.Execute("CREATE INDEX IF NOT EXISTS 'test2' on 'foo' (owner)")

	st, e := db.Prepare("INSERT INTO foo VALUES (?, ?, ?)", 42, "Hi (lol)", 22)
	if e != nil { panic(e.String() + ":" + st.SQLSource()) }
	e = st.Step()
	st.Finalize()
	if e != nil { fmt.Printf("1: %v\n", e) }
	fmt.Printf("Last row id: %v\n", db.LastInsertRowID())

	st, e = db.Prepare("INSERT INTO foo VALUES (?, ?, ?)", 12, "nyoron", 21)
	if e != nil { panic(e.String() + ":" + st.SQLSource()) }
	e = st.Step()
	st.Finalize()
	if e != nil { fmt.Printf("2: %v\n", e) }
	fmt.Printf("Last row id: %v\n", db.LastInsertRowID())

	st, e = db.Prepare("INSERT INTO foo VALUES (?, ?, ?)", 42, "Hi (lol)", 22)
	if e != nil { panic(e.String() + ":" + st.SQLSource()) }
	e = st.Step()
	st.Finalize()
	if e != nil { fmt.Printf("3: %v\n", e) }
	fmt.Printf("Last row id: %v\n", db.LastInsertRowID())

	db.Execute("SELECT * FROM foo", func(st *sqlite3.Statement, values ...interface{}) {
		fmt.Printf("%#v\n", values)
		if v, ok := values[1].(string); ok {
			fmt.Printf("%v\n", v)
		} else {
			fmt.Printf("kk\n")
		}
	})

	st, e = db.Prepare("SELECT * FROM foo")
	if e != nil { panic(e.String() + ":" + st.SQLSource()) }
	for st.Step() == sqlite3.ROW {
		row := st.Row()
		fmt.Printf("%#v\n", row)
	}
	st.Finalize()
}
