package addresses

type service struct {
	*repo
}

func NewService(repo *repo) *service {
	return &service{repo}
}

func (this *service) Create(model *Model) error {
	this.repo.create(model)
	return nil
}
