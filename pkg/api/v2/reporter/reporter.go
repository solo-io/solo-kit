package reporter

import (
	"context"
	"os"
	"slices"
	"sort"
	"strings"

	"k8s.io/client-go/util/retry"

	"github.com/hashicorp/go-multierror"
	"github.com/solo-io/go-utils/contextutils"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/errors"
)

const (
	// 1024 chars = 1kb
	MaxStatusBytes = 1024
	MaxStatusKeys  = 100
)

var (
	// only public for unit tests!
	DisableTruncateStatus = false
)

func init() {
	DisableTruncateStatus = os.Getenv("DISABLE_TRUNCATE_STATUS") == "true"
}

type Report struct {
	Warnings []string
	Errors   error

	// Additional information about the current state of the resource.
	Messages []string
}

type ResourceReports map[resources.InputResource]Report

func (e ResourceReports) Accept(res ...resources.InputResource) ResourceReports {
	for _, r := range res {
		e[r] = Report{}
	}
	return e
}

// Merge merges the given resourceReports into this resourceReports.
// Any resources which appear in both resourceReports will
// have their warnings and errors merged.
// Errors appearing in both reports, as determined by the error strings,
// will not be duplicated in the resulting merged report.
func (e ResourceReports) Merge(resErrs ResourceReports) {
	for k, v := range resErrs {
		if firstReport, exists := e[k]; exists {
			// report already exists for this resource,
			// merge new report into existing report:
			secondReport := v

			// Merge warnings lists
			allWarnings := make(map[string]bool)
			for _, warning := range firstReport.Warnings {
				allWarnings[warning] = true
			}
			for _, warning := range secondReport.Warnings {
				if _, found := allWarnings[warning]; !found {
					firstReport.Warnings = append(firstReport.Warnings, warning)
				}
			}

			// Merge messages lists
			allMessages := make(map[string]struct{})
			for _, message := range firstReport.Messages {
				allMessages[message] = struct{}{}
			}
			for _, message := range secondReport.Messages {
				if _, found := allMessages[message]; !found {
					firstReport.Messages = append(firstReport.Messages, message)
				}
			}

			if firstReport.Errors == nil {
				// Only 2nd has errs
				firstReport.Errors = secondReport.Errors
				e[k] = firstReport
				continue
			} else if secondReport.Errors == nil {
				// Only 1st has errs
				e[k] = firstReport
				continue
			}

			// Both first and second have errors for the same resource:
			// Any errors which are identical won't be duplicated,
			// Any errors which are unique will be added to the final list
			errs1, isFirstMulti := firstReport.Errors.(*multierror.Error)
			errs2, isSecondMulti := secondReport.Errors.(*multierror.Error)

			// If the errors are not mutliErrs, wrap them in multiErrs:
			if !isFirstMulti {
				errs1 = &multierror.Error{Errors: []error{firstReport.Errors}}
			}
			if !isSecondMulti {
				errs2 = &multierror.Error{Errors: []error{secondReport.Errors}}
			}

			allErrsMap := make(map[string]error)
			for _, err := range errs1.Errors {
				allErrsMap[err.Error()] = err
			}
			for _, err := range errs2.Errors {
				if _, found := allErrsMap[err.Error()]; !found {
					allErrsMap[err.Error()] = err
					errs1.Errors = append(errs1.Errors, err)
				}
			}
			firstReport.Errors = errs1

			e[k] = firstReport
		} else {
			// Resource in 2nd report is not yet in 1st report
			e[k] = v
		}
	}
}

func (e ResourceReports) AddErrors(res resources.InputResource, errs ...error) {
	for _, err := range errs {
		e.AddError(res, err)
	}
}

func (e ResourceReports) AddError(res resources.InputResource, err error) {
	if err == nil {
		return
	}
	rpt := e[res]
	rpt.Errors = multierror.Append(rpt.Errors, err)
	e[res] = rpt
}

func (e ResourceReports) AddWarnings(res resources.InputResource, warning ...string) {
	for _, warn := range warning {
		e.AddWarning(res, warn)
	}
}

func (e ResourceReports) AddWarning(res resources.InputResource, warning string) {
	if warning == "" {
		return
	}
	rpt := e[res]
	rpt.Warnings = append(rpt.Warnings, warning)
	e[res] = rpt
}

func (e ResourceReports) AddMessages(res resources.InputResource, messages ...string) {
	for _, message := range messages {
		e.AddMessage(res, message)
	}
}

func (e ResourceReports) AddMessage(res resources.InputResource, message string) {
	if message == "" {
		return
	}
	rpt := e[res]
	rpt.Messages = append(rpt.Messages, message)
	e[res] = rpt
}

func (e ResourceReports) Find(kind string, ref *core.ResourceRef) (resources.InputResource, Report) {
	for res, rpt := range e {
		if resources.Kind(res) == kind && res.GetMetadata().Ref().Equal(ref) {
			return res, rpt
		}
	}
	return nil, Report{}
}

func (e ResourceReports) FilterByKind(kind string) ResourceReports {
	reports := make(map[resources.InputResource]Report)
	for res, rpt := range e {
		if resources.Kind(res) == kind {
			reports[res] = rpt
		}
	}
	return reports
}

// refMapAndSortedKeys returns a map of resource refs to resources and a sorted list of resource refs
// This is used to iterate over the resources in a consistent order.
// The reports are keyed by references to the resources, so are not sortable.
// There is no unique key for a resource, so we use the string representation of the name/namespace as the key,
// and collect all the resources with the same key together.
// The previous implementation did not guarantee a consistent order when iterating over the resources,
// so any order used here will be acceptable for backwards compatibility.
func (e ResourceReports) refMapAndSortedKeys() (map[string][]resources.InputResource, []string) {
	// refKeys contains all the unique keys for the resources
	var refKeys []string
	// refMaps is a map of resource keys to a slice of resources with that key
	var refMap = make(map[string][]resources.InputResource)

	// Loop over the resources
	for res := range e {

		// Get a string representation of the resource ref. This is not guaranteed to be unique.
		resKey := res.GetMetadata().Ref().String()

		// Add the key to the list of keys if it is not already there
		if !slices.Contains(refKeys, resKey) {
			refKeys = append(refKeys, resKey)
		}
		// Add the resource to the map of resources with the same key
		refMap[resKey] = append(refMap[resKey], res)
	}

	// Sort the name-namespace keys. This will allow the reports to be accssed in a consistent order,
	// except in cases where the name/namespace is not unique. In those cases, the individual validaiton
	// functions will have to handle consistent ordering.
	slices.Sort(refKeys)
	return refMap, refKeys
}

// sortErrors sorts errors based on string representation
// Note: because we are using multierror the string representation starts with "x errors occurred".
// This will be consistent, but possibly counter-intuitive for tests.
func sortErrors(errs []error) {
	sort.Slice(errs, func(i, j int) bool {
		return errs[i].Error() < errs[j].Error()
	})
}

// Validate ignores warnings
func (e ResourceReports) Validate() error {
	var errs error
	refMap, refKeys := e.refMapAndSortedKeys()

	// the refKeys are sorted/sortable and point to the unsortable resources refs that are the keys to the report map
	for _, refKey := range refKeys {
		// name/namespace is not unique, so we collect those references together
		reses := refMap[refKey]

		var errsForKey []error
		for _, res := range reses {
			rpt := e[res]

			if rpt.Errors != nil {
				errsForKey = append(errsForKey, rpt.Errors)
			}
		}

		if len(errsForKey) > 0 {
			if errs == nil {
				// All of the resources in the group have the same metadata, so use the first one
				errs = errors.Errorf("invalid resource %v.%v", reses[0].GetMetadata().Namespace, reses[0].GetMetadata().Name)
			}
			sortErrors(errsForKey)

			for _, err := range errsForKey {
				errs = multierror.Append(errs, err)
			}
		}
	}
	return errs
}

// ValidateStrict does not ignore warnings. If warnings are present, they will be included in the error.
// If an error is not present but warnings are, an "invalid resource" error will be returned along with each warning.
func (e ResourceReports) ValidateStrict() error {
	errs := e.Validate()
	refMap, refKeys := e.refMapAndSortedKeys()

	for _, refKey := range refKeys {
		var errsForKey []error
		reses := refMap[refKey]

		// name/namespace is not unique, so we collect those references together
		for _, res := range reses {
			rpt := e[res]
			if len(rpt.Warnings) > 0 {
				errsForKey = append(errsForKey, errors.Errorf("WARN: \n  %v", rpt.Warnings))

			}
		}

		if len(errsForKey) > 0 {
			if errs == nil {
				// All of the resources in the group have the same metadata, so use the first one
				errs = errors.Errorf(
					"invalid resource %v.%v",
					reses[0].GetMetadata().GetNamespace(),
					reses[0].GetMetadata().GetName(),
				)
			}
			sortErrors(errsForKey)

			for _, err := range errsForKey {
				errs = multierror.Append(errs, err)
			}
		}

	}
	return errs
}

func (e ResourceReports) ValidateSeparateWarnings() (error, error) {
	var warnings error

	errs := e.Validate()
	refMap, refKeys := e.refMapAndSortedKeys()

	for _, refKey := range refKeys {
		// name/namespace is not unique, so we collect those references together
		var warnForKey []error
		reses := refMap[refKey]

		for _, res := range reses {
			rpt := e[res]
			if len(rpt.Warnings) > 0 {
				warnForKey = append(warnForKey, errors.Errorf("WARN: \n  %v", rpt.Warnings))

			}
		}

		if len(warnForKey) > 0 {
			sortErrors(warnForKey)

			for _, err := range warnForKey {
				warnings = multierror.Append(warnings, err)
			}
		}

	}

	return errs, warnings
}

// WarningHandling is an enum for how to handle warnings when validating reports with `ValidateWithWarnings`
type WarningHandling int

const (
	// With Strict WarningHandling, warnings are treated as errors
	Strict WarningHandling = iota
	// With IgnoreWarnings WarningHandling, warnings are ignored
	IgnoreWarnings
	// With SeparateWarnings WarningHandling, warnings are returned separately from errors
	SeparateWarnings
)

// ValidateReport validates the reports according to the given validation type.
func (e ResourceReports) ValidateWithWarnings(t WarningHandling) (error, error) {
	switch t {
	case Strict:
		return e.ValidateStrict(), nil
	case IgnoreWarnings:
		return e.Validate(), nil
	case SeparateWarnings:
		return e.ValidateSeparateWarnings()
	default:
		return errors.Errorf("unknown validation type %v", t), nil
	}
}

// Minimal set of client operations required for reporters.
type ReporterResourceClient interface {
	Kind() string
	ApplyStatus(statusClient resources.StatusClient, inputResource resources.InputResource, opts clients.ApplyStatusOpts) (resources.Resource, error)
}

type Reporter interface {
	WriteReports(ctx context.Context, errs ResourceReports, subresourceStatuses map[string]*core.Status) error
}
type StatusReporter interface {
	Reporter
	StatusFromReport(report Report, subresourceStatuses map[string]*core.Status) *core.Status
}

type reporter struct {
	reporterRef  string
	statusClient resources.StatusClient
	clients      map[string]ReporterResourceClient
}

func NewReporter(reporterRef string, statusClient resources.StatusClient, reporterClients ...ReporterResourceClient) StatusReporter {
	clientsByKind := make(map[string]ReporterResourceClient)
	for _, client := range reporterClients {
		clientsByKind[client.Kind()] = client
	}
	return &reporter{
		reporterRef:  reporterRef,
		statusClient: statusClient,
		clients:      clientsByKind,
	}
}

// ResourceReports may be modified, and end up with fewer resources than originally requested.
// If resources referenced in the resourceErrs don't exist, they will be removed.
func (r *reporter) WriteReports(ctx context.Context, resourceErrs ResourceReports, subresourceStatuses map[string]*core.Status) error {
	ctx = contextutils.WithLogger(ctx, "reporter")
	logger := contextutils.LoggerFrom(ctx)

	var merr *multierror.Error

	// copy the map so we can iterate over the copy, deleting resources from
	// the original map if they are not found/no longer exist.
	resourceErrsCopy := make(ResourceReports, len(resourceErrs))
	for resource, report := range resourceErrs {
		resourceErrsCopy[resource] = report
	}

	for resource, report := range resourceErrsCopy {
		kind := resources.Kind(resource)
		client, ok := r.clients[kind]
		if !ok {
			return errors.Errorf("reporter: was passed resource of kind %v but no client to support it", kind)
		}
		status := r.StatusFromReport(report, subresourceStatuses)
		if !DisableTruncateStatus {
			status = trimStatus(status)
		}
		resourceStatus := r.statusClient.GetStatus(resource)

		if status.Equal(resourceStatus) {
			logger.Debugf("skipping report for %v as it has not changed", resource.GetMetadata().Ref())
			continue
		}

		resourceToWrite := resources.Clone(resource).(resources.InputResource)
		r.statusClient.SetStatus(resourceToWrite, status)
		writeErr := errors.RetryOnConflict(retry.DefaultBackoff, func() error {
			return r.attemptUpdateStatus(ctx, client, resourceToWrite, status)
		})

		if errors.IsNotExist(writeErr) {
			logger.Debugf("did not write report for %v : %v because resource was not found", resourceToWrite.GetMetadata().Ref(), status)
			delete(resourceErrs, resource)
			continue
		}

		if writeErr != nil {
			err := errors.Wrapf(writeErr, "failed to write status %v for resource %v", status, resource.GetMetadata().GetName())
			logger.Warn(err)
			merr = multierror.Append(merr, err)
			continue
		}
		logger.Debugf("wrote report for %v : %v", resource.GetMetadata().Ref(), status)

	}
	return merr.ErrorOrNil()
}

func (r *reporter) attemptUpdateStatus(ctx context.Context, client ReporterResourceClient, resourceToWrite resources.InputResource, statusToWrite *core.Status) error {
	_, patchErr := client.ApplyStatus(r.statusClient, resourceToWrite, clients.ApplyStatusOpts{Ctx: ctx})
	return patchErr
}

func (r *reporter) StatusFromReport(report Report, subresourceStatuses map[string]*core.Status) *core.Status {
	var messages []string
	if len(report.Messages) != 0 {
		messages = report.Messages
	}

	var warningReason string
	if len(report.Warnings) > 0 {
		warningReason = "warning: \n  " + strings.Join(report.Warnings, "\n")
	}

	if report.Errors != nil {
		errorReason := report.Errors.Error()
		if warningReason != "" {
			errorReason += "\n" + warningReason
		}
		return &core.Status{
			State:               core.Status_Rejected,
			Reason:              errorReason,
			ReportedBy:          r.reporterRef,
			SubresourceStatuses: subresourceStatuses,
			Messages:            messages,
		}
	}

	if warningReason != "" {
		return &core.Status{
			State:               core.Status_Warning,
			Reason:              warningReason,
			ReportedBy:          r.reporterRef,
			SubresourceStatuses: subresourceStatuses,
			Messages:            messages,
		}
	}

	return &core.Status{
		State:               core.Status_Accepted,
		ReportedBy:          r.reporterRef,
		SubresourceStatuses: subresourceStatuses,
		Messages:            messages,
	}
}

func trimStatus(status *core.Status) *core.Status {
	// truncate status reason to a kilobyte, with max 100 keys in subresource statuses
	return trimStatusForMaxSize(status, MaxStatusBytes, MaxStatusKeys)
}

func trimStatusForMaxSize(status *core.Status, bytesPerKey, maxKeys int) *core.Status {
	if status == nil {
		return nil
	}
	if len(status.Reason) > bytesPerKey {
		status.Reason = status.Reason[:bytesPerKey]
	}

	if len(status.SubresourceStatuses) > maxKeys {
		// sort for idempotency
		keys := make([]string, 0, len(status.SubresourceStatuses))
		for key := range status.SubresourceStatuses {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		trimmedSubresourceStatuses := make(map[string]*core.Status, maxKeys)
		for _, key := range keys[:maxKeys] {
			trimmedSubresourceStatuses[key] = status.SubresourceStatuses[key]
		}
		status.SubresourceStatuses = trimmedSubresourceStatuses
	}

	for key, childStatus := range status.SubresourceStatuses {
		// divide by two so total memory usage is bounded at: (num_keys * bytes_per_key) + (num_keys / 2 * bytes_per_key / 2) + ...
		// 100 * 1024b + 50 * 512b + 25 * 256b + 12 * 128b + 6 * 64b + 3 * 32b + 1 * 16b ~= 136 kilobytes
		//
		// 2147483647 bytes is k8s -> etcd limit in grpc connection. 2147483647 / 136 ~= 15788 resources at limit before we see an issue
		// https://github.com/solo-io/solo-projects/issues/4120
		status.SubresourceStatuses[key] = trimStatusForMaxSize(childStatus, bytesPerKey/2, maxKeys/2)
	}
	return status
}
