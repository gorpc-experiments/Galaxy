package domain

type RegisterRequest struct {
	//Component any
	Address    string
	Components []string

	Version int
	Host    string
}

type RegisterResponse struct {
	Success bool
}