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

// CopyJob creates a copy of k8s batch Job
func CopyJob(job *batch.Job) *batch.Job {
	copy := batch.Job{}
	copy.Spec.Template.Spec = job.Spec.Template.Spec

	genName := "kron-" + job.GetName() + "-"

	copy.ObjectMeta.SetGenerateName(genName)
	copy.ObjectMeta.SetLabels(map[string]string{
		"origin":   "kron",
		"template": job.GetName(),
	})

	return &copy
}

// ListJobs finds all job templates
func ListJobs(k *client.Client, namespace string) (jobs *batch.JobList, err error) {
	kronSelector, err := labels.Parse("update=true")
	if err != nil {
		return
	}

	opts := api.ListOptions{}
	opts.LabelSelector = kronSelector
	jobs, err = k.Batch().Jobs(namespace).List(opts)

	return
}
