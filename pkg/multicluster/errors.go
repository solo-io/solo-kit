package multicluster

import (
	"github.com/solo-io/go-utils/errors"
)

var (
	NoClientForClusterError = func(resourceName, cluster string) error {
		return errors.Errorf("%v client does not exist for cluster %v", resourceName, cluster)
	}
)
