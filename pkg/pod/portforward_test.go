package pod

import (
	"testing"

	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPortForward(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		builder     *Builder
		assertError func(error) bool
	}{
		{
			name:        "invalid builder - empty namespace",
			builder:     buildInvalidPodTestBuilder(buildTestClientWithDummyPod()),
			assertError: commonerrors.IsBuilderNamespaceEmpty,
		},
		{
			name:        "nil client",
			builder:     buildValidPodTestBuilder(nil),
			assertError: commonerrors.IsAPIClientNil,
		},
		{
			name:        "nil builder",
			builder:     nil,
			assertError: commonerrors.IsBuilderNil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			addr, stop, err := testCase.builder.PortForward(8080, 8080)
			require.Error(t, err)
			assert.True(t, testCase.assertError(err))
			assert.Empty(t, addr)
			assert.Nil(t, stop)
		})
	}
}
