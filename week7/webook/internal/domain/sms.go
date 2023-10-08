package domain

import "time"

type Sms struct {
	Id    int64
	TplId string
	Phone string
	Msg   string
	Ctime time.Time
}
