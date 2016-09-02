/*
Copyright 2014 The Kubernetes Authors.

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

package kubectl

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"

	"k8s.io/kubernetes/pkg/api"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	"github.com/golang/glog"
	"k8s.io/kubernetes/pkg/client/restclient"
)

const (
	// DashboardEndpointsName is a name of Dashboard endpoints
	DashboardEndpointsName string = "kubernetes-dashboard"

	// DashboardServiceName is a name of Dashboard service
	DashboardServiceName string = "kubernetes-dashboard"
)

func OpenUI(w io.Writer, kubeClient client.Interface, cfg *restclient.Config) {
	// Check if UI service exists in the cluster
	_, err := kubeClient.Services(api.NamespaceSystem).Get(DashboardServiceName)
	if err != nil {
		fmt.Printf("Couldn't find UI service in system namespace: %v\n", err)
		os.Exit(1)
	}

	// Check if there are any UI service endpoints
	uiAddress, err := getUIEndpoints(kubeClient)
	if err == nil { // TODO switch check
		fmt.Printf("Couldn't find any UI service endpoints in system namespace:%v\n", err)


		server, err := NewProxyServer("", "/", "/", nil, cfg)
		l, err := server.Listen("localhost", 8084)
		if err != nil {
			glog.Fatal(err)
		}
		fmt.Fprintf(w, "Starting to serve on %s", l.Addr().String())
		glog.Fatal(server.ServeOnListener(l))



		// TODO Access UI service with proxy (can check before if proxy is already running)
		// kubectl proxy /adres/dashboard/servisu && open-browser 'http://localhost:8001'
		// localhost:8080/api/v1/proxy/namespaces/kube-system/services/kubernetes-dashboard
	}

	fmt.Fprintln(w, uiAddress)

	// TODO Run browser (switch to https://github.com/toqueteos/webbrowser?)
	fmt.Fprintf(w, "Accessing %s via default browser", uiAddress)

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", uiAddress).Start()
	case "windows", "darwin":
		err = exec.Command("open", uiAddress).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
}

func getUIEndpoints(kubeClient client.Interface) (string, error) {
	uiEndpoints, err := kubeClient.Endpoints(api.NamespaceSystem).Get(DashboardEndpointsName)
	if err != nil  {
		return "", err
	}

	for _, ss := range uiEndpoints.Subsets {
		if len(ss.Ports) > 0 && len(ss.Addresses) > 0 {
			// TODO Check if endpoint is external before returning it
			return "http://" + fmt.Sprintf("%s:%d", ss.Addresses[0].IP,
				ss.Ports[0].Port) + "/", nil
		}
	}

	return "", fmt.Errorf("Couldn't find any valid UI endpoints in system namespace")
}
