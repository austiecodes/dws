package types

type RegisterReq struct {
	Username string `json:"user_name"`
	UnixName string `json:"unix_name"`
	Password string `json:"password"`
}
