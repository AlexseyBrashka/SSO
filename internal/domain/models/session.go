package models

type Session struct {
	Login      string
	Permission map[int64](bool)
	TimeToCash int64
}
