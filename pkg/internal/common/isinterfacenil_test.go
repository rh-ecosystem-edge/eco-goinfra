package common

import (
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/stretchr/testify/assert"
)

func TestIsInterfaceNil(t *testing.T) {
	t.Parallel()

	var typedNilSettings *clients.Settings

	testCases := []struct {
		name     string
		value    any
		expected bool
	}{
		{
			name:     "nil value",
			value:    nil,
			expected: true,
		},
		{
			name:     "typed nil settings",
			value:    typedNilSettings,
			expected: true,
		},
		{
			name:     "settings without embedded client",
			value:    &clients.Settings{},
			expected: true,
		},
		{
			name:     "settings with embedded client",
			value:    clients.GetTestClients(clients.TestClientParams{}),
			expected: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, testCase.expected, isInterfaceNil(testCase.value))
		})
	}
}
