package helm

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"

	helmaction "helm.sh/helm/v3/pkg/action"
	helmloader "helm.sh/helm/v3/pkg/chart/loader"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var helmLogger = logf.Log.WithName("helm")

const helmChartsPath string = "/helm-charts"

// InstallHelmChart installs the given Helm chart with values
func InstallHelmChart(namespace string, chartName string, releaseName string, vals map[string]interface{}) error {
	// Load Helm chart
	chartPath := helmChartsPath + "/" + chartName
	chart, err := helmloader.Load(chartPath)
	if err != nil {
		helmLogger.Error(err, "Failed to load Helm chart at", "ChartPath", chartPath)
		return err
	}

	// Install Helm chart
	actionConfig, err := NewHelmConfiguration(namespace)
	if err != nil {
		return err
	}
	install := helmaction.NewInstall(actionConfig)
	install.Namespace = namespace
	install.ReleaseName = releaseName
	install.Wait = true
	release, err := install.Run(chart, vals)
	if err != nil {
		helmLogger.Error(err, "Failed to install Helm chart", "ChartName", chartName)
		return err
	}

	helmLogger.Info("Successfully create Helm chart", "ChartName", release.Chart.Metadata.Name, "ReleaseName", release.Name)

	return nil
}

// UninstallHelmChart uninstalls the given Helm chart name
func UninstallHelmChart(namespace string, releaseName string) error {
	// Uninstall Helm chart
	actionConfig, err := NewHelmConfiguration(namespace)
	if err != nil {
		return err
	}
	uninstall := helmaction.NewUninstall(actionConfig)
	response, err := uninstall.Run(releaseName)
	if err != nil {
		helmLogger.Error(err, "Failed to uninstall Helm release", "ReleaseName", releaseName)
		return err
	}

	helmLogger.Info("Successfully uninstall Helm release", "ReleaseName", response.Release.Name)

	return nil
}

// NewHelmConfiguration creates a new Helm Configuration object under the given namespace
func NewHelmConfiguration(namespace string) (*helmaction.Configuration, error) {
	const (
		tokenFile  = "/var/run/secrets/kubernetes.io/serviceaccount/token"
		rootCAFile = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
	)

	// Create Kubernetes client config to access the api from within a pod
	// https://kubernetes.io/docs/tasks/administer-cluster/access-cluster-api/#accessing-the-api-from-within-a-pod
	serviceHost, servicePort := os.Getenv("KUBERNETES_SERVICE_HOST"), os.Getenv("KUBERNETES_SERVICE_PORT")
	apiServer := "https://" + net.JoinHostPort(serviceHost, servicePort)
	tokenBuf, err := ioutil.ReadFile(tokenFile)
	if err != nil {
		helmLogger.Error(err, "Failed to read Kubernetes token file")
		return nil, err
	}
	bearerToken := string(tokenBuf)
	caFile := rootCAFile

	cf := genericclioptions.NewConfigFlags(true)
	cf.Namespace = &namespace
	cf.APIServer = &apiServer
	cf.BearerToken = &bearerToken
	cf.CAFile = &caFile

	// Create Helm action.Configuration object
	actionConfig := new(helmaction.Configuration)
	err = actionConfig.Init(cf, namespace, "", helmDebugLog)
	if err != nil {
		helmLogger.Error(err, "Failed to get Kubernetes client config for Helm")
		return nil, err
	}

	return actionConfig, nil
}

// helmDebugLog returns a logger that writes debug strings
func helmDebugLog(format string, v ...interface{}) {
	debugMsg := fmt.Sprintf(format, v...)
	helmLogger.Info(debugMsg)
}
