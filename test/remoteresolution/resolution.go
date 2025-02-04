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

package test

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/go-cmp/cmp"
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	resource "github.com/tektoncd/pipeline/pkg/remoteresolution/resource"
	resolution "github.com/tektoncd/pipeline/pkg/resolution/common"
	"github.com/tektoncd/pipeline/test/diff"
)

var _ resource.Requester = &Requester{}
var _ resolution.ResolvedResource = &ResolvedResource{}

// NewResolvedResource creates a mock resolved resource that is
// populated with the given data and annotations or returns the given
// error from its Data() method.
func NewResolvedResource(data []byte, annotations map[string]string, source *pipelinev1.RefSource, dataErr error) *ResolvedResource {
	return &ResolvedResource{
		ResolvedData:        data,
		ResolvedAnnotations: annotations,
		ResolvedRefSource:   source,
		DataErr:             dataErr,
	}
}

// NewRequester creates a mock requester that resolves to the given
// resource or returns the given error on Submit().
func NewRequester(resource resolution.ResolvedResource, err error, resolverPayload resource.ResolverPayload) *Requester {
	return &Requester{
		ResolvedResource: resource,
		SubmitErr:        err,
		ResolverPayload:  resolverPayload,
	}
}

// Requester implements resolution.Requester and makes it easier
// to mock the outcome of a remote pipelineRef or taskRef resolution.
type Requester struct {
	// The resolved resource object to return when a request is
	// submitted.
	ResolvedResource resolution.ResolvedResource
	// An error to return when a request is submitted.
	SubmitErr error
	// ResolverPayload that should match that of the request in order to return the resolved resource
	ResolverPayload resource.ResolverPayload
}

// Submit implements resolution.Requester, accepting the name of a
// resolver and a request for a specific remote file, and then returns
// whatever mock data was provided on initialization.
func (r *Requester) Submit(ctx context.Context, resolverName resolution.ResolverName, req resource.Request) (resolution.ResolvedResource, error) {
	if (r.ResolverPayload == resource.ResolverPayload{} || r.ResolverPayload.ResolutionSpec == nil || len(r.ResolverPayload.ResolutionSpec.Params) == 0) {
		return r.ResolvedResource, r.SubmitErr
	}
	if r.ResolverPayload.ResolutionSpec.URL == "" {
		return r.ResolvedResource, r.SubmitErr
	}
	reqParams := make(map[string]pipelinev1.ParamValue)
	for _, p := range req.ResolverPayload().ResolutionSpec.Params {
		reqParams[p.Name] = p.Value
	}

	var wrongParams []string
	for _, p := range r.ResolverPayload.ResolutionSpec.Params {
		if reqValue, ok := reqParams[p.Name]; !ok {
			wrongParams = append(wrongParams, fmt.Sprintf("expected %s param to be %#v, but was %#v", p.Name, p.Value, reqValue))
		} else if d := cmp.Diff(p.Value, reqValue); d != "" {
			wrongParams = append(wrongParams, fmt.Sprintf("%s param did not match: %s", p.Name, diff.PrintWantGot(d)))
		}
	}
	if len(wrongParams) > 0 {
		return nil, errors.New(strings.Join(wrongParams, "; "))
	}
	if r.ResolverPayload.ResolutionSpec.URL != req.ResolverPayload().ResolutionSpec.URL {
		return nil, fmt.Errorf("Resolution name did not match. Got %s; Want %s", req.ResolverPayload().ResolutionSpec.URL, r.ResolverPayload.ResolutionSpec.URL)
	}

	return r.ResolvedResource, r.SubmitErr
}

// ResolvedResource implements resolution.ResolvedResource and makes
// it easier to mock the resolved content of a fetched pipeline or task.
type ResolvedResource struct {
	// The resolved bytes to return when resolution is complete.
	ResolvedData []byte
	// An error to return instead of the resolved bytes after
	// resolution completes.
	DataErr error
	// Annotations to return when resolution is complete.
	ResolvedAnnotations map[string]string
	// ResolvedRefSource to return the source reference of the remote data
	ResolvedRefSource *pipelinev1.RefSource
}

// Data implements resolution.ResolvedResource and returns the mock
// data and/or error given to it on initialization.
func (r *ResolvedResource) Data() ([]byte, error) {
	return r.ResolvedData, r.DataErr
}

// Annotations implements resolution.ResolvedResource and returns
// the mock annotations given to it on initialization.
func (r *ResolvedResource) Annotations() map[string]string {
	return r.ResolvedAnnotations
}

// RefSource is the source reference of the remote data that records where the remote
// file came from including the url, digest and the entrypoint.
func (r *ResolvedResource) RefSource() *pipelinev1.RefSource {
	return r.ResolvedRefSource
}

// RawRequest stores the raw request data
type RawRequest struct {
	ResolverPayload resource.ResolverPayload
}

// Request returns a Request interface based on the RawRequest.
func (r *RawRequest) Request() resource.Request {
	if r == nil {
		r = &RawRequest{}
	}
	return &Request{
		RawRequest: *r,
	}
}

// Request implements resolution.Request and makes it easier to mock input for submit
// Using inline structs is to avoid conflicts between field names and method names.
type Request struct {
	RawRequest
}

var _ resource.Request = &Request{}

// NewRequest creates a mock request that is populated with the given name namespace and params
func NewRequest(resolverPayload resource.ResolverPayload) *Request {
	return &Request{
		RawRequest: RawRequest{
			ResolverPayload: resolverPayload,
		},
	}
}

// Params implements resolution.Request and returns the mock params given to it on initialization.
func (r *Request) ResolverPayload() resource.ResolverPayload {
	return r.RawRequest.ResolverPayload
}

var _ resource.Request = &Request{}
