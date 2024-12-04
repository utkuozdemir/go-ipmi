package ipmi

import (
	"fmt"
)

// 30.4 Get PEF Configuration Parameters Command
type GetPEFConfigParametersRequest struct {
	// [7] - 1b = get parameter revision only. 0b = get parameter
	// [6:0] - Parameter selector
	GetRevisionOnly bool
	ParamSelector   PEFConfigParamSelector

	SetSelector   uint8 // 00h if parameter does not require a Set Selector
	BlockSelector uint8 // 00h if parameter does not require a block number
}

type GetPEFConfigParametersResponse struct {
	// Parameter revision.
	//
	// Format:
	//  - MSN = present revision.
	//  - LSN = oldest revision parameter is backward compatible with.
	//  - 11h for parameters in this specification.
	Revision uint8

	// ConfigData data bytes are not returned when the 'get parameter revision only' bit is 1b.
	ConfigData []byte
}

func (req *GetPEFConfigParametersRequest) Command() Command {
	return CommandGetPEFConfigParameters
}

func (req *GetPEFConfigParametersRequest) Pack() []byte {
	// empty request data

	out := make([]byte, 3)

	b0 := uint8(req.ParamSelector) & 0x3f
	if req.GetRevisionOnly {
		b0 = setBit7(b0)
	}
	packUint8(b0, out, 0)
	packUint8(req.SetSelector, out, 1)
	packUint8(req.BlockSelector, out, 2)

	return out
}

func (res *GetPEFConfigParametersResponse) Unpack(msg []byte) error {
	if len(msg) < 1 {
		return ErrUnpackedDataTooShort
	}

	res.Revision = msg[0]

	if len(msg) > 1 {
		res.ConfigData, _, _ = unpackBytes(msg, 1, len(msg)-1)
	}

	return nil
}

func (r *GetPEFConfigParametersResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x80: "parameter not supported",
	}
}

func (res *GetPEFConfigParametersResponse) Format() string {
	return fmt.Sprintf(`
Parameter Revision           : %#02x (%d)
Configuration Parameter Data : %# 02x`,
		res.Revision, res.Revision,
		res.ConfigData,
	)
}

func (c *Client) GetPEFConfigParameters(getRevisionOnly bool, paramSelector PEFConfigParamSelector, setSelector uint8, blockSelector uint8) (response *GetPEFConfigParametersResponse, err error) {
	request := &GetPEFConfigParametersRequest{
		GetRevisionOnly: getRevisionOnly,
		ParamSelector:   paramSelector,
		SetSelector:     setSelector,
		BlockSelector:   blockSelector,
	}
	response = &GetPEFConfigParametersResponse{}
	err = c.Exchange(request, response)
	return
}

type PEFConfig struct {
	SetInProgress                    *PEFConfigParam_SetInProgress
	Control                          *PEFConfigParam_Control
	ActionGlobalControl              *PEFConfigParam_ActionGlobalControl
	StartupDelay                     *PEFConfigParam_StartupDelay
	AlertStartupDelay                *PEFConfigParam_AlertStartupDelay
	NumberOfEventFilters             *PEFConfigParam_NumberOfEventFilters
	EventFilters                     []*PEFConfigParam_EventFilter
	EventFiltersData1                []*PEFConfigParam_EventFilterData1
	NumberOfAlertPolicies            *PEFConfigParam_NumberOfAlertPolicies
	AlertPolicies                    []*PEFConfigParam_AlertPolicy
	SystemGUID                       *PEFConfigParam_SystemGUID
	NumberOfAlertStrings             *PEFConfigParam_NumberOfAlertStrings
	AlertStringKeys                  []*PEFConfigParam_AlertStringKeys
	AlertStrings                     []*PEFConfigParam_AlertStrings
	NumberOfGroupControlTableEntries *PEFConfigParam_NumberOfGroupControlTableEntries
	GroupControlTable                *PEFConfigParam_GroupControlTable
}

func (c *Client) GetPEFConfigForParameter(param PEFConfigParameter, setSelector uint8, blockSelector uint8) error {
	paramSelector := param.PEFConfigParamSelector()

	res, err := c.GetPEFConfigParameters(false, paramSelector, setSelector, blockSelector)
	if err != nil {
		return fmt.Errorf("GetPEFConfigParameters for param (%s) failed, err: %s", paramSelector, err)
	}

	if err := param.Unpack(res.ConfigData); err != nil {
		return fmt.Errorf("unpack failed for param (%s), err: %s", paramSelector, err)
	}

	c.Debugf("PEF Config for param selector (%s)\n%s", paramSelector, param.Format())
	return nil
}

func (c *Client) GetPEFConfig() (pefConfig *PEFConfig, err error) {
	pefConfig = &PEFConfig{}

	{
		param := &PEFConfigParam_Control{}
		if err := c.GetPEFConfigForParameter(param, 0, 0); err != nil {
			return nil, err
		}
		pefConfig.Control = param
	}

	{
		param := &PEFConfigParam_ActionGlobalControl{}
		if err := c.GetPEFConfigForParameter(param, 0, 0); err != nil {
			return nil, err
		}
		pefConfig.ActionGlobalControl = param
	}

	{
		param := &PEFConfigParam_StartupDelay{}
		if err := c.GetPEFConfigForParameter(param, 0, 0); err != nil {
			return nil, err
		}
		pefConfig.StartupDelay = param
	}

	{
		param := &PEFConfigParam_AlertStartupDelay{}
		if err := c.GetPEFConfigForParameter(param, 0, 0); err != nil {
			return nil, err
		}
		pefConfig.AlertStartupDelay = param
	}

	{
		param := &PEFConfigParam_NumberOfEventFilters{}
		if err := c.GetPEFConfigForParameter(param, 0, 0); err != nil {
			return nil, err
		}
		pefConfig.NumberOfEventFilters = param
	}

	{
		for i := uint8(1); i <= pefConfig.NumberOfEventFilters.Value; i++ {
			param := &PEFConfigParam_EventFilter{}
			if err := c.GetPEFConfigForParameter(param, i, 0); err != nil {
				return nil, fmt.Errorf("get event filter number (%d) failed, err: %s", i, err)
			}
			pefConfig.EventFilters = append(pefConfig.EventFilters, param)
		}
	}

	{
		for i := uint8(1); i <= pefConfig.NumberOfEventFilters.Value; i++ {
			param := &PEFConfigParam_EventFilterData1{}
			if err := c.GetPEFConfigForParameter(param, i, 0); err != nil {
				return nil, fmt.Errorf("get event filter number (%d) failed, err: %s", i, err)
			}
			pefConfig.EventFiltersData1 = append(pefConfig.EventFiltersData1, param)
		}
	}

	{
		// ipmitool pef
		param := &PEFConfigParam_NumberOfAlertPolicies{}
		if err := c.GetPEFConfigForParameter(param, 0, 0); err != nil {
			return nil, err
		}
		pefConfig.NumberOfAlertPolicies = param
	}

	{
		for i := uint8(1); i < pefConfig.NumberOfAlertPolicies.Value; i++ {
			param := &PEFConfigParam_AlertPolicy{}

			if err := c.GetPEFConfigForParameter(param, i, 0); err != nil {
				return nil, fmt.Errorf("get event filter number (%d) failed, err: %s", i, err)
			}
			pefConfig.AlertPolicies = append(pefConfig.AlertPolicies, param)
		}
	}

	{
		param := &PEFConfigParam_SystemGUID{}
		if err := c.GetPEFConfigForParameter(param, 0, 0); err != nil {
			return nil, err
		}
		pefConfig.SystemGUID = param
	}

	{
		param := &PEFConfigParam_NumberOfAlertStrings{}
		if err := c.GetPEFConfigForParameter(param, 0, 0); err != nil {
			return nil, err
		}
		pefConfig.NumberOfAlertStrings = param
	}

	{
		for i := uint8(1); i <= pefConfig.NumberOfAlertStrings.Value; i++ {
			param := &PEFConfigParam_AlertStringKeys{}

			//
			setSelctor := i
			blockSelector := uint8(0)
			if err := c.GetPEFConfigForParameter(param, setSelctor, blockSelector); err != nil {
				return nil, fmt.Errorf("get alert strings number (%d) failed, err: %s", i, err)
			}
			pefConfig.AlertStringKeys = append(pefConfig.AlertStringKeys, param)
		}
	}

	{
		for i := uint8(1); i < pefConfig.NumberOfAlertStrings.Value; i++ {
			param := &PEFConfigParam_AlertStrings{}
			setSelctor := i
			blockSelector := uint8(1)
			if err := c.GetPEFConfigForParameter(param, setSelctor, blockSelector); err != nil {
				return nil, fmt.Errorf("get alert strings number (%d) failed, err: %s", i, err)
			}
			pefConfig.AlertStrings = append(pefConfig.AlertStrings, param)
		}
	}

	// {
	// 	param := &PEFConfigParam_NumberOfGroupControlTableEntries{}
	// 	if err := c.getPEFConfigFor(param); err != nil {
	// 		return nil, err
	// 	}
	// 	pefConfig.NumberOfGroupControlTableEntries = param
	// }

	// {
	// 	param := &PEFConfigParam_GroupControlTable{}
	// 	if err := c.getPEFConfigFor(param); err != nil {
	// 		return nil, err
	// 	}
	// 	pefConfig.GroupControlTable = param
	// }

	return pefConfig, nil
}
