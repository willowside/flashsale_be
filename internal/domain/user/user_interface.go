package user

type UserInterface interface {
	GetUserByID(userID int64) (*User, error)
}
