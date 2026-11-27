package servicemesh

import (
	"testing"
	"time"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	commonerrors "github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/errors"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	istiov1 "maistra.io/api/core/v1"
)

var (
	defaultMemberRollName       = "default"
	defaultServiceMeshNamespace = "istio-system"
	defaultMembersList          = []string(nil)
	newMembersList              = []string{"test-ns1", "test-ns2"}
	readyCondition              = istiov1.ServiceMeshMemberRollCondition{
		Type:               istiov1.ConditionTypeMemberRollReady,
		Status:             corev1.ConditionTrue,
		LastTransitionTime: metav1.Time{Time: time.Now()},
		Reason:             "Configured",
		Message:            "All namespaces have been configured successfully",
	}
	notReadyCondition = istiov1.ServiceMeshMemberRollCondition{
		Type:               istiov1.ConditionTypeMemberRollReady,
		Status:             corev1.ConditionFalse,
		LastTransitionTime: metav1.Time{Time: time.Now()},
		Reason:             "ErrSMCPMissing",
		Message:            "No ServiceMeshControlPlane exists in the namespace",
	}
)

func TestPullMemberRoll(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		PullMemberRoll,
		istiov1.AddToScheme,
		memberRollGVK,
	).ExecuteTests(t)
}

func TestNewMemberRollBuilder(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedBuilderTestConfig(
		NewMemberRollBuilder,
		istiov1.AddToScheme,
		memberRollGVK,
	).ExecuteTests(t)
}

func TestMemberRollBuilderMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[istiov1.ServiceMeshMemberRoll, MemberRollBuilder](
		istiov1.AddToScheme,
		memberRollGVK,
		testhelper.ResourceScopeNamespaced,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewCreateTestConfig(commonTestConfig)).
		With(testhelper.NewDeleterTestConfig(commonTestConfig)).
		With(testhelper.NewForceUpdateTestConfig(commonTestConfig)).
		Run(t)
}

func TestMemberRollWithMembersList(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		members        []string
		validBuilder   bool
		expectedErrMsg string
	}{
		{
			name:         "valid members list",
			members:      newMembersList,
			validBuilder: true,
		},
		{
			name:           "empty members list",
			validBuilder:   true,
			expectedErrMsg: "can not modify memberRoll config with empty membersList",
		},
		{
			name:         "invalid builder",
			members:      newMembersList,
			validBuilder: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := buildMemberRollBuilderForTest(testCase.validBuilder).
				WithMembersList(testCase.members)

			if !testCase.validBuilder {
				require.Error(t, result.GetError())
				assert.True(t, commonerrors.IsBuilderNameEmpty(result.GetError()))
				assert.Equal(t, defaultMembersList, result.Definition.Spec.Members)

				return
			}

			if testCase.expectedErrMsg != "" {
				assert.EqualError(t, result.GetError(), testCase.expectedErrMsg)

				return
			}

			require.NoError(t, result.GetError())
			assert.Equal(t, testCase.members, result.Definition.Spec.Members)
		})
	}
}

func TestMemberRollGetMembersList(t *testing.T) {
	t.Parallel()

	t.Run("existing member roll", func(t *testing.T) {
		t.Parallel()

		builder := NewMemberRollBuilder(
			buildMemberRollClientWithDummyObject(),
			defaultMemberRollName,
			defaultServiceMeshNamespace,
		)

		members, err := builder.GetMembersList()

		require.NoError(t, err)
		require.NotNil(t, members)
		assert.Equal(t, builder.Object.Spec.Members, *members)
	})

	t.Run("missing member roll", func(t *testing.T) {
		t.Parallel()

		builder := buildMemberRollBuilderForTest(true)
		members, err := builder.GetMembersList()

		assert.Nil(t, members)
		assert.EqualError(t, err,
			"memberRoll object default does not exist in namespace istio-system")
	})

	t.Run("invalid builder", func(t *testing.T) {
		t.Parallel()

		members, err := buildMemberRollBuilderForTest(false).GetMembersList()

		assert.Nil(t, members)
		assert.True(t, commonerrors.IsBuilderNameEmpty(err))
	})
}

func TestMemberRollIsReady(t *testing.T) {
	t.Parallel()

	const timeout = 2 * time.Second

	t.Run("ready member roll", func(t *testing.T) {
		t.Parallel()

		builder := NewMemberRollBuilder(
			buildMemberRollClientWithCondition(readyCondition),
			defaultMemberRollName,
			defaultServiceMeshNamespace,
		)

		ready, err := builder.IsReady(timeout)

		require.NoError(t, err)
		assert.True(t, ready)
	})

	t.Run("member roll not ready", func(t *testing.T) {
		t.Parallel()

		builder := NewMemberRollBuilder(
			buildMemberRollClientWithCondition(notReadyCondition),
			defaultMemberRollName,
			defaultServiceMeshNamespace,
		)

		ready, err := builder.IsReady(timeout)

		assert.False(t, ready)
		assert.EqualError(t, err, "the Ready condition did not reached for the Service Mesh MemberRoll "+
			"default in namespace istio-system during 2s; context deadline exceeded")
	})

	t.Run("missing member roll", func(t *testing.T) {
		t.Parallel()

		ready, err := buildMemberRollBuilderForTest(true).IsReady(timeout)

		assert.False(t, ready)
		assert.EqualError(t, err, "the Ready condition did not reached for the Service Mesh MemberRoll "+
			"default in namespace istio-system during 2s; context deadline exceeded")
	})

	t.Run("invalid builder", func(t *testing.T) {
		t.Parallel()

		ready, err := buildMemberRollBuilderForTest(false).IsReady(timeout)

		assert.False(t, ready)
		assert.True(t, commonerrors.IsBuilderNameEmpty(err))
	})
}

func buildMemberRollBuilderForTest(valid bool) *MemberRollBuilder {
	if !valid {
		return NewMemberRollBuilder(
			clients.GetTestClients(clients.TestClientParams{}),
			"",
			defaultServiceMeshNamespace,
		)
	}

	return NewMemberRollBuilder(
		clients.GetTestClients(clients.TestClientParams{}),
		defaultMemberRollName,
		defaultServiceMeshNamespace,
	)
}

func buildMemberRollClientWithDummyObject() *clients.Settings {
	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  buildDummyMemberRoll(),
		SchemeAttachers: []clients.SchemeAttacher{istiov1.AddToScheme},
	})
}

func buildMemberRollClientWithCondition(condition istiov1.ServiceMeshMemberRollCondition) *clients.Settings {
	memberRoll := &istiov1.ServiceMeshMemberRoll{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultMemberRollName,
			Namespace: defaultServiceMeshNamespace,
		},
		Status: istiov1.ServiceMeshMemberRollStatus{
			Conditions: []istiov1.ServiceMeshMemberRollCondition{condition},
		},
	}

	return clients.GetTestClients(clients.TestClientParams{
		K8sMockObjects:  []runtime.Object{memberRoll},
		SchemeAttachers: []clients.SchemeAttacher{istiov1.AddToScheme},
	})
}

func buildDummyMemberRoll() []runtime.Object {
	return []runtime.Object{&istiov1.ServiceMeshMemberRoll{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultMemberRollName,
			Namespace: defaultServiceMeshNamespace,
		},
		Spec: istiov1.ServiceMeshMemberRollSpec{
			Members: []string{},
		},
	}}
}
