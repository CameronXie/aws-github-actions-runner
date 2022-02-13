package eks

import (
	"context"
	"encoding/base64"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/stretchr/testify/assert"
	appv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	appsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/aws-iam-authenticator/pkg/token"
)

func TestGetKubeClient(t *testing.T) {
	ca, tk, endpoint := "ca", "token", "endpoint"
	cases := map[string]struct {
		describeClusterErr           error
		expectedDescribeClusterInput *eks.DescribeClusterInput
		expectedRestConfig           *rest.Config
		err                          error
	}{
		"new kube client": {
			expectedDescribeClusterInput: &eks.DescribeClusterInput{Name: aws.String("cluster")},
			expectedRestConfig: &rest.Config{
				Host:        endpoint,
				BearerToken: tk,
				TLSClientConfig: rest.TLSClientConfig{
					CAData: []byte(ca),
				},
			},
		},
		"describe cluster error": {
			describeClusterErr:           errors.New("describe cluster error"),
			expectedDescribeClusterInput: &eks.DescribeClusterInput{Name: aws.String("cluster")},
			err:                          errors.New("describe cluster error"),
		},
	}

	for n, tc := range cases {
		t.Run(n, func(t *testing.T) {
			a := assert.New(t)
			eksClient := &mockedDescribeClusterClient{
				endpoint:           endpoint,
				ca:                 base64.StdEncoding.EncodeToString([]byte(ca)),
				describeClusterErr: tc.describeClusterErr,
			}
			tg := &mockedTokenGenerator{
				token: token.Token{
					Token: tk,
				},
			}

			kubeFactory := new(mockedKubeClientFactory)

			_, err := GetKubeClient(context.TODO(), "cluster", eksClient, tg, kubeFactory.NewForConfig)

			a.Equal(tc.err, err)
			a.Equal(tc.expectedDescribeClusterInput, eksClient.input)
			a.Equal(tc.expectedRestConfig, kubeFactory.config)
		})
	}
}

type mockedDescribeClusterClient struct {
	input              *eks.DescribeClusterInput
	endpoint           string
	ca                 string
	describeClusterErr error
}

func (m *mockedDescribeClusterClient) DescribeCluster(
	_ context.Context,
	input *eks.DescribeClusterInput,
	_ ...func(*eks.Options),
) (*eks.DescribeClusterOutput, error) {
	m.input = input
	return &eks.DescribeClusterOutput{
		Cluster: &types.Cluster{
			Endpoint:             aws.String(m.endpoint),
			CertificateAuthority: &types.Certificate{Data: aws.String(m.ca)},
		},
	}, m.describeClusterErr
}

type mockedTokenGenerator struct {
	token.Generator
	token token.Token
}

func (m *mockedTokenGenerator) Get(_ string) (token.Token, error) {
	return m.token, nil
}

type mockedKubeClient struct {
	kubernetes.Interface
	appsv1.AppsV1Interface
	appsv1.DeploymentInterface

	namespace            string
	deployment           *appv1.Deployment
	deleteDeploymentName string
	deleteOpt            metav1.DeleteOptions

	deploymentErr error
}

func (m *mockedKubeClient) AppsV1() appsv1.AppsV1Interface {
	return m
}

func (m *mockedKubeClient) Deployments(namespace string) appsv1.DeploymentInterface {
	m.namespace = namespace
	return m
}

func (m *mockedKubeClient) Create(
	_ context.Context,
	deployment *appv1.Deployment,
	// nolint:gocritic
	_ metav1.CreateOptions,
) (*appv1.Deployment, error) {
	m.deployment = deployment
	return nil, m.deploymentErr
}

// nolint:gocritic
func (m *mockedKubeClient) Delete(_ context.Context, name string, opts metav1.DeleteOptions) error {
	m.deleteDeploymentName = name
	m.deleteOpt = opts

	return m.deploymentErr
}

type mockedKubeClientFactory struct {
	config *rest.Config
}

func (m *mockedKubeClientFactory) NewForConfig(c *rest.Config) (*kubernetes.Clientset, error) {
	m.config = c
	return nil, nil
}
