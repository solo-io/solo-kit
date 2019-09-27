package reporter

import (
	"context"
	"strings"

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
}

type ResourceReports map[resources.InputResource]Report

func (e ResourceReports) Accept(res ...resources.InputResource) ResourceReports {
	for _, r := range res {
		e[r] = Report{}
	}
	return e
}

func (e ResourceReports) Merge(resErrs ResourceReports) {
	for k, v := range resErrs {
		e[k] = v
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

func (e ResourceReports) AddWarning(res resources.InputResource, warning string) {
	if warning == "" {
		return
	}
	rpt := e[res]
	rpt.Warnings = append(rpt.Warnings, warning)
	e[res] = rpt
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
				errs = errors.Errorf("invalid resource %v.%v", res.GetMetadata().Namespace, res.GetMetadata().Name)
			}
			errs = multierror.Append(errs, errors.Errorf("WARN: \n  %v", rpt.Warnings))
		}
	}
	return errs
}

type Reporter interface {
	WriteReports(ctx context.Context, errs ResourceReports, subresourceStatuses map[string]*core.Status) error
}

type reporter struct {
	clients clients.ResourceClients
	ref     string
}

func NewReporter(reporterRef string, resourceClients ...clients.ResourceClient) Reporter {
	clientsByKind := make(clients.ResourceClients)
	for _, client := range resourceClients {
		clientsByKind[client.Kind()] = client
	}
	return &reporter{
		ref:     reporterRef,
		clients: clientsByKind,
	}
}

func (r *reporter) WriteReports(ctx context.Context, resourceErrs ResourceReports, subresourceStatuses map[string]*core.Status) error {
	ctx = contextutils.WithLogger(ctx, "reporter")
	logger := contextutils.LoggerFrom(ctx)

	var merr *multierror.Error

	for resource, report := range resourceErrs {
		kind := resources.Kind(resource)
		client, ok := r.clients[kind]
		if !ok {
			return errors.Errorf("reporter: was passed resource of kind %v but no client to support it", kind)
		}
		status := statusFromReport(r.ref, report, subresourceStatuses)
		resourceToWrite := resources.Clone(resource).(resources.InputResource)
		if status.Equal(resource.GetStatus()) {
			logger.Debugf("skipping report for %v as it has not changed", resourceToWrite.GetMetadata().Ref())
			continue
		}
		resourceToWrite.SetStatus(status)
		res, writeErr := client.Write(resourceToWrite, clients.WriteOpts{
			Ctx:               ctx,
			OverwriteExisting: true,
		})
		if writeErr != nil && errors.IsResourceVersion(writeErr) {
			updatedRes, readErr := client.Read(resourceToWrite.GetMetadata().Namespace, resourceToWrite.GetMetadata().Name, clients.ReadOpts{
				Ctx: ctx,
			})
			if readErr == nil {
				if hashutils.HashAll(updatedRes) == hashutils.HashAll(resourceToWrite) {
					// same hash, something not important was done, try again:
					updatedRes.(resources.InputResource).SetStatus(status)
					res, writeErr = client.Write(updatedRes, clients.WriteOpts{
						Ctx:               ctx,
						OverwriteExisting: true,
					})
				}
			} else {
				logger.Warnw("error reading client to compare conflict when writing status", "error", readErr)
			}
		}
		if writeErr != nil {
			err := errors.Wrapf(writeErr, "failed to write status %v for resource %v", status, resource.GetMetadata().Name)
			logger.Warn(err)
			merr = multierror.Append(merr, err)
			continue
		}
		resources.UpdateMetadata(resource, func(meta *core.Metadata) {
			meta.ResourceVersion = res.GetMetadata().ResourceVersion
		})

		logger.Infof("wrote report %v : %v", resourceToWrite.GetMetadata().Ref(), status)
	}
	return merr.ErrorOrNil()
}

func statusFromReport(ref string, report Report, subresourceStatuses map[string]*core.Status) core.Status {

	var warningReason string
	if len(report.Warnings) > 0 {
		warningReason = "warning: \n  " + strings.Join(report.Warnings, "\n")
	}

	if report.Errors != nil {
		errorReason := report.Errors.Error()
		if warningReason != "" {
			errorReason += "\n" + warningReason
		}
		return core.Status{
			State:               core.Status_Rejected,
			Reason:              errorReason,
			ReportedBy:          ref,
			SubresourceStatuses: subresourceStatuses,
		}
	}

	if warningReason != "" {
		return core.Status{
			State:               core.Status_Warning,
			Reason:              warningReason,
			ReportedBy:          ref,
			SubresourceStatuses: subresourceStatuses,
		}
	}

	return core.Status{
		State:               core.Status_Accepted,
		ReportedBy:          ref,
		SubresourceStatuses: subresourceStatuses,
	}
}
