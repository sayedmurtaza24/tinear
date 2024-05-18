package user

type User struct {
	ID          string
	DisplayName string
	Email       string
	IsMe        bool
	OrgName     string
}

func (u *User) SortWeight() int {
	if u.IsMe {
		return 1
	}
	return 0
}
