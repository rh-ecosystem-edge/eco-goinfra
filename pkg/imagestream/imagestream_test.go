package imagestream

import (
	"fmt"
	"testing"

	imagev1 "github.com/openshift/api/image/v1"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	imageStreamName      = "network-tools"
	imageStreamNamespace = "openshift"
	imageTag             = "latest"
	imageStreamGVK       = imagev1.GroupVersion.WithKind("ImageStream")
	testSchemes          = []clients.SchemeAttacher{
		imagev1.AddToScheme,
	}
)

func TestImageStreamPull(t *testing.T) {
	t.Parallel()

	testhelper.NewNamespacedPullTestConfig(
		Pull,
		imagev1.AddToScheme,
		imageStreamGVK,
	).ExecuteTests(t)
}

func TestImageStreamMethods(t *testing.T) {
	t.Parallel()

	commonTestConfig := testhelper.NewCommonTestConfig[imagev1.ImageStream, Builder](
		imagev1.AddToScheme,
		imageStreamGVK,
		testhelper.ResourceScopeNamespaced,
	)

	testhelper.NewTestSuite().
		With(testhelper.NewGetTestConfig(commonTestConfig)).
		With(testhelper.NewExistsTestConfig(commonTestConfig)).
		With(testhelper.NewDeleterTestConfig(commonTestConfig)).
		Run(t)
}

//nolint:funlen
func TestImageStreamGetDockerImage(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		testImageStreamTags []imagev1.TagReference
		testImageTag        string
		expectedError       error
	}{
		{
			testImageStreamTags: []imagev1.TagReference{{
				Name: "latest",
				From: &corev1.ObjectReference{
					Kind: "DockerImage",
					Name: "quay.io/dummy-server/ocp-v4.0-art-dev@sha256:c8bc1d7bdf77538653ff1cd40d9bfa00f0",
				},
			}},
			testImageTag:  imageTag,
			expectedError: nil,
		},
		{
			testImageStreamTags: []imagev1.TagReference{{
				Name: "7.4.0",
				From: &corev1.ObjectReference{
					Kind: "DockerImage",
					Name: "registry.dummy.io/jboss-eap-7/eap74-openjdk8-openshift-rhel7:latest",
				}}, {
				Name: "latest",
				From: &corev1.ObjectReference{
					Kind: "DockerImage",
					Name: "registry.dummy.io/jboss-eap-7/eap74-openjdk8-openshift-rhel7:latest",
				},
			}},
			testImageTag:  imageTag,
			expectedError: nil,
		},
		{
			testImageStreamTags: []imagev1.TagReference{},
			testImageTag:        imageTag,
			expectedError:       fmt.Errorf("imageStream object network-tools in namespace openshift has no tags"),
		},
		{
			testImageStreamTags: []imagev1.TagReference{{
				Name: "latest",
				From: &corev1.ObjectReference{
					Kind: "DockerImage",
					Name: "quay.io/dummy-server/ocp-v4.0-art-dev@sha256:c8bc1d7bdf77538653ff1cd40d9bfa00f0",
				},
			}},
			testImageTag: "4.16",
			expectedError: fmt.Errorf("image tag 4.16 not found for imageStream object network-tools " +
				"in namespace openshift"),
		},
		{
			testImageStreamTags: []imagev1.TagReference{{
				Name: "",
				From: &corev1.ObjectReference{
					Kind: "DockerImage",
					Name: "quay.io/dummy-server/ocp-v4.0-art-dev@sha256:c8bc1d7bdf77538653ff1cd40d9bfa00f0",
				},
			}},
			testImageTag:  imageTag,
			expectedError: fmt.Errorf("imageStream object network-tools in namespace openshift has no DockerImage tag value"),
		},
		{
			testImageStreamTags: []imagev1.TagReference{{
				Name: "latest",
				From: nil,
			}},
			testImageTag:  imageTag,
			expectedError: fmt.Errorf("imageStream object network-tools in namespace openshift has no DockerImage value"),
		},
		{
			testImageStreamTags: []imagev1.TagReference{{
				Name: "latest",
				From: &corev1.ObjectReference{
					Kind: "DockerImage",
					Name: "quay.io/dummy-server/ocp-v4.0-art-dev@sha256:c8bc1d7bdf77538653ff1cd40d9bfa00f0",
				},
			}},
			testImageTag:  "",
			expectedError: fmt.Errorf("imageStream 'imageTag' cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object

		testImageStreamClientWithDummyObject := buildDummyImageStreamWithTag(testCase.testImageStreamTags)

		runtimeObjects = append(runtimeObjects, testImageStreamClientWithDummyObject)

		dummyImageStream := clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects:  runtimeObjects,
			SchemeAttachers: testSchemes,
		})

		testImageStreamBuilder := buildValidImageStreamBuilder(dummyImageStream)

		dockerImage, err := testImageStreamBuilder.GetDockerImage(testCase.testImageTag)
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.NotEqual(t, "", dockerImage)
		} else {
			assert.Equal(t, "", dockerImage)
		}
	}
}

func TestImageStreamGetStatusTags(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		testImageStreamStatus imagev1.ImageStreamStatus
		expectedError         error
		expectedTags          []string
	}{
		{
			testImageStreamStatus: imagev1.ImageStreamStatus{
				Tags: []imagev1.NamedTagEventList{
					{Tag: "5.14.0-427.81.1.el9_4.x86_64"},
					{Tag: "5.14.0-427.85.1.el9_4.x86_64"},
				},
			},
			expectedError: nil,
			expectedTags:  []string{"5.14.0-427.81.1.el9_4.x86_64", "5.14.0-427.85.1.el9_4.x86_64"},
		},
		{
			testImageStreamStatus: imagev1.ImageStreamStatus{
				Tags: []imagev1.NamedTagEventList{},
			},
			expectedError: fmt.Errorf("imageStream object network-tools in namespace openshift has no status tags"),
			expectedTags:  nil,
		},
	}

	for _, testCase := range testCases {
		var runtimeObjects []runtime.Object

		testImageStream := &imagev1.ImageStream{
			ObjectMeta: metav1.ObjectMeta{
				Name:      imageStreamName,
				Namespace: imageStreamNamespace,
			},
			Status: testCase.testImageStreamStatus,
		}
		runtimeObjects = append(runtimeObjects, testImageStream)

		testSettings := clients.GetTestClients(clients.TestClientParams{
			K8sMockObjects:  runtimeObjects,
			SchemeAttachers: testSchemes,
		})

		builder, err := Pull(testSettings, imageStreamName, imageStreamNamespace)
		assert.NoError(t, err)
		assert.NotNil(t, builder)

		tags, err := builder.GetStatusTags()
		assert.Equal(t, testCase.expectedError, err)

		if testCase.expectedError == nil {
			assert.Equal(t, testCase.expectedTags, tags)
		} else {
			assert.Nil(t, tags)
		}
	}
}

func buildValidImageStreamBuilder(apiClient *clients.Settings) *Builder {
	return common.NewNamespacedBuilder[imagev1.ImageStream, Builder](
		apiClient, imagev1.AddToScheme, imageStreamName, imageStreamNamespace)
}

func buildDummyImageStreamWithTag(imagestreamTags []imagev1.TagReference) *imagev1.ImageStream {
	return &imagev1.ImageStream{
		ObjectMeta: metav1.ObjectMeta{
			Name:      imageStreamName,
			Namespace: imageStreamNamespace,
		},
		Spec: imagev1.ImageStreamSpec{
			LookupPolicy: imagev1.ImageLookupPolicy{
				Local: false,
			},
			Tags: imagestreamTags,
		},
	}
}
