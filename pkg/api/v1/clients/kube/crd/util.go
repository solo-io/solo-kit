package crd

import (
	"github.com/solo-io/solo-kit/pkg/utils/protoutils"
)

func Copy(src, dst SoloKitCrd) error {
	srcBytes, err := protoutils.MarshalBytes(src)
	if err != nil {
		return err
	}
	err = protoutils.UnmarshalBytes(srcBytes, dst)
	if err != nil {
		return err
	}
	return nil
}
