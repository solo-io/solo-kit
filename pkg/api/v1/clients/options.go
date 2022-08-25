package clients

import (
	"k8s.io/apimachinery/pkg/labels"
)

func GetLabelSelector(listOpts ListOpts) (labels.Selector, error) {
	// https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#set-based-requirement
	if listOpts.ExpressionSelector != "" {
		return labels.Parse(listOpts.ExpressionSelector)
	}

	// https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#equality-based-requirement
	return labels.SelectorFromSet(listOpts.Selector), nil
}
