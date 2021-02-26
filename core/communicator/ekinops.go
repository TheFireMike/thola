package communicator

import (
	"context"
	"fmt"
	"github.com/inexio/thola/core/device"
	"github.com/inexio/thola/core/network"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"strings"
)

type ekinopsCommunicator struct {
	baseCommunicator
}

func (c *ekinopsCommunicator) GetInterfaces(ctx context.Context) ([]device.Interface, error) {
	interfaces, err := c.sub.GetInterfaces(ctx)
	if err != nil {
		return nil, err
	}

	con, ok := network.DeviceConnectionFromContext(ctx)
	if !ok || con.SNMP == nil {
		return nil, errors.New("no device connection available")
	}

	//get used slots
	slotResults, err := con.SNMP.SnmpClient.SNMPWalk(ctx, ".1.3.6.1.4.1.20044.7.8.1.1.2")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to snmpwalk")
	}

	//get used modules
	moduleResults, err := con.SNMP.SnmpClient.SNMPWalk(ctx, ".1.3.6.1.4.1.20044.7.8.1.1.3")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to snmpwalk")
	}

	var moduleReaders []ekinopsModuleReader

	for k, slotResult := range slotResults {
		slotIdentifier, err := slotResult.GetValueString()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get snmp result as string")
		}

		module, err := moduleResults[k].GetValueString()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get module result as string")
		}

		moduleReader, err := ekinopsGetModuleReader(slotIdentifier, module)
		if err != nil {
			log.Ctx(ctx).Trace().Err(err).Msgf("no information for reading out ekinops module '%s' available", module)
			continue
		}
		moduleReaders = append(moduleReaders, moduleReader)
	}

	for _, moduleReader := range moduleReaders {
		interfaces, err = moduleReader.readModuleMetrics(ctx, interfaces)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to read out module specific metrics for module '%s' (slot: %s)", moduleReader.getModuleName(), moduleReader.getSlotIdentifier())
		}
	}

	return normalizeEkinopsInterfaces(interfaces)
}

func ekinopsInterfacesIfIdentifierToSliceIndex(interfaces []device.Interface) (map[string]int, error) {
	m := make(map[string]int)
	for k, interf := range interfaces {
		if interf.IfName == nil {
			return nil, fmt.Errorf("no ifName set for interface ifIndex: `%d`", *interf.IfIndex)
		}
		identifier := strings.Join(strings.Split(*interf.IfName, "/")[2:], "/")

		if _, ok := m[identifier]; ok {
			return nil, fmt.Errorf("interface identifier `%s` exists multiple times", *interf.IfName)
		}

		m[identifier] = k
	}
	return m, nil
}

func normalizeEkinopsInterfaces(interfaces []device.Interface) ([]device.Interface, error) {
	var res []device.Interface

	for _, interf := range interfaces {
		if interf.IfDescr == nil {
			return nil, fmt.Errorf("no IfDescr set for interface ifIndex: `%d`", *interf.IfIndex)
		}

		// change ifType of ports of slots > 1 to "fibreChannel" if ifType equals "other"
		slotNumber := strings.Split(*interf.IfName, "/")[2]
		if !(slotNumber == "0" || slotNumber == "1") {
			if interf.IfType == nil || *interf.IfType == "other" {
				fibreChannel := "fibreChannel"
				interf.IfType = &fibreChannel
			}
		}

		// change ifType of OPM8 ports
		moduleName := strings.Split(*interf.IfDescr, "/")[3]
		if moduleName == "PM_OPM8" {
			ifType := "ChannelMonitoring" //TODO switch ifType name
			interf.IfType = &ifType
		}

		// change ifDescr and ifName of every interface
		// they no longer contain Ekinops/C...
		// and are now in the form <slot>-[Description]
		*interf.IfDescr = slotNumber + "-" + strings.Split(strings.Split(*interf.IfDescr, "/")[4], "(")[0]
		interf.IfName = interf.IfDescr

		// remove every port on slot 0 starting with "0-FE_"
		if slotNumber != "0" || !strings.HasPrefix(*interf.IfName, "0-FE_") {
			res = append(res, interf)
		}
	}

	return res, nil
}