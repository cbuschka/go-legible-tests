package replication

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.io/cbuschka/go-legible-tests/domain/product"
	"testing"
)

func TestService(t *testing.T) {
	tests := []test{
		{
			name: "fails if client returns error",
			given: []givenSpec{
				givenClientFails(),
			},
			expect: []expectSpec{
				expectFailureReported(ErrClientRequestFailed),
			},
			verify: verifyErrorReturned(ErrClientRequestFailed),
		},
		{
			name: "fails if client returns no products",
			given: []givenSpec{
				givenClientReturnsNoProducts(),
			},
			expect: []expectSpec{
				expectFailureReported(ErrNoProducts),
			},
			verify: verifyErrorReturned(ErrNoProducts),
		},
		{
			name: "creates products if not existing yet",
			given: []givenSpec{
				givenClientReturnsProducts([]product.Product{{1, "p1"}}),
				givenRepoReturnsNoProducts(),
			},
			expect: []expectSpec{
				expectSuccessReported(1),
				expectProductsSaved([]product.Product{{1, "p1"}}),
			},
			verify: verifyNoErrorReturned(),
		},
	}

	for _, test := range tests {
		test.run(t)
	}
}

type test struct {
	name   string
	given  []givenSpec
	expect []expectSpec
	verify verifySpec
}

func (test *test) run(t *testing.T) {
	t.Run(test.name, func(t *testing.T) {
		m := mocks{
			repo:    &mockrepository{},
			metrics: &mockmetricsSender{},
			client:  &mockclient{},
		}

		for _, given := range test.given {
			given(t, &m)
		}

		for _, expect := range test.expect {
			expect(t, &m)
		}

		service := NewService(m.client, m.repo, m.metrics)
		err := service.Replicate()

		test.verify(t, err)
	})
}

type mocks struct {
	repo    *mockrepository
	metrics *mockmetricsSender
	client  *mockclient
}

type givenSpec func(t *testing.T, m *mocks)
type expectSpec func(t *testing.T, m *mocks)
type verifySpec func(t *testing.T, err error)

func givenClientFails() givenSpec {
	return func(t *testing.T, m *mocks) {
		m.client.EXPECT().Fetch().Return(nil, ErrClientRequestFailed).Once()
	}
}

func givenClientReturnsNoProducts() givenSpec {
	return givenClientReturnsProducts([]product.Product{})
}
func givenClientReturnsProducts(products []product.Product) givenSpec {
	return func(t *testing.T, m *mocks) {
		m.client.EXPECT().Fetch().Return(products, nil).Once()
	}
}

func givenRepoReturnsNoProducts() givenSpec {
	return func(t *testing.T, m *mocks) {
		m.repo.EXPECT().FindByIDs(mock.Anything).Return(map[product.ID]product.Product{}, nil).Once()
	}
}

func expectFailureReported(err error) expectSpec {
	return func(t *testing.T, m *mocks) {
		m.metrics.EXPECT().ReportFailure(err).Once()
	}
}

func expectSuccessReported(count int) expectSpec {
	return func(t *testing.T, m *mocks) {
		m.metrics.EXPECT().ReportSuccess(count).Once()
	}
}

func expectProductsSaved(expectedProducts []product.Product) expectSpec {
	return func(t *testing.T, m *mocks) {
		m.repo.EXPECT().Save(expectedProducts).Return(nil).Once()
	}
}

func verifyErrorReturned(expectedErr error) verifySpec {
	return func(t *testing.T, err error) {
		assert.Equal(t, expectedErr, err)
	}
}

func verifyNoErrorReturned() verifySpec {
	return func(t *testing.T, err error) {
		assert.Nil(t, err)
	}
}
