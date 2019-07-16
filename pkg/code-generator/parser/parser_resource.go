package parser

import (
	"sort"
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/iancoleman/strcase"
	"github.com/ilackarms/protokit"
	"github.com/solo-io/go-utils/log"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/code-generator/model"
	"github.com/solo-io/solo-kit/pkg/errors"
)

const (
	// solo-kit types
	// required fields
	metadataTypeName = ".core.solo.io.Metadata"
	statusTypeName   = ".core.solo.io.Status"

	// magic comments
	// Deprecated, use Message Option (core.solo.io.resource).short_name
	shortNameDeclaration = "@solo-kit:resource.short_name="
	// Deprecated, use Message Option (core.solo.io.resource).plural_name
	pluralNameDeclaration = "@solo-kit:resource.plural_name="
	// Deprecated, use projectConfig.ResourceGroups
	resourceGroupsDeclaration = "@solo-kit:resource.resource_groups="
)

// add some data we need to the regular proto message
type ProtoMessageWrapper struct {
	GoPackage string
	Message   *protokit.Descriptor
}

func getResource(resources []*model.Resource, cfg model.ResourceConfig) (*model.Resource, error) {
	matches := func(res *model.Resource) bool {
		return res.Name == cfg.ResourceName //&& (res.ProtoPackage == cfg.ResourcePackage || res.GoPackage == cfg.ResourcePackage)
	}

	// collect all resources that match on package and name
	var possibleResources []*model.Resource
	for _, res := range resources {
		if matches(res) {
			possibleResources = append(possibleResources, res)
		}
	}
	switch len(possibleResources) {
	case 1:
		return possibleResources[0], nil
	case 0:
		return nil, errors.Errorf("getting resource: message %v not found", cfg)
	}

	return possibleResources[0], nil
}

func getResources(version *model.Version, apiGroup *model.ApiGroup, messages []ProtoMessageWrapper) ([]*model.Resource, error) {
	var (
		resources []*model.Resource
	)
	for _, msg := range messages {
		resource, err := describeResource(msg)
		if err != nil {
			return nil, err
		}
		if resource == nil {
			// not a solo-kit resource, ignore
			continue
		}
		for _, vc := range apiGroup.VersionConfigs {
			if vc.IsOurProto(resource.Filename) {
				resource.Version = vc.Version
				break
			}
		}
		resource.Project = version
		resources = append(resources, resource)
	}

	for _, custom := range version.VersionConfig.CustomResources {
		impPrefix := strings.Replace(custom.Package, "/", "_", -1)
		impPrefix = strings.Replace(impPrefix, ".", "_", -1)
		impPrefix = strings.Replace(impPrefix, "-", "_", -1)
		resources = append(resources, &model.Resource{
			Name:               custom.Type,
			ShortName:          custom.ShortName,
			PluralName:         custom.PluralName,
			GoPackage:          custom.Package,
			ClusterScoped:      custom.ClusterScoped,
			CustomImportPrefix: impPrefix,
			SkipDocsGen:        true,
			Project:            version,
			IsCustom:           true,
			CustomResource:     custom,
		})
	}

	return resources, nil
}

func GetResourceGroups(apiGroup *model.ApiGroup, resources []*model.Resource) ([]*model.ResourceGroup, error) {
	var (
		resourceGroups []*model.ResourceGroup
	)

	for groupName, resourcesCfg := range apiGroup.ResourceGroups {
		var resourcesForGroup []*model.Resource
		for _, resourceCfg := range resourcesCfg {
			resource, err := getResource(resources, resourceCfg)
			if err != nil {
				return nil, err
			}

			var importPrefix string
			if !apiGroup.IsOurProto(resource.Filename) && !resource.IsCustom {
				importPrefix = resource.ProtoPackage
			} else if resource.IsCustom && resource.CustomResource.Imported {
				// If is custom resource from a different version use import prefix
				importPrefix = resource.CustomImportPrefix
			}

			if importPrefix != "" {
				resource.ImportPrefix = strings.Replace(importPrefix, ".", "_", -1) + "."
			}
			resourcesForGroup = append(resourcesForGroup, resource)
		}

		log.Printf("creating resource group: %v", groupName)
		rg := &model.ResourceGroup{
			Name:      groupName,
			GoName:    goName(groupName),
			ApiGroup:  apiGroup,
			Resources: resourcesForGroup,
		}
		for _, res := range resourcesForGroup {
			res.ResourceGroups = append(res.ResourceGroups, rg)
		}

		imports := make(map[string]string)
		for _, res := range rg.Resources {
			// only generate files for the resources in our group, otherwise we import
			if res.GoPackage != rg.ApiGroup.ResourceGroupGoPackage {
				// add import
				imports[strings.TrimSuffix(res.ImportPrefix, ".")] = res.GoPackage
			}
		}
		var sortedImports []string
		for k := range imports {
			sortedImports = append(sortedImports, k)
		}
		sort.Strings(sortedImports)
		for _, imp := range sortedImports {
			rg.Imports += imp + " \"" + imports[imp] + "\"\n\t"
		}

		resourceGroups = append(resourceGroups, rg)
	}

	// sort for stability
	for _, res := range resources {
		sort.SliceStable(res.ResourceGroups, func(i, j int) bool {
			return res.ResourceGroups[i].Name < res.ResourceGroups[j].Name
		})
	}
	sort.SliceStable(resources, func(i, j int) bool {
		return resources[i].Name < resources[j].Name
	})
	sort.SliceStable(resourceGroups, func(i, j int) bool {
		return resourceGroups[i].Name < resourceGroups[j].Name
	})
	return resourceGroups, nil
}

func describeResource(messageWrapper ProtoMessageWrapper) (*model.Resource, error) {
	msg := messageWrapper.Message
	// not a solo kit resource, or you messed up!
	if !hasField(msg, "metadata", metadataTypeName) {
		return nil, nil
	}

	comments := strings.Split(msg.GetComments().Leading, "\n")

	name := msg.GetName()
	var (
		shortName, pluralName      string
		clusterScoped, skipDocsGen bool
	)
	resourceOpts, err := proto.GetExtension(msg.Options, core.E_Resource)
	if err != nil {
		log.Warnf("failed to get solo-kit message options for resource %v: %v", msg.GetName(), err)
		log.Warnf("use of magic comments is deprecated, use Message Option (core.solo.io.resource)")
		// required flags
		sn, ok := getCommentValue(comments, shortNameDeclaration)
		if !ok {
			return nil, errors.Errorf("must provide %s", shortNameDeclaration)
		}
		shortName = sn
		pn, ok := getCommentValue(comments, pluralNameDeclaration)
		if !ok {
			return nil, errors.Errorf("must provide %s", pluralNameDeclaration)
		}
		pluralName = pn
	} else {
		res, ok := resourceOpts.(*core.Resource)
		if !ok {
			return nil, errors.Errorf("internal error: message options were not type *core.Resource: %+v", resourceOpts)
		}
		shortName = res.ShortName
		pluralName = res.PluralName
		clusterScoped = res.ClusterScoped
		skipDocsGen = res.SkipDocsGen
	}

	// always make it upper camel
	pluralName = strcase.ToCamel(pluralName)

	hasStatus := hasField(msg, "status", statusTypeName)

	fields := collectFields(msg)
	oneofs := collectOneofs(msg)

	return &model.Resource{
		Name:          name,
		ProtoPackage:  msg.GetPackage(),
		GoPackage:     messageWrapper.GoPackage,
		ShortName:     shortName,
		PluralName:    pluralName,
		HasStatus:     hasStatus,
		Fields:        fields,
		Oneofs:        oneofs,
		ClusterScoped: clusterScoped,
		SkipDocsGen:   skipDocsGen,
		Filename:      msg.GetFile().GetName(),
		Original:      msg,
	}, nil
}
