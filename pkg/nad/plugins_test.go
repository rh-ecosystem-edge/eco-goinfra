package nad

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var sysctlBondPluginOptions = BondPluginOptions{
	FailOverMac:      1,
	LinksInContainer: true,
	Miimon:           "100",
}

func TestBondPlugin(t *testing.T) {
	testCases := []struct {
		name       string
		ipam       *IPAM
		bondPorts  []string
		bondMode   string
		opts       BondPluginOptions
		expectJSON []string
	}{
		{
			name:      "active-backup with static ipam",
			ipam:      IPAMStatic(),
			bondPorts: []string{"net1", "net2"},
			bondMode:  bondModeActiveBackup,
			opts:      sysctlBondPluginOptions,
			expectJSON: []string{
				`"type":"bond"`,
				fmt.Sprintf(`"mode":%q`, bondModeActiveBackup),
				`"failOverMac":1`,
				`"linksInContainer":true`,
				`"miimon":"100"`,
				`"ipam":{"type":"static"}`,
				`"links":[{"name":"net1"},{"name":"net2"}]`,
			},
		},
		{
			name:      "balance-xor with nil ipam",
			ipam:      nil,
			bondPorts: []string{"net4", "net5"},
			bondMode:  bondModeBalanceXOR,
			opts:      sysctlBondPluginOptions,
			expectJSON: []string{
				`"type":"bond"`,
				fmt.Sprintf(`"mode":%q`, bondModeBalanceXOR),
				`"links":[{"name":"net4"},{"name":"net5"}]`,
			},
		},
		{
			name:      "empty bondPorts",
			ipam:      IPAMStatic(),
			bondPorts: []string{},
			bondMode:  bondModeActiveBackup,
			opts:      sysctlBondPluginOptions,
		},
		{
			name:      "single bondPort",
			ipam:      IPAMStatic(),
			bondPorts: []string{"net1"},
			bondMode:  bondModeActiveBackup,
			opts:      sysctlBondPluginOptions,
			expectJSON: []string{
				`"links":[{"name":"net1"}]`,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			plugin := BondPlugin(testCase.ipam, testCase.bondPorts, testCase.bondMode, testCase.opts)
			assertBondPluginFields(t, plugin, testCase.ipam, testCase.bondMode, testCase.bondPorts, testCase.opts)
			assertBondPluginJSONFields(t, plugin, testCase.expectJSON)
		})
	}
}

func TestBondPluginInvalidMode(t *testing.T) {
	testCases := []struct {
		name     string
		bondMode string
	}{
		{name: "invalid bond mode", bondMode: "invalid-mode"},
		{name: "empty bond mode", bondMode: ""},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			plugin := BondPlugin(IPAMStatic(), []string{"net1", "net2"}, testCase.bondMode, sysctlBondPluginOptions)
			assert.Nil(t, plugin)
		})
	}
}

func assertBondPluginFields(
	t *testing.T,
	plugin *Plugin,
	ipam *IPAM,
	bondMode string,
	bondPorts []string,
	opts BondPluginOptions,
) {
	t.Helper()

	assert.NotNil(t, plugin)
	assert.Equal(t, "bond", plugin.Type)
	assert.Equal(t, bondMode, plugin.Mode)
	assert.Equal(t, opts.FailOverMac, plugin.FailOverMac)
	assert.Equal(t, opts.LinksInContainer, plugin.LinksInContainer)
	assert.Equal(t, opts.Miimon, plugin.Miimon)
	assert.Equal(t, ipam, plugin.Ipam)

	expectedLinks := make([]Link, 0, len(bondPorts))
	for _, port := range bondPorts {
		expectedLinks = append(expectedLinks, Link{Name: port})
	}

	assert.Equal(t, expectedLinks, plugin.Links)
}

func assertBondPluginJSONFields(t *testing.T, plugin *Plugin, expectJSON []string) {
	t.Helper()

	if len(expectJSON) == 0 {
		return
	}

	marshaled, err := json.Marshal(plugin)
	assert.NoError(t, err)

	marshaledJSON := string(marshaled)
	for _, expectedField := range expectJSON {
		assert.Contains(t, marshaledJSON, expectedField)
	}
}

func TestBondPluginOptions(t *testing.T) {
	opts := BondPluginOptions{
		FailOverMac:      2,
		LinksInContainer: false,
		Miimon:           "200",
	}

	plugin := BondPlugin(
		IPAMStatic(),
		[]string{"net1", "net2"},
		bondModeActiveBackup,
		opts,
	)

	assert.NotNil(t, plugin)
	assert.Equal(t, 2, plugin.FailOverMac)
	assert.False(t, plugin.LinksInContainer)
	assert.Equal(t, "200", plugin.Miimon)

	marshaled, err := json.Marshal(plugin)
	assert.NoError(t, err)
	assert.Contains(t, string(marshaled), `"failOverMac":2`)
	assert.Contains(t, string(marshaled), `"miimon":"200"`)
	assert.NotContains(t, string(marshaled), `"linksInContainer":true`)
}

func TestBondPluginValidModesMatchNewMasterBondPlugin(t *testing.T) {
	validModes := []string{
		bondModeBalanceRR,
		bondModeActiveBackup,
		bondModeBalanceXOR,
		bondModeBalanceTLB,
		bondModeBalanceALB,
	}

	for _, mode := range validModes {
		t.Run(mode, func(t *testing.T) {
			plugin := BondPlugin(IPAMStatic(), []string{"net1", "net2"}, mode, sysctlBondPluginOptions)
			assert.NotNil(t, plugin)
			assert.Equal(t, mode, plugin.Mode)

			masterBond, err := NewMasterBondPlugin("bond0", mode).GetMasterPluginConfig()
			assert.NoError(t, err)
			assert.Equal(t, mode, masterBond.Mode)
		})
	}
}
