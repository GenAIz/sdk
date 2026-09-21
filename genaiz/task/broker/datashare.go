package broker

import (
	"strings"

	"genaiz.com/genaiz-lib/lang/slicez"
	"genaiz.com/genaiz/task"
)

var (
	errorDataShareLinkInvalid    = task.NewError("can not query for inactive links")
	errorDataShareHandleInvalid  = task.NewError("handle is invalid for version")
	errorDataShareOemRequired    = task.NewError("oem is required for filtering")
	errorDataShareVersionInvalid = task.NewError("version is invalid for sequence")
)

type DataInstanceListParams struct {
	Broker
	*DataLink
}

func (dsl DataInstanceListParams) filter(instances []DataLinkInstance) []DataLinkInstance {
	return slicez.Filter(instances, func(instance DataLinkInstance) bool {
		if instance.HasLink(dsl.DataLink) {
			return dsl.Seq == nil || (*dsl.Seq == *instance.DataLink.Seq)
		}

		// Case where we only filter by oem
		if strings.EqualFold(dsl.Oem, instance.DataLink.Oem) {
			if dsl.Handle != "" {
				// Case where we filter by oem and handle
				return strings.EqualFold(dsl.Handle, instance.DataLink.Handle)
			}

			// If params had a version, HasLink would return true
			return true
		}

		return false
	})
}

func (dsl DataInstanceListParams) getLinkId() *int64 {
	if dsl.DataLink != nil {
		return dsl.DataLink.Id
	}

	return nil
}

func (dsl DataInstanceListParams) hasDataLinkFilter() bool {
	return dsl.DataLink != nil && dsl.Oem != ""
}

func (dsl DataInstanceListParams) isVersionEqual(instance *DataLinkInstance) bool {
	if dsl.DataLink != nil && instance != nil {
		return dsl.Version == instance.DataLink.Version &&
			((dsl.Seq == nil && instance.DataLink.Seq == nil) ||
				(dsl.Seq != nil && instance.DataLink.Seq != nil && *dsl.Seq == *instance.DataLink.Seq))
	}

	return false
}

func (dsl DataInstanceListParams) validateDataLinkFilter() error {
	if dsl.DataLink != nil {
		if dsl.Oem == "" && (dsl.Handle != "" || dsl.Version != "" || dsl.Seq != nil) {
			return errorDataShareOemRequired
		}

		if dsl.Handle == "" && (dsl.Version != "" || dsl.Seq != nil) {
			return errorDataShareHandleInvalid
		}

		if dsl.Version == "" && dsl.Seq != nil {
			return errorDataShareVersionInvalid
		}

		if dsl.Id != nil && !dsl.IsActive() {
			return errorDataShareLinkInvalid
		}
	}

	return nil
}
