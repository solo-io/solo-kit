package model

func CustomResourcesIncludesResource(crds []CustomResourceConfig, resource *Resource) bool {
	for _, crd := range crds {
		if crd.Type == resource.CustomResource.Type &&
			crd.Package == resource.CustomResource.Package {
			return true
		}
	}
	return false
}
