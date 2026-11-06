package clusteroperator

import (
	"fmt"
	"testing"

	configv1 "github.com/openshift/api/config/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
)

func TestList(t *testing.T) {
	t.Parallel()

	testhelper.NewListTestConfig(
		List,
		configv1.Install,
		clusterOperatorGVK,
	).ExecuteTests(t)
}

func TestVerifyClusterOperatorsVersion(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                string
		desiredVersion      string
		clusterOperatorList []*Builder
		expectedOutput      bool
		expectedError       error
	}{
		{
			name:                "empty desired version",
			desiredVersion:      "",
			expectedOutput:      false,
			clusterOperatorList: buildFakeClusterOperatorListWithDesiredVersion(buildClusterOperatorClientWithDummyObject()),
			expectedError:       fmt.Errorf("desiredVersion can't be empty"),
		},
		{
			name:                "desired version missing",
			desiredVersion:      "4.14.0",
			expectedOutput:      false,
			clusterOperatorList: buildFakeClusterOperatorListWithoutDesiredVersion(buildClusterOperatorClientWithDummyObject()),
			expectedError: fmt.Errorf("the clusterOperator %s doesn't have the desired version 4.14.0",
				defaultClusterOperatorName),
		},
		{
			name:                "desired version present",
			desiredVersion:      "4.14.0",
			expectedOutput:      true,
			clusterOperatorList: buildFakeClusterOperatorListWithDesiredVersion(buildClusterOperatorClientWithDummyObject()),
			expectedError:       nil,
		},
		{
			name:                "nil cluster operator list",
			desiredVersion:      "4.14.0",
			expectedOutput:      false,
			clusterOperatorList: nil,
			expectedError:       fmt.Errorf("clusterOperatorList is invalid"),
		},
		{
			name:                "empty cluster operator list",
			desiredVersion:      "4.14.0",
			expectedOutput:      false,
			clusterOperatorList: []*Builder{},
			expectedError:       fmt.Errorf("clusterOperatorList can't be empty"),
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result, err := VerifyClusterOperatorsVersion(testCase.desiredVersion, testCase.clusterOperatorList)
			assert.Equal(t, testCase.expectedOutput, result)
			assert.Equal(t, testCase.expectedError, err)
		})
	}
}
