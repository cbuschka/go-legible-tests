package replication

import (
	"github.io/cbuschka/go-legible-tests/domain/product"
)

type metricsSender interface {
	ReportSuccess(count int)
	ReportFailure(err error)
}

type client interface {
	Fetch() ([]product.Product, error)
}

type repository interface {
	FindByIDs(productIDs []product.ID) (map[product.ID]product.Product, error)
	Save(products []product.Product) error
}

type Service struct {
	client  client
	repo    repository
	metrics metricsSender
}

func NewService(client client, repo repository, metrics metricsSender) *Service {
	return &Service{client, repo, metrics}
}

func (s *Service) Replicate() error {
	count, err := s.replicate()
	if err != nil {
		s.metrics.ReportFailure(err)
		return err
	}

	s.metrics.ReportSuccess(count)
	return nil
}

func (s *Service) replicate() (int, error) {
	products, err := s.client.Fetch()
	if err != nil {
		return 0, err
	}

	existingProductsByID, err := s.findExistingProducts(products)
	if err != nil {
		return 0, err
	}

	changedProducts := make([]product.Product, 0, len(existingProductsByID))
	for _, newProduct := range products {
		changedProduct, found := existingProductsByID[newProduct.ID]
		if !found {
			changedProduct = product.Product{Name: newProduct.Name}
		} else {
			changedProduct.Name = newProduct.Name
		}
		changedProducts = append(changedProducts, changedProduct)
	}

	err = s.repo.Save(changedProducts)
	if err != nil {
		return 0, err
	}

	return len(changedProducts), nil
}

func (s *Service) findExistingProducts(newProducts []product.Product) (map[product.ID]product.Product, error) {
	prouctIds := collect[product.Product, product.ID](newProducts, func(p product.Product) product.ID {
		return p.ID
	})
	return s.repo.FindByIDs(prouctIds)
}

func collect[A any, B any](as []A, f func(element A) B) []B {
	bs := make([]B, len(as))
	for i, a := range as {
		b := f(a)
		bs[i] = b
	}
	return bs
}
