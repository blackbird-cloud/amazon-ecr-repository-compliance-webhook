module github.com/aws-samples/amazon-ecr-repository-compliance-webhook

go 1.14

replace github.com/aws-samples/amazon-ecr-repository-compliance-webhook/pkg/webhook => ./pkg/webhook

replace github.com/aws-samples/amazon-ecr-repository-compliance-webhook/pkg/function => ./pkg/function

require (
	github.com/aws/aws-lambda-go v1.16.0
	github.com/aws/aws-sdk-go v1.30.26
	github.com/sirupsen/logrus v1.6.0
	github.com/stretchr/objx v0.1.1 // indirect
	github.com/stretchr/testify v1.7.0
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	k8s.io/api v0.25.0
	k8s.io/apimachinery v0.25.0
)
