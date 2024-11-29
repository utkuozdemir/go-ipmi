package ipmi

import (
	"fmt"
)

type DCMICapParameter interface {
	_isDCMICapParamter()
	Parameter
}

type DCMIConfigParameter interface {
	_isDCMIConfigParameter()
	Parameter
}

var (
	_ DCMICapParameter = (*DCMICapParam_SupportedDCMICapabilities)(nil)
	_ DCMICapParameter = (*DCMICapParam_MandatoryPlatformAttributes)(nil)
	_ DCMICapParameter = (*DCMICapParam_OptionalPlatformAttributes)(nil)
	_ DCMICapParameter = (*DCMICapParam_ManageabilityAccessAttributes)(nil)
	_ DCMICapParameter = (*DCMICapParam_EnhancedSystemPowerStatisticsAttributes)(nil)

	_ DCMIConfigParameter = (*DCMIConfigParam_ActivateDHCP)(nil)
	_ DCMIConfigParameter = (*DCMIConfigParam_DiscoveryConfiguration)(nil)
	_ DCMIConfigParameter = (*DCMIConfigParam_DHCPTiming1)(nil)
	_ DCMIConfigParameter = (*DCMIConfigParam_DHCPTiming2)(nil)
	_ DCMIConfigParameter = (*DCMIConfigParam_DHCPTiming3)(nil)
)

type DCMICapParamSelector uint8

const (
	DCMICapParamSelector_SupportedDCMICapabilities               = DCMICapParamSelector(0x01)
	DCMICapParamSelector_MandatoryPlatformAttributes             = DCMICapParamSelector(0x02)
	DCMICapParamSelector_OptionalPlatformAttributes              = DCMICapParamSelector(0x03)
	DCMICapParamSelector_ManageabilityAccessAttributes           = DCMICapParamSelector(0x04)
	DCMICapParamSelector_EnhancedSystemPowerStatisticsAttributes = DCMICapParamSelector(0x05)
)

type DCMICapabilities struct {
	SupportedDCMICapabilities               DCMICapParam_SupportedDCMICapabilities
	MandatoryPlatformAttributes             DCMICapParam_MandatoryPlatformAttributes
	OptionalPlatformAttributes              DCMICapParam_OptionalPlatformAttributes
	ManageabilityAccessAttributes           DCMICapParam_ManageabilityAccessAttributes
	EnhancedSystemPowerStatisticsAttributes DCMICapParam_EnhancedSystemPowerStatisticsAttributes
}

func (dcmiCap *DCMICapabilities) Format() string {
	return fmt.Sprintf(`
%s
%s
%s
%s
%s`,
		dcmiCap.SupportedDCMICapabilities.Format(),
		dcmiCap.MandatoryPlatformAttributes.Format(),
		dcmiCap.OptionalPlatformAttributes.Format(),
		dcmiCap.ManageabilityAccessAttributes.Format(),
		dcmiCap.EnhancedSystemPowerStatisticsAttributes.Format(),
	)
}

type DCMICapParam_SupportedDCMICapabilities struct {
	SupportPowerManagement bool
	SupportInBandKCS       bool
	SupportOutOfBandSerial bool
	SupportOutOfBandLAN    bool
}

func (dcmiCap *DCMICapParam_SupportedDCMICapabilities) _isDCMICapParamter() {}

func (dcmiCap *DCMICapParam_SupportedDCMICapabilities) Pack() []byte {
	return []byte{}
}

func (dcmiCap *DCMICapParam_SupportedDCMICapabilities) Unpack(paramData []byte) error {
	if len(paramData) < 3 {
		return ErrUnpackedDataTooShortWith(len(paramData), 3)
	}

	dcmiCap.SupportPowerManagement = isBit0Set(paramData[1])
	dcmiCap.SupportInBandKCS = isBit0Set(paramData[2])
	dcmiCap.SupportOutOfBandSerial = isBit1Set(paramData[2])
	dcmiCap.SupportOutOfBandLAN = isBit2Set(paramData[2])

	return nil
}

func (dcmiCap *DCMICapParam_SupportedDCMICapabilities) Format() string {
	return fmt.Sprintf(`
    Supported DCMI capabilities:

        Optional platform capabilities
            Power management                  (%s)

        Manageability access capabilities
            In-band KCS channel               (%s)
            Out-of-band serial TMODE          (%s)
            Out-of-band secondary LAN channel (%s)
`,
		formatBool(dcmiCap.SupportPowerManagement, "available", "unavailable"),
		formatBool(dcmiCap.SupportInBandKCS, "available", "unavailable"),
		formatBool(dcmiCap.SupportOutOfBandSerial, "available", "unavailable"),
		formatBool(dcmiCap.SupportOutOfBandLAN, "available", "unavailable"),
	)
}

type DCMICapParam_MandatoryPlatformAttributes struct {
	SELAutoRolloverEnabled           bool
	EntireSELFlushUponRollOver       bool
	RecordLevelSELFlushUponRollOver  bool
	SELEntriesCount                  uint16 //only 12 bits, [11-0] Number of SEL entries (Maximum 4096)
	TemperatrureSamplingFrequencySec uint8
}

func (dcmiCap *DCMICapParam_MandatoryPlatformAttributes) _isDCMICapParamter() {}

func (dcmiCap *DCMICapParam_MandatoryPlatformAttributes) Pack() []byte {
	return []byte{}
}

func (dcmiCap *DCMICapParam_MandatoryPlatformAttributes) Unpack(paramData []byte) error {
	if len(paramData) < 5 {
		return ErrUnpackedDataTooShortWith(len(paramData), 5)
	}

	b1 := paramData[1]
	dcmiCap.SELAutoRolloverEnabled = isBit7Set(b1)
	dcmiCap.EntireSELFlushUponRollOver = isBit6Set(b1)
	dcmiCap.RecordLevelSELFlushUponRollOver = isBit5Set(b1)

	b_0_1, _, _ := unpackUint16L(paramData, 0)
	dcmiCap.SELEntriesCount = b_0_1 & 0x0FFF

	dcmiCap.TemperatrureSamplingFrequencySec = paramData[4]

	return nil
}

func (dcmiCap *DCMICapParam_MandatoryPlatformAttributes) Format() string {
	return fmt.Sprintf(`
    Mandatory platform attributes:

        SEL Attributes:
            SEL automatic rollover is  (%s)
            %d SEL entries

        Identification Attributes:

        Temperature Monitoring Attributes:
            Temperature sampling frequency is %d seconds
`,
		formatBool(dcmiCap.SELAutoRolloverEnabled, "enabled", "disabled"),
		dcmiCap.SELEntriesCount,
		dcmiCap.TemperatrureSamplingFrequencySec,
	)
}

type DCMICapParam_OptionalPlatformAttributes struct {
	PowerMgmtDeviceSlaveAddr         uint8
	PewerMgmtControllerChannelNumber uint8
	DeviceRevision                   uint8
}

func (dcmiCap *DCMICapParam_OptionalPlatformAttributes) _isDCMICapParamter() {}

func (param *DCMICapParam_OptionalPlatformAttributes) Pack() []byte {
	return []byte{}
}

func (param *DCMICapParam_OptionalPlatformAttributes) Unpack(paramData []byte) error {
	if len(paramData) < 2 {
		return ErrUnpackedDataTooShortWith(len(paramData), 3)
	}

	param.PowerMgmtDeviceSlaveAddr = paramData[0]
	param.PewerMgmtControllerChannelNumber = paramData[1] & 0xF0
	param.DeviceRevision = paramData[1] & 0x0F

	return nil
}

func (param *DCMICapParam_OptionalPlatformAttributes) Format() string {
	return fmt.Sprintf(`
    Optional Platform Attributes:

        Power Management:
            Slave address of device: %#02x
            Channel number is %#02x %s
            Device revision is %d
`,
		param.PowerMgmtDeviceSlaveAddr,
		param.PewerMgmtControllerChannelNumber,
		formatBool(param.PewerMgmtControllerChannelNumber == 0, "(Primary BMC)", ""),
		param.DeviceRevision,
	)
}

type DCMICapParam_ManageabilityAccessAttributes struct {
	PrimaryLANChannelNumber   uint8
	SecondaryLANChannelNumber uint8
	SerialChannelNumber       uint8
}

func (dcmiCap *DCMICapParam_ManageabilityAccessAttributes) _isDCMICapParamter() {}

func (param *DCMICapParam_ManageabilityAccessAttributes) Pack() []byte {
	return []byte{}
}

func (param *DCMICapParam_ManageabilityAccessAttributes) Unpack(paramData []byte) error {
	if len(paramData) < 3 {
		return ErrUnpackedDataTooShortWith(len(paramData), 3)
	}

	param.PrimaryLANChannelNumber = paramData[0]
	param.SecondaryLANChannelNumber = paramData[1]
	param.SerialChannelNumber = paramData[2]

	return nil
}
func (param *DCMICapParam_ManageabilityAccessAttributes) Format() string {
	return fmt.Sprintf(`
    Manageability Access Attributes:
        Primary LAN channel number: %d is (%s)
        Secondary LAN channel number: %d is (%s)
        Serial channel number: %d is (%s)
`,
		param.PrimaryLANChannelNumber,
		formatBool(param.PrimaryLANChannelNumber != 0xFF, "available", "unavailable"),
		param.SecondaryLANChannelNumber,
		formatBool(param.SecondaryLANChannelNumber != 0xFF, "available", "unavailable"),
		param.SerialChannelNumber,
		formatBool(param.SerialChannelNumber != 0xFF, "available", "unavailable"),
	)
}

type DCMICapParam_EnhancedSystemPowerStatisticsAttributes struct {
	RollingAverageTimePeriodsSec []int
}

func (dcmiCap *DCMICapParam_EnhancedSystemPowerStatisticsAttributes) _isDCMICapParamter() {}

func (param *DCMICapParam_EnhancedSystemPowerStatisticsAttributes) Pack() []byte {
	return []byte{}
}

func (param *DCMICapParam_EnhancedSystemPowerStatisticsAttributes) Unpack(paramData []byte) error {
	if len(paramData) < 2 {
		return ErrUnpackedDataTooShortWith(len(paramData), 2)
	}

	periodsCount := int(paramData[0])
	if len(paramData) < 1+periodsCount {
		return ErrNotEnoughDataWith("rolling average time periods", len(paramData), 1+periodsCount)
	}

	periodsData, _, _ := unpackBytes(paramData, 1, periodsCount)
	for _, periodData := range periodsData {
		durationUnit := periodData >> 6
		durationNumber := periodData & 0x3F

		durationSec := 0
		switch durationUnit {
		case 0b00: // seconds
			durationSec = int(durationNumber)
		case 0b01: // minutes
			durationSec = int(durationNumber) * 60
		case 0b10: // hours
			durationSec = int(durationNumber) * 60 * 60
		case 0b11: // days
			durationSec = int(durationNumber) * 60 * 60 * 24
		}

		param.RollingAverageTimePeriodsSec = append(param.RollingAverageTimePeriodsSec, durationSec)
	}

	return nil
}
func (param *DCMICapParam_EnhancedSystemPowerStatisticsAttributes) Format() string {
	return fmt.Sprintf(`
    Enhanced System Power Statistics Attributes:

        Number of rolling average time periods: %d
        rolling average time periods: %v
`,
		len(param.RollingAverageTimePeriodsSec),
		param.RollingAverageTimePeriodsSec,
	)
}

type DCMIConfigParamSelector uint8

const (
	DCMIConfigParamSelector_ActivateDHCP           DCMIConfigParamSelector = 0x01
	DCMIConfigParamSelector_DiscoveryConfiguration DCMIConfigParamSelector = 0x02
	DCMIConfigParamSelector_DHCPTiming1            DCMIConfigParamSelector = 0x03
	DCMIConfigParamSelector_DHCPTiming2            DCMIConfigParamSelector = 0x04
	DCMIConfigParamSelector_DHCPTiming3            DCMIConfigParamSelector = 0x05
)

type DCMIConfig struct {
	ActivateDHCP           DCMIConfigParam_ActivateDHCP
	DiscoveryConfiguration DCMIConfigParam_DiscoveryConfiguration
	DHCPTiming1            DCMIConfigParam_DHCPTiming1
	DHCPTiming2            DCMIConfigParam_DHCPTiming2
	DHCPTiming3            DCMIConfigParam_DHCPTiming3
}

func (dcmiConfig *DCMIConfig) Format() string {
	return fmt.Sprintf(`
%s
%s
%s
%s`,
		dcmiConfig.DiscoveryConfiguration.Format(),
		dcmiConfig.DHCPTiming1.Format(),
		dcmiConfig.DHCPTiming2.Format(),
		dcmiConfig.DHCPTiming3.Format(),
	)
}

type DCMIConfigParam_ActivateDHCP struct {
	// Writing 01h to this parameter will trigger DHCP protocol restart using the latest parameter
	// settings, if DHCP is enabled. This can be used to ensure that the other DHCP configuration
	// parameters take effect immediately. Otherwise, the parameters may not take effect until the
	// next time the protocol restarts or a protocol timeout or lease expiration occurs. This is not a
	// non-volatile setting. It is only used to trigger a restart of the DHCP protocol.
	//
	// This parameter shall always return 0x00 when read.
	Activate bool
}

func (param *DCMIConfigParam_ActivateDHCP) _isDCMIConfigParameter() {}

func (param *DCMIConfigParam_ActivateDHCP) Pack() []byte {
	b := uint8(0)
	if param.Activate {
		b = 1
	}

	return []byte{b}
}

func (param *DCMIConfigParam_ActivateDHCP) Unpack(paramData []byte) error {
	if len(paramData) < 1 {
		return ErrUnpackedDataTooShortWith(len(paramData), 1)
	}

	param.Activate = paramData[0] == 1
	return nil
}

func (param *DCMIConfigParam_ActivateDHCP) Format() string {
	return fmt.Sprintf(`
Activate DHCP: %v
`,
		param.Activate,
	)
}

type DCMIConfigParam_DiscoveryConfiguration struct {
	RandomBackoffEnabled     bool
	IncludeDHCPOption60And43 bool
	IncludeDHCPOption12      bool
}

func (param *DCMIConfigParam_DiscoveryConfiguration) _isDCMIConfigParameter() {}

func (param *DCMIConfigParam_DiscoveryConfiguration) Pack() []byte {
	b := uint8(0)
	if param.RandomBackoffEnabled {
		b = setBit7(b)
	}

	if param.IncludeDHCPOption60And43 {
		b = setBit1(b)
	}

	if param.IncludeDHCPOption12 {
		b = setBit0(b)
	}

	return []byte{b}
}

func (param *DCMIConfigParam_DiscoveryConfiguration) Unpack(paramData []byte) error {
	if len(paramData) < 1 {
		return ErrUnpackedDataTooShortWith(len(paramData), 1)
	}

	param.RandomBackoffEnabled = isBit7Set(paramData[0])
	param.IncludeDHCPOption60And43 = isBit1Set(paramData[0])
	param.IncludeDHCPOption12 = isBit0Set(paramData[0])
	return nil
}

func (param *DCMIConfigParam_DiscoveryConfiguration) Format() string {
	return fmt.Sprintf(`
DHCP Discovery method:
    Random Backoff Enabled:             %v
    Include DHCPOption60AndOption43:    %v (Vendor class identifier using DCMI IANA, plus Vendor class
-specific Information)
    Include DHCPOption12:               %v (Management Controller ID String)
`,
		param.RandomBackoffEnabled,
		formatBool(param.IncludeDHCPOption60And43, "enabled", "disabled"),
		formatBool(param.IncludeDHCPOption12, "enabled", "disabled"),
	)
}

type DCMIConfigParam_DHCPTiming1 struct {
	// This parameter sets the amount of time between the first attempt to reach a server and the
	// second attempt to reach a server.
	//
	// Each time a message is sent the timeout interval between messages is incremented by
	// twice the current interval multiplied by a pseudo random number between zero and one
	// if random back-off is enabled, or multiplied by one if random back-off is disabled.
	//
	// The recommended default is four seconds
	InitialTimeoutIntervalSec uint8
}

func (param *DCMIConfigParam_DHCPTiming1) _isDCMIConfigParameter() {}

func (param *DCMIConfigParam_DHCPTiming1) Pack() []byte {
	return []byte{param.InitialTimeoutIntervalSec}
}

func (param *DCMIConfigParam_DHCPTiming1) Unpack(paramData []byte) error {
	if len(paramData) < 1 {
		return ErrUnpackedDataTooShortWith(len(paramData), 1)
	}

	param.InitialTimeoutIntervalSec = paramData[0]
	return nil
}

func (param *DCMIConfigParam_DHCPTiming1) Format() string {
	return fmt.Sprintf(`Initial timeout interval: %d seconds`,
		param.InitialTimeoutIntervalSec,
	)
}

type DCMIConfigParam_DHCPTiming2 struct {
	// This parameter determines the amount of time that must pass between the time that the
	// client initially tries to determine its address and the time that it decides that it cannot contact
	// a server. If the last lease is expired, the client will restart the protocol after the defined retry
	// interval. The recommended default timeout is two minutes. After server contact timeout, the
	// client must wait for Server Contact Retry Interval before attempting to contact the server
	// again.
	ServerContactTimeoutIntervalSec uint8
}

func (param *DCMIConfigParam_DHCPTiming2) _isDCMIConfigParameter() {}

func (param *DCMIConfigParam_DHCPTiming2) Pack() []byte {
	return []byte{param.ServerContactTimeoutIntervalSec}
}

func (param *DCMIConfigParam_DHCPTiming2) Unpack(paramData []byte) error {
	if len(paramData) < 1 {
		return ErrUnpackedDataTooShortWith(len(paramData), 1)
	}

	param.ServerContactTimeoutIntervalSec = paramData[0]
	return nil
}

func (param *DCMIConfigParam_DHCPTiming2) Format() string {
	return fmt.Sprintf(`Server contact timeout interval: %d seconds`,
		param.ServerContactTimeoutIntervalSec)
}

type DCMIConfigParam_DHCPTiming3 struct {
	// This is the period between DHCP retries after Server contact timeout interval expires. This
	// parameter determines the time that must pass after the client has determined that there is no
	// DHCP server present before it tries again to contact a DHCP server.
	//
	// The recommended default timeout is sixty-four seconds
	ServerContactRetryIntervalSec uint8
}

func (param *DCMIConfigParam_DHCPTiming3) _isDCMIConfigParameter() {}

func (param *DCMIConfigParam_DHCPTiming3) Pack() []byte {
	return []byte{param.ServerContactRetryIntervalSec}
}

func (param *DCMIConfigParam_DHCPTiming3) Unpack(paramData []byte) error {
	if len(paramData) < 1 {
		return ErrUnpackedDataTooShortWith(len(paramData), 1)
	}

	param.ServerContactRetryIntervalSec = paramData[0]
	return nil
}

func (param *DCMIConfigParam_DHCPTiming3) Format() string {
	return fmt.Sprintf(`Server contact retry interval: %d seconds`,
		param.ServerContactRetryIntervalSec,
	)
}
