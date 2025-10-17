package request

type Page struct {
	Pn int `json:"pn" binding:"min=1"`
	Ps int `json:"ps" binding:"min=1"`
}
