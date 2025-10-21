package entity

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type UserInfoToken struct {
	ID   int64    `json:"id"`
	Role UserRole `json:"role"`
}

func (u *UserInfoToken) IsEqualRole(role UserRole) bool {
	return u.Role == role
}
