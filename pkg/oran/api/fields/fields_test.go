package fields

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInclude(t *testing.T) {
	t.Parallel()

	selection := Include("name", "nodeClusterId")
	fields, excludeFields, allFields := selection.Params()

	require.NotNil(t, fields)
	assert.Equal(t, "name,nodeClusterId", *fields)
	assert.Nil(t, excludeFields)
	assert.Nil(t, allFields)
}

func TestExclude(t *testing.T) {
	t.Parallel()

	selection := Exclude("extensions/country")
	fields, excludeFields, allFields := selection.Params()

	assert.Nil(t, fields)
	require.NotNil(t, excludeFields)
	assert.Equal(t, "extensions/country", *excludeFields)
	assert.Nil(t, allFields)
}

func TestAll(t *testing.T) {
	t.Parallel()

	selection := All()
	fields, excludeFields, allFields := selection.Params()

	assert.Nil(t, fields)
	assert.Nil(t, excludeFields)
	require.NotNil(t, allFields)
	assert.Equal(t, "", *allFields)
}

func TestPath(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "extensions/country", Path("extensions", "country"))
}

func TestWithExclude(t *testing.T) {
	t.Parallel()

	selection := Include("name", "extensions").WithExclude("extensions/country")
	fields, excludeFields, allFields := selection.Params()

	require.NotNil(t, fields)
	assert.Equal(t, "name,extensions", *fields)
	require.NotNil(t, excludeFields)
	assert.Equal(t, "extensions/country", *excludeFields)
	assert.Nil(t, allFields)
}

func TestParamsNilSelection(t *testing.T) {
	t.Parallel()

	var selection *Selection

	fields, excludeFields, allFields := selection.Params()

	assert.Nil(t, fields)
	assert.Nil(t, excludeFields)
	assert.Nil(t, allFields)
}
