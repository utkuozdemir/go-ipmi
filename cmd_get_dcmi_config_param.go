package ipmi

import "fmt"

// [DCMI specification v1.5] 6.1.3 Get DCMI Configuration Parameters Command
type GetDCMIConfigurationParamsRequest struct {
	ParamSelector DCMIConfigParamSelector
	SetSelector   uint8 // use 00h for parameters that only have one set
}

type GetDCMIConfigurationParamsResponse struct {
	MajorVersion  uint8
	MinorVersion  uint8
	ParamRevision uint8
	ParamData     []byte
}

func (req *GetDCMIConfigurationParamsRequest) Pack() []byte {
	out := make([]byte, 3)

	packUint8(GroupExtensionDCMI, out, 0)
	packUint8(uint8(req.ParamSelector), out, 1)
	packUint8(req.SetSelector, out, 2)

	return out
}

func (req *GetDCMIConfigurationParamsRequest) Command() Command {
	return CommandGetDCMIConfigurationParams
}

func (res *GetDCMIConfigurationParamsResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetDCMIConfigurationParamsResponse) Unpack(msg []byte) error {
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

func (res *GetDCMIConfigurationParamsResponse) Format() string {
	return ""
}

func (c *Client) GetDCMIConfigurationParams(paramSelector DCMIConfigParamSelector, setSelector uint8) (response *GetDCMIConfigurationParamsResponse, err error) {
	request := &GetDCMIConfigurationParamsRequest{
		ParamSelector: paramSelector,
		SetSelector:   setSelector,
	}
	response = &GetDCMIConfigurationParamsResponse{}
	err = c.Exchange(request, response)
	return
}

func (c *Client) GetDCMIConfigurations() (*DCMIConfig, error) {
	out := &DCMIConfig{}

	{
		request := &GetDCMIConfigurationParamsRequest{ParamSelector: DCMIConfigParamSelector_ActivateDHCP}
		response := &GetDCMIConfigurationParamsResponse{}
		if err := c.Exchange(request, response); err != nil {
			return nil, fmt.Errorf("failed to get DCMI Config ActivateDHCP, err: %s", err)
		}
		param := DCMIConfigParam_ActivateDHCP{}
		if err := param.Unpack(response.ParamData); err != nil {
			return nil, fmt.Errorf("failed to unpack DCMI Config ActivateDHCP, err: %s", err)
		}
		out.ActivateDHCP = param
	}

	{
		request := &GetDCMIConfigurationParamsRequest{ParamSelector: DCMIConfigParamSelector_DiscoveryConfiguration}
		response := &GetDCMIConfigurationParamsResponse{}
		if err := c.Exchange(request, response); err != nil {
			return nil, fmt.Errorf("failed to get DCMI Config DiscoveryConfiguration, err: %s", err)
		}
		param := DCMIConfigParam_DiscoveryConfiguration{}
		if err := param.Unpack(response.ParamData); err != nil {
			return nil, fmt.Errorf("failed to unpack DCMI Config DiscoveryConfiguration, err: %s", err)
		}
		out.DiscoveryConfiguration = param
	}

	{
		request := &GetDCMIConfigurationParamsRequest{ParamSelector: DCMIConfigParamSelector_DHCPTiming1}
		response := &GetDCMIConfigurationParamsResponse{}
		if err := c.Exchange(request, response); err != nil {
			return nil, fmt.Errorf("failed to get DCMI Config DHCPTiming1, err: %s", err)
		}
		param := DCMIConfigParam_DHCPTiming1{}
		if err := param.Unpack(response.ParamData); err != nil {
			return nil, fmt.Errorf("failed to unpack DCMI Config DHCPTiming1, err: %s", err)
		}
		out.DHCPTiming1 = param
	}

	{
		request := &GetDCMIConfigurationParamsRequest{ParamSelector: DCMIConfigParamSelector_DHCPTiming2}
		response := &GetDCMIConfigurationParamsResponse{}
		if err := c.Exchange(request, response); err != nil {
			return nil, fmt.Errorf("failed to get DCMI Config DHCPTiming2, err: %s", err)
		}
		param := DCMIConfigParam_DHCPTiming2{}
		if err := param.Unpack(response.ParamData); err != nil {
			return nil, fmt.Errorf("failed to unpack DCMI Config DHCPTiming2, err: %s", err)
		}
		out.DHCPTiming2 = param
	}

	{
		request := &GetDCMIConfigurationParamsRequest{ParamSelector: DCMIConfigParamSelector_DHCPTiming3}
		response := &GetDCMIConfigurationParamsResponse{}
		if err := c.Exchange(request, response); err != nil {
			return nil, fmt.Errorf("failed to get DCMI Config DHCPTiming3, err: %s", err)
		}
		param := DCMIConfigParam_DHCPTiming3{}
		if err := param.Unpack(response.ParamData); err != nil {
			return nil, fmt.Errorf("failed to unpack DCMI Config DHCPTiming3, err: %s", err)
		}
		out.DHCPTiming3 = param
	}

	return out, nil
}
