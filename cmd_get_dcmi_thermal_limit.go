package ipmi

import (
	"errors"
	"fmt"
)

// [DCMI specification v1.5]: 6.7.1 Get Thermal Limit Command
type GetDCMIThermalLimitRequest struct {
	EntityID       EntityID // Entity ID = 37h or 40h (Inlet Temperature)
	EntityInstance EntityInstance
}

type GetDCMIThermalLimitResponse struct {
	ExceptionAction_PowerOffAndLogSEL bool
	ExceptionAction_LogSELOnly        bool // ignored if ExceptionAction_PowerOffAndLogSEL is true

	// Temperature Limit set in units defined by the SDR record.
	// Note: the management controller is not required to check this parameter for validity against the SDR contents.
	TemperatureLimit uint8
	// Interval in seconds over which the temperature must continuously be sampled as exceeding the set limit
	// before the specified Exception Action will be taken.
	// Samples are taken at the rate specified by the sampling frequency value in parameter #5 of the DCMI Capabilities // parameters (see Table 6-3, DCMI Capabilities Parameters).
	ExceptionTimeSec uint16
}

func (req *GetDCMIThermalLimitRequest) Pack() []byte {
	return []byte{GroupExtensionDCMI, byte(req.EntityID), byte(req.EntityInstance)}
}

func (req *GetDCMIThermalLimitRequest) Command() Command {
	return CommandGetDCMIThermalLimit
}

func (res *GetDCMIThermalLimitResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetDCMIThermalLimitResponse) Unpack(msg []byte) error {
	if len(msg) < 5 {
		return ErrUnpackedDataTooShortWith(len(msg), 5)
	}

	if CheckDCMIGroupExenstionMatch(msg[0]) != nil {
		return ErrDCMIGroupExtensionIdentificationMismatch(GroupExtensionDCMI, msg[0])
	}

	b1, _, _ := unpackUint8(msg, 1)
	res.ExceptionAction_PowerOffAndLogSEL = isBit6Set(b1)
	res.ExceptionAction_LogSELOnly = isBit5Set(b1)

	res.TemperatureLimit, _, _ = unpackUint8(msg, 2)
	res.ExceptionTimeSec, _, _ = unpackUint16L(msg, 3)

	return nil
}

func (res *GetDCMIThermalLimitResponse) Format() string {
	return fmt.Sprintf(`
    Exception Actions, taken if the Temperature Limit exceeded:
        Hard Power Off system and log event:  %s
        Log event to SEL only:                %s
    Temperature Limit                         %d degrees
    Exception Time                            %d seconds`,
		formatBool(res.ExceptionAction_PowerOffAndLogSEL, "active", "inactive"),
		formatBool(res.ExceptionAction_LogSELOnly, "active", "inactive"),
		res.TemperatureLimit,
		res.ExceptionTimeSec)
}

func (c *Client) GetDCMIThermalLimit(entityID EntityID, entityInstance EntityInstance) (response *GetDCMIThermalLimitResponse, err error) {
	if uint8(entityID) != 0x37 && uint8(entityID) != 0x40 {
		return nil, errors.New("only Inlet Temperature entityID (0x37 or 0x40) is supported")
	}
	request := &GetDCMIThermalLimitRequest{
		EntityID:       entityID,
		EntityInstance: entityInstance,
	}
	response = &GetDCMIThermalLimitResponse{}
	err = c.Exchange(request, response)
	return
}
