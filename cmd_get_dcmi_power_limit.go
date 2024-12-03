package ipmi

import (
	"fmt"
)

// [DCMI specification v1.5]: 6.6.2 Get Power Limit
type GetDCMIPowerLimitRequest struct {
}

type GetDCMIPowerLimitResponse struct {
	ExceptionAction DCMIExceptionAction
	// Power Limit Requested in Watts
	PowerLimitRequested uint16
	// Maximum time taken to limit the power after the platform power has reached
	// the power limit before the Exception Action will be taken.
	CorrectionTimeLimitMilliSec uint32
	// Management application Statistics Sampling period in seconds
	StatisticsSamplingPeriodSec uint16
}

func (req *GetDCMIPowerLimitRequest) Pack() []byte {
	return []byte{GroupExtensionDCMI, 0x00, 0x00}
}

func (req *GetDCMIPowerLimitRequest) Command() Command {
	return CommandGetDCMIPowerLimit
}

func (res *GetDCMIPowerLimitResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x80: "No Active Set Power Limit",
	}
}

func (res *GetDCMIPowerLimitResponse) Unpack(msg []byte) error {
	if len(msg) < 14 {
		return ErrUnpackedDataTooShortWith(len(msg), 14)
	}

	if CheckDCMIGroupExenstionMatch(msg[0]) != nil {
		return ErrDCMIGroupExtensionIdentificationMismatch(GroupExtensionDCMI, msg[0])
	}

	exceptionAction, _, _ := unpackUint8(msg, 3)
	res.ExceptionAction = DCMIExceptionAction(exceptionAction)
	res.PowerLimitRequested, _, _ = unpackUint16L(msg, 4)
	res.CorrectionTimeLimitMilliSec, _, _ = unpackUint32L(msg, 6)
	res.StatisticsSamplingPeriodSec, _, _ = unpackUint16L(msg, 12)

	return nil
}

func (res *GetDCMIPowerLimitResponse) Format() string {
	return "Power limit exception action : " + res.ExceptionAction.String() + "\n" +
		"Power limit requested        : " + fmt.Sprintf("%d", res.PowerLimitRequested) + " Watts\n" +
		"Correction Time Limit        : " + fmt.Sprintf("%d", res.CorrectionTimeLimitMilliSec) + " Milliseconds\n" +
		"Statistics Sampling period   : " + fmt.Sprintf("%d", res.StatisticsSamplingPeriodSec) + " Seconds\n"

}

// GetDCMIPowerLimit sends a DCMI "Get Power Reading" command.
// See [GetDCMIPowerLimitRequest] for details.
func (c *Client) GetDCMIPowerLimit() (response *GetDCMIPowerLimitResponse, err error) {
	request := &GetDCMIPowerLimitRequest{}
	response = &GetDCMIPowerLimitResponse{}
	err = c.Exchange(request, response)
	return
}
