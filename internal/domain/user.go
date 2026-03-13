package domain

// User 定義使用者模型
// PasswordHash 不會被 JSON 回傳
// Username 必須唯一
type User struct {
	ID           int    `json:"id"`
	Username     string `json:"username" gorm:"uniqueIndex;size:64"`
	PasswordHash string `json:"-"`
}

// UserRepository 定義使用者資料操作介面
type UserRepository interface {
	GetByUsername(username string) (*User, error)
	Create(u User) (User, error)
}
