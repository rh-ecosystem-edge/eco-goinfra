package ocm

import (
	"fmt"
	"testing"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/labels"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func TestListPlacementBindingsInAllNamespaces(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		listOptions   []runtimeclient.ListOptions
		expectedError error
		client        bool
	}{
		{
			name:          "list all",
			listOptions:   nil,
			expectedError: nil,
			client:        true,
		},
		{
			name:          "list with options",
			listOptions:   []runtimeclient.ListOptions{{LabelSelector: labels.NewSelector()}},
			expectedError: nil,
			client:        true,
		},
		{
			name: "too many list options",
			listOptions: []runtimeclient.ListOptions{
				{LabelSelector: labels.NewSelector()},
				{LabelSelector: labels.NewSelector()},
			},
			expectedError: fmt.Errorf("error: more than one ListOptions was passed"),
			client:        true,
		},
		{
			name:          "nil client",
			listOptions:   []runtimeclient.ListOptions{{LabelSelector: labels.NewSelector()}},
			expectedError: fmt.Errorf("apiClient for PlacementBinding is nil"),
			client:        false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var testSettings *clients.Settings

			if testCase.client {
				testSettings = buildTestClientWithDummyPlacementBinding()
			}

			builders, err := ListPlacementBindingsInAllNamespaces(testSettings, testCase.listOptions...)

			switch testCase.name {
			case "nil client":
				require.Error(t, err)
				assert.True(t, commonerrors.IsAPIClientNil(err))
				assert.Nil(t, builders)
			case "too many list options":
				assert.Equal(t, testCase.expectedError, err)
			default:
				assert.NoError(t, err)

				if len(testCase.listOptions) == 0 {
					assert.Len(t, builders, 1)
				}
			}
		})
	}
}
