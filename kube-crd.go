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
package main

import (
	"fmt"
	"time"

	"github.com/yaronha/kube-crd/client"
	"github.com/yaronha/kube-crd/client_state"
	"github.com/yaronha/kube-crd/crd"
	"github.com/yaronha/kube-crd/crd_state"

	"flag"
	apiextcs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

// return rest config, if path not specified assume in cluster config
func GetClientConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}

func main() {

	kubeconf := flag.String("kubeconf", "admin.conf", "Path to a kube config. Only required if out-of-cluster.")
	flag.Parse()

	config, err := GetClientConfig(*kubeconf)
	if err != nil {
		panic(err.Error())
	}

	// create clientset and create our CRD, this only need to run once
	clientset, err := apiextcs.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// note: if the CRD exist our CreateCRD function is set to exit without an error
	err = crd.CreateCRD(clientset)
	if err != nil {
		panic(err)
	}

	// Wait for the CRD to be created before we use it (only needed if its a new one)
	time.Sleep(3 * time.Second)

	// Create a new clientset which include our CRD schema
	crdcs, scheme, err := crd.NewClient(config)
	if err != nil {
		panic(err)
	}

	// Create a CRD client interface
	crdclient := client.CrdClient(crdcs, scheme, "default")

	// Create a new Example object and write to k8s
	example := &crd.Example{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:   "example123",
			Labels: map[string]string{"mylabel": "test"},
		},
		Spec: crd.ExampleSpec{
			Foo: "example-text",
			Bar: true,
		},
		Status: crd.ExampleStatus{
			State:   "created",
			Message: "Created, not processed yet",
		},
	}

	result, err := crdclient.Create(example)
	if err == nil {
		fmt.Printf("CREATED: %#v\n", result)
	} else if apierrors.IsAlreadyExists(err) {
		fmt.Printf("ALREADY EXISTS: %#v\n", result)
	} else {
		panic(err)
	}

	// List all Example objects
	items, err := crdclient.List(meta_v1.ListOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Printf("List:\n%s\n", items)
	/*
		result, err = crdclient.Get("example123")
		if err != nil {
			panic(err)
		}
		fmt.Printf("Get:\n%v\n", result)

		result.Status.Message = "Hello There"

		fmt.Println("\n Result is: %v \n", result)
		up, uperr := crdclient.Update(result)
		if uperr != nil {
			panic(uperr)
		}
		fmt.Printf("Update:\n%s\n", up)

		result, err = crdclient.Get("example123")
		if err != nil {
			panic(err)
		}
		fmt.Printf("Get:\n%s\n", result)

		err = crdclient.Delete("example123", nil)
		if err != nil {
			panic(err)
		}
	*/
	// Example Controller
	// Watch for changes in Example objects and fire Add, Delete, Update callbacks
	_, controller := cache.NewInformer(
		crdclient.NewListWatch(),
		&crd.Example{},
		time.Minute*10,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				fmt.Printf("add: %s \n", obj)
			},
			DeleteFunc: func(obj interface{}) {
				fmt.Printf("delete: %s \n", obj)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				var newexample *crd.Example
				//fmt.Printf("Update old: %s \n      New: %s\n", oldObj, newObj)
				newexample = newObj.(*crd.Example)
				fmt.Printf("NewExample %s \n", newexample)
				//copyexample = newexample.DeepCopy()
				//copyexample.Status.Message = "HelloWorld"

			},
		},
	)

	stop := make(chan struct{})
	go controller.Run(stop)

	// note: if the CRD exist our CreateCRD function is set to exit without an error
	err = crd_state.CreateCRDState(clientset)
	if err != nil {
		panic(err)
	}

	// Wait for the CRD to be created before we use it (only needed if its a new one)
	time.Sleep(3 * time.Second)

	// Create a new clientset which include our CRD schema
	crdcs, scheme, err = crd_state.NewClientState(config)
	if err != nil {
		panic(err)
	}

	// Create a CRD client interface
	crdclientstate := client_state.CrdClientState(crdcs, scheme, "default")

	// Create a new Example object and write to k8s
	examplestate := &crd_state.ExampleState{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:   "examplestate123",
			Labels: map[string]string{"mylabel": "check"},
		},
		Spec: crd_state.ExampleStateSpec{
			What: "example-text",
			Up:   true,
		},
		Status: crd_state.ExampleStateStatus{
			State:   "created",
			Message: "Created, not processed yet",
		},
	}

	resultstate, errs := crdclientstate.Create(examplestate)
	if errs == nil {
		fmt.Printf("CREATED: %#v\n", resultstate)
	} else if apierrors.IsAlreadyExists(errs) {
		fmt.Printf("ALREADY EXISTS: %#v\n", resultstate)
	} else {
		panic(errs)
	}

	// List all Example objects
	itemstate, errstate := crdclientstate.List(meta_v1.ListOptions{})
	if errstate != nil {
		panic(errstate)
	}
	fmt.Printf("List:\n%s\n", itemstate)
	/*
		result, err = crdclientstate.Get("example123")
		if err != nil {
			panic(err)
		}
		fmt.Printf("Get:\n%v\n", result)

		result.Status.Message = "Hello There"

		fmt.Println("\n Result is: %v \n", result)
		up, uperr := crdclientstate.Update(result)
		if uperr != nil {
			panic(uperr)
		}
		fmt.Printf("Update:\n%s\n", up)

		result, err = crdclientstate.Get("example123")
		if err != nil {
			panic(err)
		}
		fmt.Printf("Get:\n%s\n", result)

		err = crdclientstate.Delete("example123", nil)
		if err != nil {
			panic(err)
		}
	*/
	// Example Controller
	// Watch for changes in Example objects and fire Add, Delete, Update callbacks
	_, controllerstate := cache.NewInformer(
		crdclientstate.NewListWatch(),
		&crd_state.ExampleState{},
		time.Minute*10,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				fmt.Printf("add: %s \n", obj)
			},
			DeleteFunc: func(obj interface{}) {
				fmt.Printf("delete: %s \n", obj)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				var newexample *crd_state.ExampleState
				//fmt.Printf("Update old: %s \n      New: %s\n", oldObj, newObj)
				newexample = newObj.(*crd_state.ExampleState)
				fmt.Printf("NewExample %s \n", newexample)
				//copyexample = newexample.DeepCopy()

			},
		},
	)

	stopstate := make(chan struct{})
	go controllerstate.Run(stopstate)

	// Wait forever
	select {}
}
