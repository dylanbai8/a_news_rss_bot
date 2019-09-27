package model

type User struct {
	ID     int64    `gorm:"primary_key"`
	Source []Source `gorm:"many2many:subscribes;"`
	EditTime
}

func FindOrInitUser(userID int64) *User {
	db := getConnect()
	defer db.Close()
	var user User
	db.Where(User{ID: userID}).FirstOrCreate(&user)
	return &user
}
