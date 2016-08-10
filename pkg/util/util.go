// Copyright © 2016 Sabaka OU <hello@sabaka.io>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package util

import (
	"fmt"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/batch"
	"k8s.io/kubernetes/pkg/client/restclient"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/labels"
)

// CreateClient creates a client for Kubernetes cluster
func CreateClient(host string) (k *client.Client, err error) {
	var config *restclient.Config

	if len(host) == 0 {
		config, err = restclient.InClusterConfig()
		if err != nil {
			return k, err
		}
	} else {
		config = &restclient.Config{
			Host: host,
		}
	}

	k, err = client.New(config)
	return
}

// DeletePodsInJob deletes all Pods which were created for a Job
func DeletePodsInJob(k *client.Client, job *batch.Job) (err error) {
	namespace := job.Namespace
	deleteOpts := api.DeleteOptions{}
	listOpts := api.ListOptions{}
	uid := job.GetObjectMeta().GetUID()
	label := "controller-uid=" + fmt.Sprintf("%s", uid)

	selector, err := labels.Parse(label)
	if err != nil {
		return
	}

	listOpts.LabelSelector = selector
	pods, err := k.Pods(namespace).List(listOpts)
	if err != nil {
		return
	}

	for _, pod := range pods.Items {
		if e := k.Pods(namespace).Delete(pod.GetName(), &deleteOpts); e != nil {
			err = e
			return
		}
	}

	return
}
