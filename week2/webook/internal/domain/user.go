package domain

import "time"

type User struct {
	Id       int64
	Email    string
	Password string
	//扩展 用户名、生日、简介
	Name     string
	Birthday string
	Intro    string

	Phone string
	Ctime time.Time
}
