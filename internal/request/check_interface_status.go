package request

import (
	"context"
	"github.com/inexio/thola/internal/device"
	"github.com/pkg/errors"
)

// CheckInterfaceStatusRequest
//
// CheckInterfaceStatusRequest is the request struct for the check interface status request.
//
// swagger:model
type CheckInterfaceStatusRequest struct {
	OperStatus   bool          `yaml:"oper_status" json:"oper_status" xml:"oper_status"`
	AdminStatus  bool          `yaml:"admin_status" json:"admin_status" xml:"admin_status"`
	DesiredState device.Status `yaml:"desired_state" json:"desired_state" xml:"desired_state"`
	InterfaceOptions
	CheckDeviceRequest
}

func (r *CheckInterfaceStatusRequest) validate(ctx context.Context) error {
	if r.DesiredState != device.StatusUp &&
		r.DesiredState != device.StatusDown &&
		r.DesiredState != device.StatusTesting &&
		r.DesiredState != device.StatusUnknown &&
		r.DesiredState != device.StatusDormant &&
		r.DesiredState != device.StatusNotPresent &&
		r.DesiredState != device.StatusLowerLayerDown {
		return errors.New("invalid desired status")
	}
	if err := r.InterfaceOptions.validate(); err != nil {
		return err
	}
	return r.CheckDeviceRequest.validate(ctx)
}
