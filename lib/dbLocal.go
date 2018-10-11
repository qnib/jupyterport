package qniblib

type LocalDB struct {
	data map[string]Notebook
}

func NewLocalDB() LocalDB {
	return LocalDB{data: make(map[string]Notebook)}
}

func (db *LocalDB) Init() (err error) {
	return
}
func (db *LocalDB) ListNotebooks(user string) (nbs map[string]Notebook, err error) {
	return
}


func (db *LocalDB) AddNotebook(notebook Notebook) (err error) {
	return
}

func (db *LocalDB) RemoveNotebook(notebook Notebook) (err error) {
	return
}
