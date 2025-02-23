package request

type Register struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Login struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Withdraw struct {
	Order string `json:"order"`
	Sum   int    `json:"sum"`
}
