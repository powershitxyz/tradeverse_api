package preauth

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,min=5"`
	Password string `json:"password" binding:"required,min=6"`
	Ref      string `json:"ref"`
}

type ConfirmRegisterRequest struct {
	Email string `json:"email" binding:"required,min=5"`
	Code  string `json:"code" binding:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,min=5"`
	Password string `json:"password" binding:"required,min=6"`
}

type DeveloperRegisterRequest struct {
	Email    string `json:"email" binding:"required,min=5"`
	Password string `json:"password" binding:"required,min=6"`
}

type DeveloperVerifyRequest struct {
	Email string `json:"email" binding:"required,min=5"`
	Code  string `json:"code" binding:"required"`
}
