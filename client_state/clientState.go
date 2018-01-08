/*
Copyright 2016 Iguazio Systems Ltd.

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
package client_state

import (
	"github.com/yaronha/kube-crd/crd_state"

	"fmt"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

// This file implement all the (CRUD) client methods we need to access our CRD object

func CrdClientState(cl *rest.RESTClient, scheme *runtime.Scheme, namespace string) *crdclientState {
	return &crdclientState{cl: cl, ns: namespace, plural: crd_state.CRDPlural,
		codec: runtime.NewParameterCodec(scheme)}
}

type crdclientState struct {
	cl     *rest.RESTClient
	ns     string
	plural string
	codec  runtime.ParameterCodec
}

func (f *crdclientState) Create(obj *crd_state.ExampleState) (*crd_state.ExampleState, error) {
	var result crd_state.ExampleState
	fmt.Println("In Create  call")
	err := f.cl.Post().
		Namespace(f.ns).Resource(f.plural).
		Body(obj).Do().Into(&result)
	return &result, err
}

func (f *crdclientState) Update(obj *crd_state.ExampleState) (*crd_state.ExampleState, error) {
	var result crd_state.ExampleState
	err := f.cl.Put().
		Namespace(f.ns).Resource(f.plural).
		Body(obj).Do().Into(&result)
	return &result, err
}

func (f *crdclientState) Delete(name string, options *meta_v1.DeleteOptions) error {
	return f.cl.Delete().
		Namespace(f.ns).Resource(f.plural).
		Name(name).Body(options).Do().
		Error()
}

func (f *crdclientState) Get(name string) (*crd_state.ExampleState, error) {
	var result crd_state.ExampleState
	err := f.cl.Get().
		Namespace(f.ns).Resource(f.plural).
		Name(name).Do().Into(&result)
	return &result, err
}

func (f *crdclientState) List(opts meta_v1.ListOptions) (*crd_state.ExampleStateList, error) {
	var result crd_state.ExampleStateList
	err := f.cl.Get().
		Namespace(f.ns).Resource(f.plural).
		VersionedParams(&opts, f.codec).
		Do().Into(&result)
	return &result, err
}

// Create a new List watch for our TPR
func (f *crdclientState) NewListWatch() *cache.ListWatch {
	return cache.NewListWatchFromClient(f.cl, f.plural, f.ns, fields.Everything())
}
