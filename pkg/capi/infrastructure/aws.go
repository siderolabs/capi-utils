// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package infrastructure

import (
	"context"
	"fmt"
	"os"

	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// AWSProviderName is the string id of the AWS provider.
const AWSProviderName = "aws"

// NewAWSProvider creates new AWS infrastructure provider.
func NewAWSProvider(creds, version string) *AWSProvider {
	return &AWSProvider{
		B64EncodedCredentials: creds,
		Version:               version,
	}
}

// AWSProvider infrastructure provider.
type AWSProvider struct {
	B64EncodedCredentials string
	Version               string
}

// Name implements Provider interface.
func (s *AWSProvider) Name() string {
	if s.Version != "" {
		return fmt.Sprintf("%s:%s", AWSProviderName, s.Version)
	}

	return AWSProviderName
}

// Env implements Provider interface.
func (s *AWSProvider) Env() error {
	vars := map[string]string{
		"AWS_B64ENCODED_CREDENTIALS": s.B64EncodedCredentials,
	}

	for key, value := range vars {
		if value != "" {
			err := os.Setenv(key, value)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// IsInstalled implements Provider interface.
func (s *AWSProvider) IsInstalled(ctx context.Context, clientset *kubernetes.Clientset) (bool, error) {
	_, err := clientset.CoreV1().Namespaces().Get(ctx, "capa-system", v1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}
