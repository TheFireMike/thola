package codecommunicator

import (
	"context"
	"github.com/inexio/thola/internal/device"
	"github.com/inexio/thola/internal/network"
	"github.com/pkg/errors"
)

type sophosXGCommunicator struct {
	codeCommunicator
}

func (c *sophosXGCommunicator) GetHighAvailabilityComponentState(ctx context.Context) (device.HighAvailabilityComponentState, error) {
	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.SNMP == nil {
		return "", errors.New("no device connection available")
	}

	statusRes, err := con.SNMP.SnmpClient.SNMPWalk(ctx, "1.3.6.1.4.1.2604.5.1.4.1") // sfosHAStatus
	if err != nil {
		return "", errors.Wrap(err, "failed to read out high-availability status")
	}

	if len(statusRes) < 1 {
		return "", errors.New("failed to read out high-availability status")
	}

	status, err := statusRes[0].GetValue()
	if err != nil {
		return "", errors.Wrap(err, "failed to get high-availability status")
	}

	if status.String() == "0" {
		return device.HighAvailabilityComponentStateStandalone, nil
	}

	modeRes, err := con.SNMP.SnmpClient.SNMPWalk(ctx, "1.3.6.1.4.1.2604.5.1.4.4") // sfosDeviceCurrentHAState
	if err != nil {
		return "", errors.Wrap(err, "failed to read out high-availability mode")
	}

	if len(modeRes) < 1 {
		return "", errors.New("failed to read out high-availability mode")
	}

	mode, err := modeRes[0].GetValue()
	if err != nil {
		return "", errors.Wrap(err, "failed to get high-availability mode")
	}

	if mode.String() == "3" || mode.String() == "1" {
		return device.HighAvailabilityComponentStateSynchronized, nil
	} else {
		return device.HighAvailabilityComponentStateUnsynchronized, nil
	}
}

func (c *sophosXGCommunicator) GetHighAvailabilityComponentRole(ctx context.Context) (string, error) {
	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.SNMP == nil {
		return "", errors.New("no device connection available")
	}

	state, err := c.GetHighAvailabilityComponentState(ctx)
	if err != nil {
		return "", errors.Wrap(err, "failed to read out high-availability state")
	}

	if state == device.HighAvailabilityComponentStateStandalone {
		return "", errors.New("device is not in high-availability mode (state = standalone)")
	}

	modeRes, err := con.SNMP.SnmpClient.SNMPWalk(ctx, "1.3.6.1.4.1.2604.5.1.4.4") // sfosDeviceCurrentHAState
	if err != nil {
		return "", errors.Wrap(err, "failed to read out high-availability mode")
	}

	if len(modeRes) < 1 {
		return "", errors.New("failed to read out high-availability mode")
	}

	mode, err := modeRes[0].GetValue()
	if err != nil {
		return "", errors.Wrap(err, "failed to get high-availability mode")
	}

	if mode.String() == "3" || mode.String() == "2" {
		return "primary", nil
	} else if mode.String() == "1" {
		return "auxiliary", nil
	} else {
		return "", errors.New("currently neither primary nor auxiliary")
	}
}

func (c *sophosXGCommunicator) GetHighAvailabilityComponentNodes(ctx context.Context) (int, error) {
	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.SNMP == nil {
		return 0, errors.New("no device connection available")
	}

	state, err := c.GetHighAvailabilityComponentState(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "failed to read out high-availability state")
	}

	if state == device.HighAvailabilityComponentStateStandalone {
		return 0, errors.New("device is not in high-availability mode (state = standalone)")
	}

	return 2, nil
}
