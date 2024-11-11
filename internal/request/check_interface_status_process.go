//go:build !client
// +build !client

package request

import (
	"context"
	"fmt"
	"github.com/inexio/go-monitoringplugin"
	"github.com/inexio/thola/internal/deviceclass/groupproperty"
	"github.com/inexio/thola/internal/network"
	"github.com/pkg/errors"
)

func (r *CheckInterfaceStatusRequest) process(ctx context.Context) (Response, error) {
	r.init()

	ctx = network.NewContextWithSNMPGetsInsteadOfWalk(ctx, r.SNMPGetsInsteadOfWalk)

	com, err := GetCommunicator(ctx, r.BaseRequest)
	if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "failed to get communicator", true) {
		r.mon.PrintPerformanceData(false)
		return &CheckResponse{r.mon.GetInfo()}, nil
	}

	interfaces, err := com.GetInterfaces(ctx, r.getFilter()...)
	if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "failed to read out interfaces", true) {
		r.mon.PrintPerformanceData(false)
		return &CheckResponse{r.mon.GetInfo()}, nil
	}

	err = normalizeInterfaces(interfaces, r.ifDescrRegex, r.IfDescrRegexReplace)
	if r.mon.UpdateStatusOnError(err, monitoringplugin.UNKNOWN, "error while normalizing interfaces", true) {
		r.mon.PrintPerformanceData(false)
		return &CheckResponse{r.mon.GetInfo()}, nil
	}

	for _, i := range interfaces {
		if r.OperStatus {
			if i.IfOperStatus == nil {
				r.mon.UpdateStatus(monitoringplugin.UNKNOWN, fmt.Sprintf("could not determine oper status of interface %s", *i.IfDescr))
			} else {
				value, err := i.IfOperStatus.ToStatusCode()
				if err != nil {
					r.mon.PrintPerformanceData(false)
					return &CheckResponse{r.mon.GetInfo()}, errors.Wrap(err, "failed to convert oper status")
				}
				err = r.mon.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("interface_oper_status", value).SetLabel(*i.IfDescr))
				if err != nil {
					r.mon.PrintPerformanceData(false)
					return &CheckResponse{r.mon.GetInfo()}, errors.Wrap(err, "failed to add oper status")
				}

				if *i.IfOperStatus != r.DesiredState {
					r.mon.UpdateStatus(monitoringplugin.CRITICAL, fmt.Sprintf("oper status of interface %s is %s (wanted: %s)", *i.IfDescr, string(*i.IfOperStatus), string(r.DesiredState)))
				}
			}
		}

		if r.AdminStatus {
			if i.IfAdminStatus == nil {
				r.mon.UpdateStatus(monitoringplugin.UNKNOWN, fmt.Sprintf("could not determine admin status of interface %s", *i.IfDescr))
			} else {
				value, err := i.IfAdminStatus.ToStatusCode()
				if err != nil {
					r.mon.PrintPerformanceData(false)
					return &CheckResponse{r.mon.GetInfo()}, errors.Wrap(err, "failed to convert admin status")
				}
				err = r.mon.AddPerformanceDataPoint(monitoringplugin.NewPerformanceDataPoint("interface_admin_status", value).SetLabel(*i.IfDescr))
				if err != nil {
					r.mon.PrintPerformanceData(false)
					return &CheckResponse{r.mon.GetInfo()}, errors.Wrap(err, "failed to add admin status")
				}

				if *i.IfAdminStatus != r.DesiredState {
					r.mon.UpdateStatus(monitoringplugin.CRITICAL, fmt.Sprintf("admin status of interface %s is %s (wanted: %s)", *i.IfDescr, string(*i.IfAdminStatus), string(r.DesiredState)))
				}
			}
		}
	}

	return &CheckResponse{r.mon.GetInfo()}, nil
}

func (r *CheckInterfaceStatusRequest) getFilter() []groupproperty.Filter {
	requiredValues := []string{"ifDescr", "ifIndex"}
	if r.AdminStatus {
		requiredValues = append(requiredValues, "ifAdminStatus")
	}
	if r.OperStatus {
		requiredValues = append(requiredValues, "ifOperStatus")
	}
	r.Values = requiredValues

	return r.InterfaceOptions.getFilter()
}
