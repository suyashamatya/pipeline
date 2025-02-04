/*
 Copyright 2024 The Tekton Authors

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package main

import (
	"encoding/base64"
	"errors"
	"testing"
	"time"

	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	"github.com/tektoncd/pipeline/pkg/apis/resolution/v1beta1"
	ttesting "github.com/tektoncd/pipeline/pkg/reconciler/testing"
	frtesting "github.com/tektoncd/pipeline/pkg/remoteresolution/resolver/framework/testing"
	resolutioncommon "github.com/tektoncd/pipeline/pkg/resolution/common"
	"github.com/tektoncd/pipeline/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "knative.dev/pkg/apis/duck/v1"
	_ "knative.dev/pkg/system/testing"
)

func TestResolver(t *testing.T) {
	ctx, _ := ttesting.SetupFakeContext(t)

	r := &resolver{}

	request := &v1beta1.ResolutionRequest{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "resolution.tekton.dev/v1beta1",
			Kind:       "ResolutionRequest",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              "rr",
			Namespace:         "foo",
			CreationTimestamp: metav1.Time{Time: time.Now()},
			Labels: map[string]string{
				resolutioncommon.LabelKeyResolverType: "demo",
			},
		},
		Spec: v1beta1.ResolutionRequestSpec{
			URL: "demoscheme://foo/bar",
		},
	}
	d := test.Data{
		ResolutionRequests: []*v1beta1.ResolutionRequest{request},
	}

	expectedStatus := &v1beta1.ResolutionRequestStatus{
		ResolutionRequestStatusFields: v1beta1.ResolutionRequestStatusFields{
			Data: base64.StdEncoding.Strict().EncodeToString([]byte(pipeline)),
		},
	}

	// If you want to test scenarios where an error should occur, pass a non-nil error to RunResolverReconcileTest
	var expectedErr error

	frtesting.RunResolverReconcileTest(ctx, t, d, r, request, expectedStatus, expectedErr)
}

func TestResolver_Failure_Wrong_Scheme(t *testing.T) {
	ctx, _ := ttesting.SetupFakeContext(t)

	r := &resolver{}

	request := &v1beta1.ResolutionRequest{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "resolution.tekton.dev/v1beta1",
			Kind:       "ResolutionRequest",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              "rr",
			Namespace:         "foo",
			CreationTimestamp: metav1.Time{Time: time.Now()},
			Labels: map[string]string{
				resolutioncommon.LabelKeyResolverType: "demo",
			},
		},
		Spec: v1beta1.ResolutionRequestSpec{
			URL: "wrongscheme://foo/bar",
		},
	}
	d := test.Data{
		ResolutionRequests: []*v1beta1.ResolutionRequest{request},
	}

	expectedStatus := &v1beta1.ResolutionRequestStatus{
		Status: v1.Status{
			Conditions: v1.Conditions{
				{
					Type:    "Succeeded",
					Status:  "False",
					Reason:  "ResolutionFailed",
					Message: `invalid resource request "foo/rr": Invalid Scheme. Want demoscheme, Got wrongscheme`,
				},
			},
		},
	}

	// If you want to test scenarios where an error should occur, pass a non-nil error to RunResolverReconcileTest
	expectedErr := errors.New(`invalid resource request "foo/rr": Invalid Scheme. Want demoscheme, Got wrongscheme`)
	frtesting.RunResolverReconcileTest(ctx, t, d, r, request, expectedStatus, expectedErr)
}

func TestResolver_Failure_InvalidUrl(t *testing.T) {
	ctx, _ := ttesting.SetupFakeContext(t)

	r := &resolver{}

	request := &v1beta1.ResolutionRequest{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "resolution.tekton.dev/v1beta1",
			Kind:       "ResolutionRequest",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              "rr",
			Namespace:         "foo",
			CreationTimestamp: metav1.Time{Time: time.Now()},
			Labels: map[string]string{
				resolutioncommon.LabelKeyResolverType: "demo",
			},
		},
		Spec: v1beta1.ResolutionRequestSpec{
			URL: "foo/bar",
		},
	}
	d := test.Data{
		ResolutionRequests: []*v1beta1.ResolutionRequest{request},
	}

	expectedStatus := &v1beta1.ResolutionRequestStatus{
		Status: v1.Status{
			Conditions: v1.Conditions{
				{
					Type:    "Succeeded",
					Status:  "False",
					Reason:  "ResolutionFailed",
					Message: `invalid resource request "foo/rr": parse "foo/bar": invalid URI for request`,
				},
			},
		},
	}

	// If you want to test scenarios where an error should occur, pass a non-nil error to RunResolverReconcileTest
	expectedErr := errors.New(`invalid resource request "foo/rr": parse "foo/bar": invalid URI for request`)
	frtesting.RunResolverReconcileTest(ctx, t, d, r, request, expectedStatus, expectedErr)
}

func TestResolver_Failure_InvalidParams(t *testing.T) {
	ctx, _ := ttesting.SetupFakeContext(t)

	r := &resolver{}

	request := &v1beta1.ResolutionRequest{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "resolution.tekton.dev/v1beta1",
			Kind:       "ResolutionRequest",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              "rr",
			Namespace:         "foo",
			CreationTimestamp: metav1.Time{Time: time.Now()},
			Labels: map[string]string{
				resolutioncommon.LabelKeyResolverType: "demo",
			},
		},
		Spec: v1beta1.ResolutionRequestSpec{
			Params: []pipelinev1.Param{{
				Name:  "foo",
				Value: *pipelinev1.NewStructuredValues("bar"),
			}},
		},
	}
	d := test.Data{
		ResolutionRequests: []*v1beta1.ResolutionRequest{request},
	}

	expectedStatus := &v1beta1.ResolutionRequestStatus{
		Status: v1.Status{
			Conditions: v1.Conditions{
				{
					Type:    "Succeeded",
					Status:  "False",
					Reason:  "ResolutionFailed",
					Message: `invalid resource request "foo/rr": no params allowed`,
				},
			},
		},
	}

	// If you want to test scenarios where an error should occur, pass a non-nil error to RunResolverReconcileTest
	expectedErr := errors.New(`invalid resource request "foo/rr": no params allowed`)
	frtesting.RunResolverReconcileTest(ctx, t, d, r, request, expectedStatus, expectedErr)
}
