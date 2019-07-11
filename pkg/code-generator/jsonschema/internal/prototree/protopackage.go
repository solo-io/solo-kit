package prototree

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"github.com/solo-io/go-utils/contextutils"
)

var (
	logger = contextutils.LoggerFrom(context.TODO())
)

// ProtoPackage describes a package of Protobuf, which is an container of message types.
type ProtoPackage struct {
	name     string
	parent   *ProtoPackage
	children map[string]*ProtoPackage
	types    map[string]*descriptor.DescriptorProto

	sync.RWMutex
}

func NewProtoTree() *ProtoPackage {
	return &ProtoPackage{
		name:     "",
		parent:   nil,
		children: make(map[string]*ProtoPackage),
		types:    make(map[string]*descriptor.DescriptorProto),
	}
}

func (pkg *ProtoPackage) GetName() string {
	return pkg.name
}

func (pkg *ProtoPackage) reset() {
	pkg.Lock()
	defer pkg.Unlock()
	pkg.children = make(map[string]*ProtoPackage)
	pkg.types = make(map[string]*descriptor.DescriptorProto)
}

func (pkg *ProtoPackage) RegisterMessage(pkgName *string, msg *descriptor.DescriptorProto) {
	pkg.registerType(pkgName, msg)
	for _, v := range msg.GetNestedType() {
		pkg.RegisterMessage(proto.String(fmt.Sprintf("%s.%s", *pkgName, msg.GetName())), v)
	}
}

func (pkg *ProtoPackage) registerType(pkgName *string, msg *descriptor.DescriptorProto) {
	pkg.Lock()
	defer pkg.Unlock()
	if pkgName != nil {
		for _, node := range strings.Split(*pkgName, ".") {
			child, ok := pkg.children[node]
			if !ok {
				child = &ProtoPackage{
					name:     pkg.name + "." + node,
					parent:   pkg,
					children: make(map[string]*ProtoPackage),
					types:    make(map[string]*descriptor.DescriptorProto),
				}
				if pkg.name == "" {
					child.name = node
				}
				pkg.children[node] = child
			}
			pkg = child
		}
	}
	pkg.types[msg.GetName()] = msg
}

func (pkg *ProtoPackage) LookupType(name string) (*descriptor.DescriptorProto, bool) {
	pkg.RLock()
	defer pkg.RUnlock()
	if strings.HasPrefix(name, ".") {
		return pkg.relativelyLookupType(name[1:])
	}

	for ; pkg != nil; pkg = pkg.parent {
		if desc, ok := pkg.relativelyLookupType(name); ok {
			return desc, ok
		}
	}
	return nil, false
}

// func (pkg *ProtoPackage) GetJsonName(name string) (string, bool) {
// 	pkg, ok := pkg.RelativelyLookupPackage(name)
// 	if !ok {
// 		return "", false
// 	}
// 	if pkg.parent == nil {
// 		return "", false
// 	}
// 	pkg = pkg.parent
// 	for _, child := range pkg.types {
// 		if child.GetName() == name {
// 			return child
// 		}
// 	}
// }

func relativelyLookupNestedType(desc *descriptor.DescriptorProto, name string) (*descriptor.DescriptorProto, bool) {
	components := strings.Split(name, ".")
componentLoop:
	for _, component := range components {
		for _, nested := range desc.GetNestedType() {
			if nested.GetName() == component {
				desc = nested
				continue componentLoop
			}
		}
		logger.Infof("no such nested message %s in %s", component, desc.GetName())
		return nil, false
	}
	return desc, true
}

func (pkg *ProtoPackage) relativelyLookupType(name string) (*descriptor.DescriptorProto, bool) {
	components := strings.SplitN(name, ".", 2)
	switch len(components) {
	case 0:
		logger.Debugf("empty message name")
		return nil, false
	case 1:
		found, ok := pkg.types[components[0]]
		return found, ok
	case 2:
		logger.Debugf("looking for %s in %s at %s (%v)", components[1], components[0], pkg.name, pkg)
		if child, ok := pkg.children[components[0]]; ok {
			found, ok := child.relativelyLookupType(components[1])
			return found, ok
		}
		if msg, ok := pkg.types[components[0]]; ok {
			found, ok := relativelyLookupNestedType(msg, components[1])
			return found, ok
		}
		logger.Infof("no such package nor message %s in %s", components[0], pkg.name)
		return nil, false
	default:
		logger.Fatalf("not reached")
		return nil, false
	}
}

func (pkg *ProtoPackage) RelativelyLookupPackage(name string) (*ProtoPackage, bool) {
	pkg.RLock()
	defer pkg.RUnlock()
	components := strings.Split(name, ".")
	for _, c := range components {
		var ok bool
		pkg, ok = pkg.children[c]
		if !ok {
			return nil, false
		}
	}
	return pkg, true
}
