package ipmi

import "fmt"

// GetDCMICapabilitiesInfoRequest provides version information for DCMI and information about
// the mandatory and optional DCMI capabilities that are available on the particular platform.
//
// The command is session-less and can be called similar to the Get Authentication Capability command.
// This command is a bare-metal provisioning command, and the availability of features does not imply
// the features are configured.
//
// [DCMI specification v1.5]: https://www.intel.com/content/dam/www/public/us/en/documents/technical-specifications/dcmi-v1-5-rev-spec.pdf
// 6.1.1 Get DCMI Capabilities Info Command
type GetDCMICapabilitiesInfoRequest struct {
	ParamSelector DCMICapParamSelector
}

type GetDCMICapabilitiesInfoResponse struct {
	MajorVersion  uint8
	MinorVersion  uint8
	ParamRevision uint8
	ParamData     []byte
}

func (req *GetDCMICapabilitiesInfoRequest) Pack() []byte {
	return []byte{GroupExtensionDCMI, byte(req.ParamSelector)}
}

func (req *GetDCMICapabilitiesInfoRequest) Command() Command {
	return CommandGetDCMICapabilitiesInfo
}

func (res *GetDCMICapabilitiesInfoResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetDCMICapabilitiesInfoResponse) Unpack(msg []byte) error {
	if len(msg) < 5 {
		return ErrUnpackedDataTooShortWith(len(msg), 5)
	}

	if grpExt, _, _ := unpackUint8(msg, 0); grpExt != GroupExtensionDCMI {
		return ErrDCMIGroupExtensionIdentificationMismatch(GroupExtensionDCMI, grpExt)
	}

	res.MajorVersion, _, _ = unpackUint8(msg, 1)
	res.MinorVersion, _, _ = unpackUint8(msg, 2)
	res.ParamRevision, _, _ = unpackUint8(msg, 3)
	res.ParamData, _, _ = unpackBytes(msg, 4, len(msg)-4)

	return nil
}

func (res *GetDCMICapabilitiesInfoResponse) Format() string {
	return fmt.Sprintf(`
  Major version  : %d
  Minor version  : %d
  Param revision : %d
	Param data     : %v`,
		res.MajorVersion,
		res.MinorVersion,
		res.ParamRevision,
		res.ParamData,
	)
}

func (c *Client) GetDCMICapabilitiesInfo(paramSelector DCMICapParamSelector) (response *GetDCMICapabilitiesInfoResponse, err error) {
	request := &GetDCMICapabilitiesInfoRequest{ParamSelector: paramSelector}
	response = &GetDCMICapabilitiesInfoResponse{}
	err = c.Exchange(request, response)
	return
}

func (c *Client) DiscoveryDCMICapabilities() (*DCMICapabilities, error) {
	out := &DCMICapabilities{}

	{
		request := &GetDCMICapabilitiesInfoRequest{ParamSelector: DCMICapParamSelector_SupportedDCMICapabilities}
		response := &GetDCMICapabilitiesInfoResponse{}
		if err := c.Exchange(request, response); err != nil {
			return nil, fmt.Errorf("failed to get DCMI SupportedDCMICapabilities, err: %s", err)
		}
		param := DCMICapParam_SupportedDCMICapabilities{}
		if err := param.Unpack(response.ParamData); err != nil {
			return nil, fmt.Errorf("failed to unpack DCMI SupportedDCMICapabilities, err: %s", err)
		}
		out.SupportedDCMICapabilities = param
	}

	{
		request := &GetDCMICapabilitiesInfoRequest{ParamSelector: DCMICapParamSelector_MandatoryPlatformAttributes}
		response := &GetDCMICapabilitiesInfoResponse{}
		if err := c.Exchange(request, response); err != nil {
			return nil, fmt.Errorf("failed to get DCMI MandatoryPlatformAttributes, err: %s", err)
		}
		param := DCMICapParam_MandatoryPlatformAttributes{}
		if err := param.Unpack(response.ParamData); err != nil {
			return nil, fmt.Errorf("failed to unpack DCMI MandatoryPlatformAttributes, err: %s", err)
		}
		out.MandatoryPlatformAttributes = param
	}

	{
		request := &GetDCMICapabilitiesInfoRequest{ParamSelector: DCMICapParamSelector_OptionalPlatformAttributes}
		response := &GetDCMICapabilitiesInfoResponse{}
		if err := c.Exchange(request, response); err != nil {
			return nil, fmt.Errorf("failed to get DCMI OptionalPlatformAttributes, err: %s", err)
		}
		param := DCMICapParam_OptionalPlatformAttributes{}
		if err := param.Unpack(response.ParamData); err != nil {
			return nil, fmt.Errorf("failed to unpack DCMI OptionalPlatformAttributes, err: %s", err)
		}
		out.OptionalPlatformAttributes = param
	}

	{
		request := &GetDCMICapabilitiesInfoRequest{ParamSelector: DCMICapParamSelector_ManageabilityAccessAttributes}
		response := &GetDCMICapabilitiesInfoResponse{}
		if err := c.Exchange(request, response); err != nil {
			return nil, fmt.Errorf("failed to get DCMI ManageabilityAccessAttributes, err: %s", err)
		}
		param := DCMICapParam_ManageabilityAccessAttributes{}
		if err := param.Unpack(response.ParamData); err != nil {
			return nil, fmt.Errorf("failed to unpack DCMI ManageabilityAccessAttributes, err: %s", err)
		}
		out.ManageabilityAccessAttributes = param
	}

	{
		request := &GetDCMICapabilitiesInfoRequest{ParamSelector: DCMICapParamSelector_EnhancedSystemPowerStatisticsAttributes}
		response := &GetDCMICapabilitiesInfoResponse{}
		if err := c.Exchange(request, response); err != nil {
			return nil, fmt.Errorf("failed to get DCMI EnhancedSystemPowerStatisticsAttributes, err: %s", err)
		}
		param := DCMICapParam_EnhancedSystemPowerStatisticsAttributes{}
		if err := param.Unpack(response.ParamData); err != nil {
			return nil, fmt.Errorf("failed to unpack DCMI EnhancedSystemPowerStatisticsAttributes, err: %s", err)
		}
		out.EnhancedSystemPowerStatisticsAttributes = param
	}

	return out, nil
}
