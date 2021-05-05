package adapter

import (
	"os"

	"github.com/layer5io/meshkit/models"
	mesherykube "github.com/layer5io/meshkit/utils/kubernetes"
	yaml "gopkg.in/yaml.v2"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// Instantiates clients used in deploying and managing mesh instances, e.g. Kubernetes clients.
// This needs to be called before applying operations.
func (h *Adapter) CreateInstance(kubeconfig []byte, contextName string, ch *chan interface{}) error {
	err := h.validateKubeconfig(kubeconfig)
	if err != nil {
		return ErrCreateInstance(err)
	}

	err = h.createKubeconfig(kubeconfig)
	if err != nil {
		return ErrCreateInstance(err)
	}

	h.MesheryKubeclient, err = mesherykube.New(kubeconfig)
	if err != nil {
		return ErrClientSet(err)
	}

	h.DynamicKubeClient = h.MesheryKubeclient.DynamicKubeClient
	h.RestConfig = h.MesheryKubeclient.RestConfig

	h.KubeClient, err = kubernetes.NewForConfig(&h.RestConfig)
	if err != nil {
		return ErrClientSet(err)
	}

	h.ClientcmdConfig.CurrentContext = contextName
	h.Channel = ch

	return nil
}

func (h *Adapter) validateKubeconfig(kubeconfig []byte) error {
	var err error
	h.ClientcmdConfig, err = clientcmd.Load(kubeconfig)
	if err != nil {
		return ErrValidateKubeconfig(err)
	}

	// If kubeconfig provided to validate function is empty
	// and the service is deployed within k8s then skip the validation
	if len(kubeconfig) == 0 && isDeployedWithinK8s() {
		return nil
	}

	if err := filterK8sConfigAuthInfos(h.ClientcmdConfig.AuthInfos); err != nil {
		return ErrValidateKubeconfig(err)
	}

	if err := clientcmdapi.FlattenConfig(h.ClientcmdConfig); err != nil {
		return ErrValidateKubeconfig(err)
	}

	if err := clientcmdapi.MinifyConfig(h.ClientcmdConfig); err != nil {
		return ErrValidateKubeconfig(err)
	}

	return nil
}

func (h *Adapter) createKubeconfig(kubeconfig []byte) error {
	kconfig := models.Kubeconfig{}
	err := yaml.Unmarshal(kubeconfig, &kconfig)
	if err != nil {
		return err
	}

	// To have control over what exactly to take in on kubeconfig
	h.KubeconfigHandler.SetKey("kind", kconfig.Kind)
	h.KubeconfigHandler.SetKey("apiVersion", kconfig.APIVersion)
	h.KubeconfigHandler.SetKey("current-context", kconfig.CurrentContext)
	err = h.KubeconfigHandler.SetObject("preferences", kconfig.Preferences)
	if err != nil {
		return err
	}

	err = h.KubeconfigHandler.SetObject("clusters", kconfig.Clusters)
	if err != nil {
		return err
	}

	err = h.KubeconfigHandler.SetObject("users", kconfig.Users)
	if err != nil {
		return err
	}

	err = h.KubeconfigHandler.SetObject("contexts", kconfig.Contexts)
	if err != nil {
		return err
	}

	return nil
}

// filterK8sConfigAuthInfos takes in the authInfos map and deletes any invalid
// authInfo.
//
// An authInfo is invalid if the certificate path or the bearer token path mentioned in it is either
// invalid or is inaccessible to the adapter
//
// The function will throw an error if after filtering the authInfos it becomes
// empty which indicates that the kubeconfig cannot be used for communicating
// with the kubernetes server.
func filterK8sConfigAuthInfos(authInfos map[string]*clientcmdapi.AuthInfo) error {
	for key, authInfo := range authInfos {
		// If clientCertficateData or the bearer token is not present then proceed to check
		// the client certicate path
		if len(authInfo.ClientCertificateData) == 0 && len(authInfo.Token) == 0 && authInfo.AuthProvider == nil {
			// If the path for clientCertficate and the bearer token, both are inaccessible or invalid then delete that authinfo
			_, errCC := os.Stat(authInfo.ClientCertificate)
			_, errToken := os.Stat(authInfo.TokenFile)
			if errCC != nil && errToken != nil {
				delete(authInfos, key)
			}
		}
	}

	// In the end if the authInfos map is empty then the kubeconfig is
	// invalid and cannot be used for communicating with kubernetes
	if len(authInfos) == 0 {
		return ErrAuthInfosInvalidMsg
	}

	return nil
}

// isDeployedWithinK8s returns true if the adapter is running
// inside a kubernetes cluster
func isDeployedWithinK8s() bool {
	return os.Getenv("KUBERNETES_SERVICE_HOST") != ""
}
