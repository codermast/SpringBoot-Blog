package models

type User struct {
	ID       string `gorm:"primaryKey;autoIncrement" json:"id"`
	Username string `gorm:"size:255" json:"username"`
	Password string `gorm:"size:255" json:"password"`
	Nickname string `gorm:"size:255" json:"nickname"`
	Mail     string `gorm:"size:255" json:"mail"`
	Tel      string `gorm:"size:255" json:"tel"`
	Admin    bool   `gorm:"default:false" json:"admin"`
}
