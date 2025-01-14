package httpserver

import (
	"context"
	"fmt"
	"strconv"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"

	"github.com/kubeedge/kubeedge/cloud/pkg/common/client"
	"github.com/kubeedge/kubeedge/common/constants"
)

const (
	TokenSecretName      string = "tokensecret"
	TokenDataName        string = "tokendata"
	CaSecretName         string = "casecret"
	CloudCoreSecretName  string = "cloudcoresecret"
	CaDataName           string = "cadata"
	CaKeyDataName        string = "cakeydata"
	CloudCoreCertName    string = "cloudcoredata"
	CloudCoreKeyDataName string = "cloudcorekeydata"
)

func GetSecret(secretName string, ns string) (*corev1.Secret, error) {
	cli := client.GetKubeClient()
	return cli.CoreV1().Secrets(ns).Get(context.Background(), secretName, metav1.GetOptions{})
}

// CreateSecret creates a secret
func CreateSecret(secret *corev1.Secret, ns string) error {
	cli := client.GetKubeClient()
	if err := CreateNamespaceIfNeeded(cli, ns); err != nil {
		return fmt.Errorf("failed to create Namespace kubeedge, error: %v", err)
	}
	if _, err := cli.CoreV1().Secrets(ns).Create(context.Background(), secret, metav1.CreateOptions{}); err != nil {
		if errors.IsAlreadyExists(err) {
			if secret.ObjectMeta.Name == TokenSecretName {
				klog.Infof("Never update token secret if exist")
				return nil
			}
			if _, err := cli.CoreV1().Secrets(ns).Update(context.Background(), secret, metav1.UpdateOptions{}); err != nil {
				return fmt.Errorf("failed to update the secret, namespace: %s, name: %s, err: %v", ns, secret.Name, err)
			}
		} else {
			return fmt.Errorf("failed to create the secret, namespace: %s, name: %s, err: %v", ns, secret.Name, err)
		}
	}
	return nil
}

func CreateTokenSecret(caHashAndToken []byte, cloudcoreID string) error {
	token := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      TokenSecretName,
			Namespace: constants.SystemNamespace,
			Labels: map[string]string{
				"updatedTime": strconv.Itoa(int(time.Now().Unix())),
				"updatedBy":   cloudcoreID,
			},
		},
		Data: map[string][]byte{
			TokenDataName: caHashAndToken,
		},
		StringData: map[string]string{},
		Type:       "Opaque",
	}
	return CreateSecret(token, constants.SystemNamespace)
}

func CreateCaSecret(certDER, key []byte) error {
	caSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      CaSecretName,
			Namespace: constants.SystemNamespace,
		},
		Data: map[string][]byte{
			CaDataName:    certDER,
			CaKeyDataName: key,
		},
		StringData: map[string]string{},
		Type:       "Opaque",
	}
	return CreateSecret(caSecret, constants.SystemNamespace)
}

func CreateCloudCoreSecret(certDER, key []byte) error {
	cloudCoreCert := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      CloudCoreSecretName,
			Namespace: constants.SystemNamespace,
		},
		Data: map[string][]byte{
			CloudCoreCertName:    certDER,
			CloudCoreKeyDataName: key,
		},
		StringData: map[string]string{},
		Type:       "Opaque",
	}
	return CreateSecret(cloudCoreCert, constants.SystemNamespace)
}

func CreateNamespaceIfNeeded(cli kubernetes.Interface, ns string) error {
	c := cli.CoreV1()
	if _, err := c.Namespaces().Get(context.Background(), ns, metav1.GetOptions{}); err == nil {
		return nil
	}
	newNs := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: ns,
		},
	}
	_, err := c.Namespaces().Create(context.Background(), newNs, metav1.CreateOptions{})
	if err != nil && errors.IsAlreadyExists(err) {
		err = nil
	}
	return err
}
