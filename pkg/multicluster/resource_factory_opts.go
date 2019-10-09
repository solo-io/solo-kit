package multicluster

import "time"

type KubeResourceFactoryOpts struct {
	SkipCrdCreation    bool
	NamespaceWhitelist []string
	ResyncPeriod       time.Duration
}
