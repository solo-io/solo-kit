package clients

import (
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"k8s.io/apimachinery/pkg/labels"
)

// GetLabelSelector will parse ExpresionSelector if present, else it selects Selector.
func GetLabelSelector(listOpts ListOpts) (labels.Selector, error) {
	// https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#set-based-requirement
	if listOpts.ExpressionSelector != "" {
		return labels.Parse(listOpts.ExpressionSelector)
	}

	// https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#equality-based-requirement
	return labels.SelectorFromSet(listOpts.Selector), nil
}

// TranslateWatchOptsIntoListOpts translates the watch options into list options
func TranslateWatchOptsIntoListOpts(wopts WatchOpts) ListOpts {
	clopts := ListOpts{Ctx: wopts.Ctx, ExpressionSelector: wopts.ExpressionSelector, Selector: wopts.Selector}
	return clopts
}

// TranslateResourceNamespaceListToListOptions translates the resource namespace list options to List Options
func TranslateResourceNamespaceListToListOptions(lopts resources.ResourceNamespaceListOptions) ListOpts {
	clopts := ListOpts{Ctx: lopts.Ctx, ExpressionSelector: lopts.ExpressionSelector}
	return clopts
}

// TranslateResourceNamespaceListToWatchOptions translates the resource namespace watch options to Watch Options
func TranslateResourceNamespaceListToWatchOptions(wopts resources.ResourceNamespaceWatchOptions) WatchOpts {
	clopts := WatchOpts{Ctx: wopts.Ctx, ExpressionSelector: wopts.ExpressionSelector}
	return clopts
}
