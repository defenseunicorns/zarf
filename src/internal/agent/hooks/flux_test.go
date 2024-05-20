// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

package hooks

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/defenseunicorns/zarf/src/internal/agent/http/admission"
	"github.com/defenseunicorns/zarf/src/internal/agent/operations"
	"github.com/defenseunicorns/zarf/src/pkg/cluster"
	"github.com/defenseunicorns/zarf/src/pkg/k8s"
	"github.com/defenseunicorns/zarf/src/types"
	flux "github.com/fluxcd/source-controller/api/v1"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

// createFluxGitRepoAdmissionRequest creates an admission request for a Flux GitRepository.
func createFluxGitRepoAdmissionRequest(t *testing.T, op v1.Operation, fluxGitRepo *flux.GitRepository) *v1.AdmissionRequest {
	t.Helper()
	raw, err := json.Marshal(fluxGitRepo)
	require.NoError(t, err)
	return &v1.AdmissionRequest{
		Operation: op,
		Object: runtime.RawExtension{
			Raw: raw,
		},
	}
}

func TestFluxMutationWebhook(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	c := &cluster.Cluster{K8s: &k8s.K8s{Clientset: fake.NewSimpleClientset()}}
	handler := admission.NewHandler().Serve(NewGitRepositoryMutationHook(ctx, c))

	state, err := json.Marshal(&types.ZarfState{GitServer: types.GitServerInfo{
		Address:      "https://git-server.com",
		PushUsername: "a-push-user",
	}})
	require.NoError(t, err)

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.ZarfStateSecretName,
			Namespace: cluster.ZarfNamespaceName,
		},
		Data: map[string][]byte{
			cluster.ZarfStateDataKey: state,
		},
	}
	c.Clientset.CoreV1().Secrets(cluster.ZarfNamespaceName).Create(ctx, secret, metav1.CreateOptions{})

	tests := []struct {
		name          string
		admissionReq  *v1.AdmissionRequest
		expectedPatch []operations.PatchOperation
		code          int
	}{
		{
			name: "should be mutated",
			admissionReq: createFluxGitRepoAdmissionRequest(t, v1.Create, &flux.GitRepository{
				ObjectMeta: metav1.ObjectMeta{
					Name: "mutate-this",
				},
				Spec: flux.GitRepositorySpec{
					URL: "https://github.com/stefanprodan/podinfo.git",
					Reference: &flux.GitRepositoryRef{
						Tag: "6.4.0",
					},
				},
			}),
			expectedPatch: []operations.PatchOperation{
				operations.ReplacePatchOperation(
					"/spec/url",
					"https://git-server.com/a-push-user/podinfo-1646971829.git",
				),
				operations.AddPatchOperation(
					"/spec/secretRef",
					map[string]string{"name": "private-git-server"},
				),
			},
			code: http.StatusOK,
		},
		{
			name: "should not mutate invalid git url",
			admissionReq: createFluxGitRepoAdmissionRequest(t, v1.Update, &flux.GitRepository{
				ObjectMeta: metav1.ObjectMeta{
					Name: "mutate-this",
				},
				Spec: flux.GitRepositorySpec{
					URL: "not-a-git-url",
					Reference: &flux.GitRepositoryRef{
						Tag: "6.4.0",
					},
				},
			}),
			expectedPatch: nil,
			code:          http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			resp := sendAdmissionRequest(t, tt.admissionReq, handler, tt.code)
			if tt.expectedPatch != nil {
				expectedPatchJSON, err := json.Marshal(tt.expectedPatch)
				require.NoError(t, err)
				require.JSONEq(t, string(expectedPatchJSON), string(resp.Patch))
			} else if tt.code != http.StatusInternalServerError {
				require.Empty(t, string(resp.Patch))
			}
		})
	}
}
