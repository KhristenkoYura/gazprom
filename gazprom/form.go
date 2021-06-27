package gazprom

type User struct {
	ID int `json:"id"`
	Login string `json:"login" validate:"required"`
	Password string `json:"password" validate:"required"`
	Role string `json:"role"`
}
