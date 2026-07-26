package nad

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMasterBondPluginWithXmitHashPolicy(t *testing.T) {
	testCases := []struct {
		name            string
		xmitHashPolicy  string
		expectError     bool
		expectedError   string
		expectJSONField string
	}{
		{
			name:            "valid layer2",
			xmitHashPolicy:  "layer2",
			expectError:     false,
			expectJSONField: `"xmitHashPolicy":"layer2"`,
		},
		{
			name:            "valid layer2+3",
			xmitHashPolicy:  "layer2+3",
			expectError:     false,
			expectJSONField: `"xmitHashPolicy":"layer2+3"`,
		},
		{
			name:            "valid layer3+4",
			xmitHashPolicy:  "layer3+4",
			expectError:     false,
			expectJSONField: `"xmitHashPolicy":"layer3+4"`,
		},
		{
			name:           "invalid policy",
			xmitHashPolicy: "invalid",
			expectError:    true,
			expectedError:  invalidXmitHashPolicyMsg,
		},
		{
			name:           "empty policy",
			xmitHashPolicy: "",
			expectError:    true,
			expectedError:  invalidXmitHashPolicyMsg,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			cfg, err := NewMasterBondPlugin("bond0", "balance-xor").
				WithXmitHashPolicy(testCase.xmitHashPolicy).
				GetMasterPluginConfig()

			if testCase.expectError {
				assert.Nil(t, cfg)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.expectedError)

				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, cfg)
			assert.Equal(t, testCase.xmitHashPolicy, cfg.XmitHashPolicy)

			marshaled, marshalErr := json.Marshal(cfg)
			assert.NoError(t, marshalErr)
			assert.Contains(t, string(marshaled), testCase.expectJSONField)
		})
	}
}

func TestMasterBondPluginWithXmitHashPolicyAcceptance(t *testing.T) {
	cfg, err := NewMasterBondPlugin("bond0", "balance-xor").
		WithLinksInContainer(true).
		WithFailOverMac(1).
		WithMiimon(100).
		WithLinks([]Link{{Name: "net1"}, {Name: "net2"}}).
		WithXmitHashPolicy("layer2+3").
		WithIPAM(&IPAM{Type: "static"}).
		GetMasterPluginConfig()

	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "layer2+3", cfg.XmitHashPolicy)

	marshaled, marshalErr := json.Marshal(cfg)
	assert.NoError(t, marshalErr)
	assert.Contains(t, string(marshaled), `"xmitHashPolicy":"layer2+3"`)
}
