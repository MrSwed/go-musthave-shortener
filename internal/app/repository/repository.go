package repository

type Repository struct {
	MemStorage
	FileStorage
}

func NewRepository(f string) Repository {
	return Repository{
		MemStorage:  NewMemRepository(),
		FileStorage: NewFileStorage(f),
	}
}
