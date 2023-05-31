package domain

type LookUpRequest struct {
	ServiceMethod string
}

type LookUpResponse struct {
	Address string
}