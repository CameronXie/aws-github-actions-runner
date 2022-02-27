package eks

import (
	"context"
	"encoding/base64"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/aws-iam-authenticator/pkg/token"
)

const (
	RunnerType = "eks"
)

type KubeClientFactory func(c *rest.Config) (*kubernetes.Clientset, error)

func GetKubeClient(
	ctx context.Context,
	cluster string,
	client eks.DescribeClusterAPIClient,
	tokenGenerator token.Generator,
	kubeFactory KubeClientFactory,
) (kubernetes.Interface, error) {
	if tokenGenerator == nil {
		g, _ := token.NewGenerator(false, false)
		tokenGenerator = g
	}

	if kubeFactory == nil {
		kubeFactory = kubernetes.NewForConfig
	}

	res, err := client.DescribeCluster(ctx, &eks.DescribeClusterInput{Name: aws.String(cluster)})

	if err != nil {
		return nil, err
	}

	ca, _ := base64.StdEncoding.DecodeString(*res.Cluster.CertificateAuthority.Data)
	tk, _ := tokenGenerator.Get(cluster)

	return kubeFactory(&rest.Config{
		Host:        aws.ToString(res.Cluster.Endpoint),
		BearerToken: tk.Token,
		TLSClientConfig: rest.TLSClientConfig{
			CAData: ca,
		},
	})
}

func uint64ToString(n uint64) string {
	base := 10
	return strconv.FormatUint(n, base)
}
