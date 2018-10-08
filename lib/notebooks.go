package qniblib


type Notebook struct {
	ID 		string
	Url 	string
	Token 	string
}

func NewNotebook(id, url, token string) Notebook {
	return Notebook{ID: id, Url: url, Token: token}
}
