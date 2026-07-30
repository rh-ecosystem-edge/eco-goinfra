package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/oran/api/fields"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/oran/api/filter"
)

func TestApplyListOptions_LastFilterWins(t *testing.T) {
	t.Parallel()

	query := applyListOptions(
		WithFilter(filter.Equals("name", "first")),
		WithFilter(filter.Equals("name", "last")),
	)

	require.NotNil(t, query.filter)
	assert.Equal(t, "(eq,name,last)", *query.filter)
}

func TestApplyListOptions_LastFieldsWins(t *testing.T) {
	t.Parallel()

	query := applyListOptions(
		WithFields(fields.Include("first")),
		WithFields(fields.Include("last")),
	)

	require.NotNil(t, query.fields)
	assert.Equal(t, "last", *query.fields)
	assert.Nil(t, query.excludeFields)
	assert.Nil(t, query.allFields)
}

func TestApplyListOptions_FilterAndFieldsComposition(t *testing.T) {
	t.Parallel()

	query := applyListOptions(
		WithFilter(filter.Equals("name", "test-cluster")),
		WithFields(fields.Include("name", "nodeClusterId")),
	)

	require.NotNil(t, query.filter)
	assert.Equal(t, "(eq,name,test-cluster)", *query.filter)
	require.NotNil(t, query.fields)
	assert.Equal(t, "name,nodeClusterId", *query.fields)
}

func TestListQuery_HasOptions(t *testing.T) {
	t.Parallel()

	t.Run("no options", func(t *testing.T) {
		t.Parallel()

		assert.False(t, applyListOptions().hasOptions())
	})

	t.Run("with filter", func(t *testing.T) {
		t.Parallel()

		query := applyListOptions(WithFilter(filter.Equals("name", "test")))
		assert.True(t, query.hasOptions())
	})

	t.Run("with fields", func(t *testing.T) {
		t.Parallel()

		query := applyListOptions(WithFields(fields.Include("name")))
		assert.True(t, query.hasOptions())
	})
}
