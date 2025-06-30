package replication

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.io/cbuschka/go-legible-tests/domain/product"
	"testing"
)

func TestService(t *testing.T) {
	tests := map[string]func(t *test){
		"fails if client returns error": func(t *test) {
			t.given(
				clientFails(),
			).expect(
				failureReported(ErrClientRequestFailed),
			).then(
				errorReturned(ErrClientRequestFailed),
			)
		},
		"fails if client returns no products": func(t *test) {
			t.given(
				clientReturnsNoProducts(),
			).expect(
				failureReported(ErrNoProducts),
			).then(
				errorReturned(ErrNoProducts),
			)
		},
		"creates products if not existing yet": func(t *test) {
			products := []product.Product{{1, "p1"}}
			t.given(
				clientReturnsProducts(products),
				repoReturnsNoProducts(),
			).expect(
				successReported(1),
				productsSaved(products),
			).then(
				noErrorReturned(),
			)
		},
	}

	for name, testSpec := range tests {
		t.Run(name, func(t *testing.T) {
			test := &test{}
			testSpec(test)
			test.run(t)
		})
	}
}

type test struct {
	givenSpecs  []givenSpec
	expectSpecs []expectSpec
	verifySpec  verifySpec
}

func (tc *test) given(givenSpecs ...givenSpec) *test {
	tc.givenSpecs = append(tc.givenSpecs, givenSpecs...)
	return tc
}

func (tc *test) expect(expectSpecs ...expectSpec) *test {
	tc.expectSpecs = append(tc.expectSpecs, expectSpecs...)
	return tc
}

func (tc *test) then(verifySpec verifySpec) *test {
	tc.verifySpec = verifySpec
	return tc
}

func (tc *test) run(t *testing.T) {
	m := mocks{
		repo:    &mockrepository{},
		metrics: &mockmetricsSender{},
		client:  &mockclient{},
	}

	for _, givenSpec := range tc.givenSpecs {
		givenSpec(t, &m)
	}

	for _, expect := range tc.expectSpecs {
		expect(t, &m)
	}

	service := NewService(m.client, m.repo, m.metrics)
	err := service.Replicate()

	tc.verifySpec(t, err)
}

type mocks struct {
	repo    *mockrepository
	metrics *mockmetricsSender
	client  *mockclient
}

type givenSpec func(t *testing.T, m *mocks)
type expectSpec func(t *testing.T, m *mocks)
type verifySpec func(t *testing.T, err error)

func clientFails() givenSpec {
	return func(t *testing.T, m *mocks) {
		m.client.EXPECT().Fetch().Return(nil, ErrClientRequestFailed).Once()
	}
}

func clientReturnsNoProducts() givenSpec {
	return clientReturnsProducts([]product.Product{})
}
func clientReturnsProducts(products []product.Product) givenSpec {
	return func(t *testing.T, m *mocks) {
		m.client.EXPECT().Fetch().Return(products, nil).Once()
	}
}

func repoReturnsNoProducts() givenSpec {
	return func(t *testing.T, m *mocks) {
		m.repo.EXPECT().FindByIDs(mock.Anything).Return(map[product.ID]product.Product{}, nil).Once()
	}
}

func failureReported(err error) expectSpec {
	return func(t *testing.T, m *mocks) {
		m.metrics.EXPECT().ReportFailure(err).Once()
	}
}

func successReported(count int) expectSpec {
	return func(t *testing.T, m *mocks) {
		m.metrics.EXPECT().ReportSuccess(count).Once()
	}
}

func productsSaved(expectedProducts []product.Product) expectSpec {
	return func(t *testing.T, m *mocks) {
		m.repo.EXPECT().Save(expectedProducts).Return(nil).Once()
	}
}

func errorReturned(expectedErr error) verifySpec {
	return func(t *testing.T, err error) {
		assert.Equal(t, expectedErr, err)
	}
}

func noErrorReturned() verifySpec {
	return func(t *testing.T, err error) {
		assert.Nil(t, err)
	}
}
