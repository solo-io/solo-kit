package reporter

import (
	"context"
	"strings"

	"k8s.io/client-go/util/retry"

	"github.com/hashicorp/go-multierror"
	"github.com/solo-io/go-utils/contextutils"

	"github.com/solo-io/go-utils/hashutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/errors"
)

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
			allMessages := make(map[string]bool)
			for _, message := range firstReport.Messages {
				allMessages[message] = true
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

// ignores warnings
func (e ResourceReports) Validate() error {
	var errs error
	for res, rpt := range e {
		if rpt.Errors != nil {
			if errs == nil {
				errs = errors.Errorf("invalid resource %v.%v", res.GetMetadata().Namespace, res.GetMetadata().Name)
			}
			errs = multierror.Append(errs, rpt.Errors)
		}
	}
	return errs
}

// does not ignore warnings
func (e ResourceReports) ValidateStrict() error {
	errs := e.Validate()
	for res, rpt := range e {
		if len(rpt.Warnings) > 0 {
			if errs == nil {
				errs = errors.Errorf(
					"invalid resource %v.%v",
					res.GetMetadata().GetNamespace(),
					res.GetMetadata().GetName(),
				)
			}
			errs = multierror.Append(errs, errors.Errorf("WARN: \n  %v", rpt.Warnings))
		}
	}
	return errs
}

// Minimal set of client operations required for reporters.
type ReporterResourceClient interface {
	Kind() string
	Read(namespace, name string, opts clients.ReadOpts) (resources.Resource, error)
	Write(resource resources.Resource, opts clients.WriteOpts) (resources.Resource, error)
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
		resourceToWrite := resources.Clone(resource).(resources.InputResource)
		resourceStatus := r.statusClient.GetStatus(resource)

		if status.Equal(resourceStatus) {
			logger.Debugf("skipping report for %v as it has not changed", resourceToWrite.GetMetadata().Ref())
			continue
		}

		r.statusClient.SetStatus(resourceToWrite, status)
		var updatedResource resources.Resource
		writeErr := errors.RetryOnConflict(retry.DefaultBackoff, func() error {
			var writeErr error
			updatedResource, resourceToWrite, writeErr = r.attemptUpdateStatus(ctx, client, resourceToWrite, status)
			return writeErr
		})

		if writeErr != nil {
			err := errors.Wrapf(writeErr, "failed to write status %v for resource %v", status, resource.GetMetadata().Name)
			logger.Warn(err)
			merr = multierror.Append(merr, err)
			continue
		}
		if updatedResource != nil {
			logger.Debugf("wrote report for %v : %v", updatedResource.GetMetadata().Ref(), status)
		} else {
			logger.Debugf("did not write report for %v : %v because resource was not found", resourceToWrite.GetMetadata().Ref(), status)
			delete(resourceErrs, resource)
		}
	}
	return merr.ErrorOrNil()
}

// Ideally, this and its caller, WriteReports, would just take the resource ref and its status and/or messages, rather than the resource itself,
//    to avoid confusion about whether this may update the resource rather than just its fields.
//    However, this change is not worth the effort and risk right now. (Ariana, June 2020)
func (r *reporter) attemptUpdateStatus(ctx context.Context, client ReporterResourceClient, resourceToWrite resources.InputResource, statusToWrite *core.Status) (resources.Resource, resources.InputResource, error) {
	var readErr error
	resourceFromRead, readErr := client.Read(resourceToWrite.GetMetadata().Namespace, resourceToWrite.GetMetadata().Name, clients.ReadOpts{Ctx: ctx})
	if readErr != nil && errors.IsNotExist(readErr) { // resource has been deleted, don't re-create
		return nil, resourceToWrite, nil
	}
	if readErr == nil {
		// set resourceToWrite to the resource we read but with the new status and new messages
		// Note: it's possible that this resourceFromRead is newer than the resourceToWrite and therefore the status will be out of sync.
		//    If so, we will soon recalculate the status. The interim incorrect status is not dangerous since the status is informational only.
		//    Also, the status is accurate for the resource as it's stored in Gloo's memory in the interim.
		//    This is explained further here: https://github.com/solo-io/solo-kit/pull/360#discussion_r433397163
		if inputResourceFromRead, ok := resourceFromRead.(resources.InputResource); ok {
			resourceToWrite = inputResourceFromRead
			r.statusClient.SetStatus(resourceToWrite, statusToWrite)
		}
	}
	updatedResource, writeErr := client.Write(resourceToWrite, clients.WriteOpts{Ctx: ctx, OverwriteExisting: true})
	if writeErr == nil {
		return updatedResource, resourceToWrite, nil
	}
	updatedResource, readErr = client.Read(resourceToWrite.GetMetadata().Namespace, resourceToWrite.GetMetadata().Name, clients.ReadOpts{Ctx: ctx})
	if readErr != nil {
		if errors.IsResourceVersion(writeErr) {
			// we don't want to return the unwrapped resource version writeErr if we also had a read error
			// otherwise we could get into infinite retry loop if reads repeatedly failed (e.g., no read RBAC)
			return nil, resourceToWrite, errors.Wrapf(writeErr, "unable to read updated resource, no reason to retry resource version conflict; readErr %v", readErr)
		}
		return nil, resourceToWrite, writeErr
	}

	// we successfully read an updated version of the resource we are
	// trying to update. let's update resourceToWrite for the next iteration
	equal, _ := hashutils.HashableEqual(updatedResource, resourceToWrite)
	if !equal {
		// different hash, something important was done, do not try again:
		return updatedResource, resourceToWrite, nil
	}
	resourceToWriteUpdated := resources.Clone(updatedResource).(resources.InputResource)
	r.statusClient.SetStatus(resourceToWriteUpdated, r.statusClient.GetStatus(resourceToWrite))

	return updatedResource, resourceToWriteUpdated, writeErr
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
