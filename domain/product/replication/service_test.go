package replication

import (
	"github.com/stretchr/testify/assert"
	"github.io/cbuschka/go-legible-tests/domain/product"
	"testing"
)

type testMocks struct {
	repo    *mockrepository
	metrics *mockmetricsSender
	client  *mockclient
}

type givenSpec func(t *testing.T, m *testMocks)
type expectSpec func(t *testing.T, m *testMocks)
type verifySpec func(t *testing.T, err error)

func givenClientFails() givenSpec {
	return func(t *testing.T, m *testMocks) {
		m.client.EXPECT().Fetch().Return(nil, ErrClientRequestFailed).Once()
	}
}

func givenClientReturnsNoProducts() givenSpec {
	return func(t *testing.T, m *testMocks) {
		m.client.EXPECT().Fetch().Return([]product.Product{}, nil).Once()
	}
}

func expectFailureReported(err error) expectSpec {
	return func(t *testing.T, m *testMocks) {
		m.metrics.EXPECT().ReportFailure(err).Once()
	}
}

func verifyErrorReturned(expectedErr error) verifySpec {
	return func(t *testing.T, err error) {
		assert.Equal(t, expectedErr, err)
	}
}

func TestService(t *testing.T) {
	tests := []struct {
		name   string
		given  []givenSpec
		expect []expectSpec
		verify verifySpec
	}{
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
		t.Run(test.name, func(t *testing.T) {
			t.Log(test.name)
			mocks := testMocks{
				repo:    &mockrepository{},
				metrics: &mockmetricsSender{},
				client:  &mockclient{},
			}

			for _, given := range test.given {
				given(t, &mocks)
			}

			for _, expect := range test.expect {
				expect(t, &mocks)
			}

			service := NewService(mocks.client, mocks.repo, mocks.metrics)
			err := service.Replicate()

			test.verify(t, err)
		})
	}
}
