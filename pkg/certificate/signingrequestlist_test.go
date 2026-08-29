package certificate

import (
	"context"
	"testing"
	"time"

	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/clients"
	"github.com/rh-ecosystem-edge/eco-goinfra/pkg/internal/common/testhelper"
	"github.com/stretchr/testify/assert"
	certificatesv1 "k8s.io/api/certificates/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const defaultSigningRequestName = "test-signing-request"

func TestListSigningRequests(t *testing.T) {
	t.Parallel()

	testhelper.NewListTestConfig(
		ListSigningRequests,
		certificatesv1.AddToScheme,
		signingRequestGVK,
	).ExecuteTests(t)
}

func TestWaitUntilSigningRequestsApproved(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name          string
		approved      bool
		expectedError error
	}{
		{
			name:          "all approved",
			approved:      true,
			expectedError: nil,
		},
		{
			name:          "not approved times out",
			approved:      false,
			expectedError: context.DeadlineExceeded,
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var runtimeObjects []runtime.Object

			if testCase.approved {
				runtimeObjects = append(runtimeObjects, buildDummyApprovedSigningRequest())
			} else {
				runtimeObjects = append(runtimeObjects, buildDummySigningRequest())
			}

			testSettings := clients.GetTestClients(clients.TestClientParams{
				K8sMockObjects:  runtimeObjects,
				SchemeAttachers: []clients.SchemeAttacher{certificatesv1.AddToScheme},
			})

			err := WaitUntilSigningRequestsApproved(testSettings, time.Second)
			assert.Equal(t, testCase.expectedError, err)
		})
	}
}

func buildDummySigningRequest() *certificatesv1.CertificateSigningRequest {
	return &certificatesv1.CertificateSigningRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name: defaultSigningRequestName,
		},
	}
}

func buildDummyApprovedSigningRequest() *certificatesv1.CertificateSigningRequest {
	signingRequest := buildDummySigningRequest()
	signingRequest.Status = certificatesv1.CertificateSigningRequestStatus{
		Conditions: []certificatesv1.CertificateSigningRequestCondition{{
			Type:   certificatesv1.CertificateApproved,
			Status: corev1.ConditionTrue,
		}},
	}

	return signingRequest
}
