---
name: common-builder-refactor
description: Refactor a builder to use the EmbeddableBuilder, its mixins, and the testhelper package.
disable-model-invocation: true
---

# Common Builder Refactor Guide

Refactors a package's `Builder` struct to use `pkg/internal/common`.

## Framework overview

- `EmbeddableBuilder[O, *O]` — stores Definition, Object, error, apiClient, gvk; provides `Get` and `Exists`
- Mixins — each provides one CRUD method; embedded in Builder and wired via `AttachMixins()`
- `common.NewNamespacedBuilder` / `common.NewClusterScopedBuilder` — replace `NewBuilder` boilerplate; auto-call `AttachMixins()` and `SetGVK()`
- `common.PullNamespacedBuilder` / `common.PullClusterScopedBuilder` — replace `Pull` boilerplate
- `common.Validate(builder)` — replaces hand-written `validate()`
- `testhelper` package — standard test configs for `NewBuilder`, `Pull`, and CRUD methods

## Step 0: Understand the package

Read all `.go` and `_test.go` files. Note:

- Cluster-scoped or namespaced?
- Which CRUD methods exist and their exact signatures
- `With*` methods and their validations
- Extra parameters in `NewBuilder` beyond `apiClient`, `name`, `nsname`
- Whether the type appears in `pkg/clients/clients.go` `GetModifiableTestClients`

Run `make lint && make test` to establish a baseline.

## Step 1: Refactor the Builder struct

```go
type Builder struct {
    common.EmbeddableBuilder[foov1.Foo, *foov1.Foo]
    // Embed only mixins for methods the original builder had (see table below).
}

// AttachMixins is called automatically by the common init functions (NewNamespacedBuilder etc.).
// Call SetBase(builder) on each embedded mixin so it can access the EmbeddableBuilder.
func (builder *Builder) AttachMixins() {
    builder.EmbeddableSomeMixin.SetBase(builder)
}

// GetGVK must be overridden on the Builder struct. The common init functions call builder.GetGVK()
// then SetGVK(result) to persist the value; without this override, a zero GVK would be stored.
func (builder *Builder) GetGVK() schema.GroupVersionKind {
    return foov1.GroupVersion.WithKind("Foo")
}
```

### Mixin selection — only embed if the original builder had that method

| Original method signature | Mixin |
| --- | --- |
| `Create() (*Builder, error)` | `EmbeddableCreator[O, Builder, *O, *Builder]` |
| `Delete() error` | `EmbeddableDeleter[O, *O]` |
| `Delete() (*Builder, error)` | `EmbeddableDeleteReturner[O, Builder, *O, *Builder]` |
| `Update() (*Builder, error)` | `EmbeddableUpdater[O, Builder, *O, *Builder]` |
| `Update(force bool) (*Builder, error)` | `EmbeddableForceUpdater[O, Builder, *O, *Builder]` |
| `WithOptions(options ...AdditionalOptions) *Builder` | `EmbeddableWithOptions[O, Builder, *O, *Builder, AdditionalOptions]` |

`Get` and `Exists` come from `EmbeddableBuilder` — no mixin needed.

Run `make lint` to verify no compilation errors before continuing.

## Step 2: Refactor NewBuilder and Pull

```go
// Use ClusterScoped variants for cluster-scoped resources (no nsname parameter).
func NewBuilder(apiClient *clients.Settings, name, nsname string) *Builder {
    return common.NewNamespacedBuilder[foov1.Foo, Builder](apiClient, foov1.AddToScheme, name, nsname)
}

func Pull(apiClient *clients.Settings, name, nsname string) (*Builder, error) {
    return common.PullNamespacedBuilder[foov1.Foo, Builder](context.TODO(), apiClient, foov1.AddToScheme, name, nsname)
}
```

If `NewBuilder` has **extra required parameters**, validate them after the common call:

```go
builder := common.NewNamespacedBuilder[foov1.Foo, Builder](apiClient, foov1.AddToScheme, name, nsname)
if builder.GetError() != nil {
    return builder
}
if extraParam == "" {
    builder.SetError(fmt.Errorf("foo 'extraParam' cannot be empty"))
    return builder
}
builder.Definition.Spec.ExtraParam = extraParam
return builder
```

## Step 3: Remove replaced code and update With* methods

Delete:

- `validate()` — replaced by `common.Validate(builder)`
- Any method whose behavior is now provided by a mixin
- Unused imports (`msg`, `logging`, `k8serrors`, `metav1`, `goclient`)

In `With*` methods, replace the old validate/error pattern: swap `if valid, _ := builder.validate(); !valid { return builder }` with `if err := common.Validate(builder); err != nil { return builder }`, and swap `builder.errorMsg = "..."` / `if builder.errorMsg != ""` with `builder.SetError(fmt.Errorf(...))` / `if builder.GetError() != nil`.

## Step 4: Update clients.go

If the type appears in `GetModifiableTestClients`, remove that case. The common testhelper registers schemes via `SchemeAttachers` in `TestClientParams` instead.

---

Before moving on to step 5, you must run `make lint && make test` to verify that the refactor is correct. There may be some failures, but understand these deeply before touching test files. You should ensure that updating the tests does not cover up any regressions from your refactor. If you notice a change in behavior, alert the user and ask them if this is expected and acceptable before updating the tests.

---

## Step 5: Update test files

### NewBuilder tests

```go
func TestNewBuilder(t *testing.T) {
    t.Parallel()
    t.Run("common namespaced builder behavior", func(t *testing.T) {
        t.Parallel()
        testhelper.NewNamespacedBuilderTestConfig(
            func(apiClient *clients.Settings, name, nsname string) *Builder {
                return NewBuilder(apiClient, name, nsname /*, required extra params with valid values */)
            },
            foov1.AddToScheme,
            foov1.GroupVersion.WithKind("Foo"),
        ).ExecuteTests(t)
    })
    // Add package-specific cases only (e.g., extra parameters validation).
}
```

Use `NewClusterScopedBuilderTestConfig` for cluster-scoped resources.

### Pull tests

```go
testhelper.NewNamespacedPullTestConfig(Pull, foov1.AddToScheme, foov1.GroupVersion.WithKind("Foo")).ExecuteTests(t)
```

### Get, Exists, and mixin method tests

```go
commonTestConfig := testhelper.NewCommonTestConfig[foov1.Foo, Builder](
    foov1.AddToScheme, foov1.GroupVersion.WithKind("Foo"), testhelper.ResourceScopeNamespaced)

testhelper.NewTestSuite().
    With(testhelper.NewGetTestConfig(commonTestConfig)).
    With(testhelper.NewExistsTestConfig(commonTestConfig)).
    // With(testhelper.NewCreateTestConfig(commonTestConfig)).
    // With(testhelper.NewDeleteTestConfig(commonTestConfig)).
    // With(testhelper.NewDeleteReturnerTestConfig(commonTestConfig)).
    // With(testhelper.NewUpdateTestConfig(commonTestConfig)).
    // With(testhelper.NewForceUpdateTestConfig(commonTestConfig)).
    Run(t)
```

Only include configs for methods that exist on the builder.

### WithOptions tests

If the builder uses `EmbeddableWithOptions`, add its test config:

```go
func TestWithOptions(t *testing.T) {
    t.Parallel()
    testhelper.NewWithOptionsTestConfig(commonTestConfig).ExecuteTests(t)
}
```

### With* modifier tests

Each `With*` test must include an invalid-builder case to verify short-circuit behavior. Use `GetError()` instead of `errorMsg`. Assert the field value on success:

```go
assert.Equal(t, testCase.expectedError, testBuilder.GetError())
if testCase.expectedError == nil {
    assert.Equal(t, testCase.inputValue, testBuilder.Definition.Spec.Field)
}
```

Add `t.Parallel()` to every test function and subtest.

### Validate tests

You may remove any tests for the `validate` method, if they exist. The `common.Validate` function is already tested, so these become redundant after the refactor to use the `common` package.

### Test helpers

```go
func buildValidFooTestBuilder(apiClient *clients.Settings) *Builder {
    return NewBuilder(apiClient, "foo-name", "foo-namespace" /*, valid extra params */)
}

// buildInvalidFooTestBuilder must produce a builder already in an error state.
// For packages with extra required params, pass an invalid extra param with valid name/nsname —
// not an empty name, since that tests a different error path handled by the common framework.
func buildInvalidFooTestBuilder(apiClient *clients.Settings) *Builder {
    return NewBuilder(apiClient, "", "foo-namespace") // or: valid name/ns + invalid extra param
}

func getTestFooAPIClient() *clients.Settings {
    return clients.GetTestClients(clients.TestClientParams{})
}
```

## Step 6: Final validation

```sh
make lint && make test
```

Verify: no test cases silently removed; all `With*` tests cover both error and success paths; hand-written Get/Exists/mixin tests replaced by testhelper configs.

## Quick reference: error types

| Situation | Check with |
| --- | --- |
| nil apiClient | `commonerrors.IsAPIClientNil(err)` |
| empty name | `commonerrors.IsBuilderNameEmpty(err)` |
| empty namespace | `commonerrors.IsBuilderNamespaceEmpty(err)` |
| nil builder | `commonerrors.IsBuilderNil(err)` |

Testhelper configs assert these automatically; only use them in package-specific test cases.
