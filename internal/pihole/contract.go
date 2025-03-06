package pihole

type LoginRequest struct {
	Password string `json:"password"`
}

type LoginResponse struct {
	Session Session `json:"session"`
	Took    float64 `json:"took"`
}

type RecordsResponse struct {
	Config RecordsConfig `json:"config"`
	Took   float64       `json:"took"`
}
