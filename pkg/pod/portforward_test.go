package pod

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPortForward(t *testing.T) {
	testCases := []struct {
		name          string
		builder       *Builder
		expectedError error
	}{
		{
			name:          "invalid builder - empty namespace",
			builder:       buildInvalidPodTestBuilder(buildTestClientWithDummyPod()),
			expectedError: fmt.Errorf(errEmptyNamespace),
		},
		{
			name:          "nil builder",
			builder:       buildValidPodTestBuilder(nil),
			expectedError: fmt.Errorf("error: received nil Pod builder"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			addr, stop, err := testCase.builder.PortForward(8080, 8080)
			assert.Equal(t, testCase.expectedError, err)
			assert.Empty(t, addr)
			assert.Nil(t, stop)
		})
	}
}
