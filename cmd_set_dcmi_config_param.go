package ipmi

// [DCMI specification v1.5] 6.1.2 Set DCMI Configuration Parameters
type SetDCMIConfigurationParamsRequest struct {
	ParamSelector DCMIConfigParamSelector
	SetSelector   uint8 // use 00h for parameters that only have one set
	ParamData     []byte
}

type SetDCMIConfigurationParamsResponse struct {
}

func (req *SetDCMIConfigurationParamsRequest) Pack() []byte {
	out := make([]byte, 3+len(req.ParamData))

	packUint8(GroupExtensionDCMI, out, 0)
	packUint8(uint8(req.ParamSelector), out, 1)
	packUint8(req.SetSelector, out, 2)
	packBytes(req.ParamData, out, 3)

	return out

}

func (req *SetDCMIConfigurationParamsRequest) Command() Command {
	return CommandSetDCMIConfigurationParams
}

func (res *SetDCMIConfigurationParamsResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *SetDCMIConfigurationParamsResponse) Unpack(msg []byte) error {
	if len(msg) < 1 {
		return ErrUnpackedDataTooShortWith(len(msg), 2)
	}

	if grpExt, _, _ := unpackUint8(msg, 0); grpExt != GroupExtensionDCMI {
		return ErrDCMIGroupExtensionIdentificationMismatch(GroupExtensionDCMI, grpExt)
	}

	return nil
}

func (res *SetDCMIConfigurationParamsResponse) Format() string {
	return ""
}

func (c *Client) SetDCMIConfigurationParams(paramSelector DCMIConfigParamSelector, setSelector uint8, param DCMIConfigParameter) (response *SetDCMIConfigurationParamsResponse, err error) {
	request := &SetDCMIConfigurationParamsRequest{
		ParamSelector: paramSelector,
		SetSelector:   setSelector,
		ParamData:     param.Pack(),
	}
	response = &SetDCMIConfigurationParamsResponse{}
	err = c.Exchange(request, response)
	return
}
