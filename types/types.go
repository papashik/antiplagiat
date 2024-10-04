package types

import (
	"database/sql"
	"time"
)

type Delivery struct {
	Id         int
	User_id    int
	Course_id  int
	Lab        int
	Variant    int
	Date       time.Time
	Solution   sql.NullString
	Error      sql.NullString
	Language   string
	Ant_mark   int
	Ant_review sql.NullString
}

type User struct {
	Id      int
	Name    string
	Surname string
}

type Course struct {
	Id        int
	Name      string
	Long_name string
}
