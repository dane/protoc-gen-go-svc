package internal

type RegisterService struct {
	PackageName string
	Services    []*Service
	Private     *Service
}
