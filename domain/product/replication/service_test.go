package replication

import (
	"github.com/stretchr/testify/assert"
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
	return func(t *testing.T, m *mocks) {
		m.client.EXPECT().Fetch().Return([]product.Product{}, nil).Once()
	}
}

func expectFailureReported(err error) expectSpec {
	return func(t *testing.T, m *mocks) {
		m.metrics.EXPECT().ReportFailure(err).Once()
	}
}

func verifyErrorReturned(expectedErr error) verifySpec {
	return func(t *testing.T, err error) {
		assert.Equal(t, expectedErr, err)
	}
}
