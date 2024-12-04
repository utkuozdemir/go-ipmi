package ipmi

import "fmt"

type PEFConfigParamSelector uint8

const (
	PEFConfigParamSelector_SetInProgress                    PEFConfigParamSelector = 0x00
	PEFConfigParamSelector_Control                          PEFConfigParamSelector = 0x01
	PEFConfigParamSelector_ActionGlobalControl              PEFConfigParamSelector = 0x02
	PEFConfigParamSelector_StartupDelay                     PEFConfigParamSelector = 0x03
	PEFConfigParamSelector_AlertStartDelay                  PEFConfigParamSelector = 0x04
	PEFConfigParamSelector_NumberOfEventFilter              PEFConfigParamSelector = 0x05
	PEFConfigParamSelector_EventFilter                      PEFConfigParamSelector = 0x06
	PEFConfigParamSelector_EventFilterData1                 PEFConfigParamSelector = 0x07
	PEFConfigParamSelector_NumberOfAlertPolicyEntries       PEFConfigParamSelector = 0x08
	PEFConfigParamSelector_AlertPolicy                      PEFConfigParamSelector = 0x09
	PEFConfigParamSelector_SystemGUID                       PEFConfigParamSelector = 0x0a
	PEFConfigParamSelector_NumberOfAlertStrings             PEFConfigParamSelector = 0x0b
	PEFConfigParamSelector_AlertStringKeys                  PEFConfigParamSelector = 0x0c
	PEFConfigParamSelector_AlertStrings                     PEFConfigParamSelector = 0x0d
	PEFConfigParamSelector_NumberOfGroupControlTableEntries PEFConfigParamSelector = 0x0e
	PEFConfigParamSelector_GroupControlTable                PEFConfigParamSelector = 0x0f

	// 96:127
	// OEM Parameters (optional. Non-volatile or volatile as specified by OEM)
	// This range is available for special OEM configuration parameters.
	// The OEM is identified according to the Manufacturer ID field returned by the Get Device ID command.
)

func (p PEFConfigParamSelector) String() string {
	m := map[PEFConfigParamSelector]string{
		PEFConfigParamSelector_SetInProgress:                    "SetInProgress",
		PEFConfigParamSelector_Control:                          "Control",
		PEFConfigParamSelector_ActionGlobalControl:              "ActionGlobalControl",
		PEFConfigParamSelector_StartupDelay:                     "StartupDelay",
		PEFConfigParamSelector_AlertStartDelay:                  "AlertStartDelay",
		PEFConfigParamSelector_NumberOfEventFilter:              "NumberOfEventFilter",
		PEFConfigParamSelector_EventFilter:                      "EventFilterTable",
		PEFConfigParamSelector_EventFilterData1:                 "EventFilterTableData1",
		PEFConfigParamSelector_NumberOfAlertPolicyEntries:       "NumberOfAlertPolicyEntries",
		PEFConfigParamSelector_AlertPolicy:                      "AlertPolicyTable",
		PEFConfigParamSelector_SystemGUID:                       "SystemGUID",
		PEFConfigParamSelector_NumberOfAlertStrings:             "NumberOfAlertStrings",
		PEFConfigParamSelector_AlertStringKeys:                  "AlertStringKeys",
		PEFConfigParamSelector_AlertStrings:                     "AlertStrings",
		PEFConfigParamSelector_NumberOfGroupControlTableEntries: "NumberOfGroupControlTableEntries",
		PEFConfigParamSelector_GroupControlTable:                "GroupControlTable",
	}

	if s, ok := m[p]; ok {
		return s
	}

	return fmt.Sprintf("Unknown (%#02x)", p)
}

type PEFConfigParameter interface {
	PEFConfigParamSelector() PEFConfigParamSelector
	Unpack(data []byte) error
	Pack() []byte
	Format() string
}

var (
	_ PEFConfigParameter = (*PEFConfigParam_SetInProgress)(nil)
	_ PEFConfigParameter = (*PEFConfigParam_Control)(nil)
	_ PEFConfigParameter = (*PEFConfigParam_ActionGlobalControl)(nil)
	_ PEFConfigParameter = (*PEFConfigParam_StartupDelay)(nil)
	_ PEFConfigParameter = (*PEFConfigParam_AlertStartupDelay)(nil)
	_ PEFConfigParameter = (*PEFConfigParam_NumberOfEventFilters)(nil)
	_ PEFConfigParameter = (*PEFConfigParam_EventFilter)(nil)
	_ PEFConfigParameter = (*PEFConfigParam_NumberOfAlertPolicies)(nil)
	_ PEFConfigParameter = (*PEFConfigParam_AlertPolicy)(nil)
	_ PEFConfigParameter = (*PEFConfigParam_SystemGUID)(nil)
	_ PEFConfigParameter = (*PEFConfigParam_NumberOfAlertStrings)(nil)
	_ PEFConfigParameter = (*PEFConfigParam_AlertStringKeys)(nil)
	_ PEFConfigParameter = (*PEFConfigParam_AlertStrings)(nil)
	_ PEFConfigParameter = (*PEFConfigParam_NumberOfGroupControlTableEntries)(nil)
	_ PEFConfigParameter = (*PEFConfigParam_GroupControlTable)(nil)
)

type PEFConfigParam_SetInProgress SetInProgress

func (*PEFConfigParam_SetInProgress) PEFConfigParamSelector() PEFConfigParamSelector {
	return PEFConfigParamSelector_SetInProgress
}

func (param *PEFConfigParam_SetInProgress) Unpack(data []byte) error {
	if len(data) < 1 {
		return ErrUnpackedDataTooShortWith(len(data), 1)
	}

	*param = PEFConfigParam_SetInProgress(data[0])
	return nil
}

func (param *PEFConfigParam_SetInProgress) Pack() []byte {
	return []byte{byte(*param)}
}

func (param *PEFConfigParam_SetInProgress) Format() string {
	return fmt.Sprintf("SetInProgress=%v", *param)
}

type PEFConfigParam_Control struct {
	EnablePEFAlertStartupDelay bool
	EnablePEFStartupDelay      bool
	EnableEventMessage         bool
	EnablePEF                  bool
}

func (*PEFConfigParam_Control) PEFConfigParamSelector() PEFConfigParamSelector {
	return PEFConfigParamSelector_Control
}

func (param *PEFConfigParam_Control) Unpack(data []byte) error {
	if len(data) < 1 {
		return ErrUnpackedDataTooShortWith(len(data), 1)
	}

	param.EnablePEFAlertStartupDelay = isBit3Set(data[0])
	param.EnablePEFStartupDelay = isBit2Set(data[0])
	param.EnableEventMessage = isBit1Set(data[0])
	param.EnablePEF = isBit0Set(data[0])

	return nil
}

func (param *PEFConfigParam_Control) Pack() []byte {
	b := uint8(0x00)

	b = setOrClearBit3(b, param.EnablePEFAlertStartupDelay)
	b = setOrClearBit2(b, param.EnablePEFStartupDelay)
	b = setOrClearBit1(b, param.EnableEventMessage)
	b = setOrClearBit0(b, param.EnablePEF)

	return []byte{b}
}

func (param *PEFConfigParam_Control) Format() string {
	return fmt.Sprintf(`
PEF startup delay        : %s
Alert startup delay      : %s
PEF event messages       : %s
PEF                      : %s
`,
		formatBool(param.EnablePEFAlertStartupDelay, "enabled", "disabled"),
		formatBool(param.EnablePEFStartupDelay, "enabled", "disabled"),
		formatBool(param.EnableEventMessage, "enabled", "disabled"),
		formatBool(param.EnablePEF, "enabled", "disabled"),
	)
}

type PEFConfigParam_ActionGlobalControl struct {
	DiagnosticInterruptEnabled bool
	OEMActionEnabled           bool
	PowerCycleActionEnabled    bool
	ResetActionEnabled         bool
	PowerDownActionEnabled     bool
	AlertActionEnabled         bool
}

func (*PEFConfigParam_ActionGlobalControl) PEFConfigParamSelector() PEFConfigParamSelector {
	return PEFConfigParamSelector_ActionGlobalControl
}

func (param *PEFConfigParam_ActionGlobalControl) Unpack(data []byte) error {
	if len(data) < 1 {
		return ErrUnpackedDataTooShortWith(len(data), 1)
	}

	param.DiagnosticInterruptEnabled = isBit5Set(data[0])
	param.OEMActionEnabled = isBit4Set(data[0])
	param.PowerCycleActionEnabled = isBit3Set(data[0])
	param.ResetActionEnabled = isBit2Set(data[0])
	param.PowerDownActionEnabled = isBit1Set(data[0])
	param.AlertActionEnabled = isBit0Set(data[0])

	return nil
}

func (param *PEFConfigParam_ActionGlobalControl) Pack() []byte {
	b := uint8(0x00)

	b = setOrClearBit5(b, param.DiagnosticInterruptEnabled)
	b = setOrClearBit4(b, param.OEMActionEnabled)
	b = setOrClearBit3(b, param.PowerCycleActionEnabled)
	b = setOrClearBit2(b, param.ResetActionEnabled)
	b = setOrClearBit1(b, param.PowerDownActionEnabled)
	b = setOrClearBit0(b, param.AlertActionEnabled)

	return []byte{b}
}

func (param *PEFConfigParam_ActionGlobalControl) Format() string {
	return fmt.Sprintf(`
Diagnostic-interrupt     : %s
OEM-defined              : %s
Power-cycle              : %s
Reset                    : %s
Power-off                : %s
Alert                    : %s
`,
		formatBool(param.DiagnosticInterruptEnabled, "active", "inactive"),
		formatBool(param.OEMActionEnabled, "active", "inactive"),
		formatBool(param.PowerCycleActionEnabled, "active", "inactive"),
		formatBool(param.ResetActionEnabled, "active", "inactive"),
		formatBool(param.PowerDownActionEnabled, "active", "inactive"),
		formatBool(param.AlertActionEnabled, "active", "inactive"),
	)
}

type PEFConfigParam_StartupDelay struct {
	DelaySec uint8
}

func (*PEFConfigParam_StartupDelay) PEFConfigParamSelector() PEFConfigParamSelector {
	return PEFConfigParamSelector_StartupDelay
}

func (param *PEFConfigParam_StartupDelay) Unpack(data []byte) error {
	if len(data) < 1 {
		return ErrUnpackedDataTooShortWith(len(data), 1)
	}

	param.DelaySec = data[0]

	return nil
}

func (param *PEFConfigParam_StartupDelay) Pack() []byte {
	return []byte{param.DelaySec}
}

func (param *PEFConfigParam_StartupDelay) Format() string {
	return fmt.Sprintf("DelaySec=%v", param.DelaySec)
}

type PEFConfigParam_AlertStartupDelay struct {
	DelaySec uint8
}

func (*PEFConfigParam_AlertStartupDelay) PEFConfigParamSelector() PEFConfigParamSelector {
	return PEFConfigParamSelector_AlertStartDelay
}

func (param *PEFConfigParam_AlertStartupDelay) Unpack(data []byte) error {
	if len(data) < 1 {
		return ErrUnpackedDataTooShortWith(len(data), 1)
	}

	param.DelaySec = data[0]

	return nil
}

func (param *PEFConfigParam_AlertStartupDelay) Pack() []byte {
	return []byte{param.DelaySec}
}

func (param *PEFConfigParam_AlertStartupDelay) Format() string {
	return fmt.Sprintf("DelaySec=%v", param.DelaySec)
}

// Number of event filters supported. 1-based.
// This parameter does not need to be supported if Alerting is not supported.
type PEFConfigParam_NumberOfEventFilters struct {
	Value uint8
}

func (*PEFConfigParam_NumberOfEventFilters) PEFConfigParamSelector() PEFConfigParamSelector {
	return PEFConfigParamSelector_NumberOfEventFilter
}

func (param *PEFConfigParam_NumberOfEventFilters) Unpack(data []byte) error {
	if len(data) < 1 {
		return ErrUnpackedDataTooShortWith(len(data), 1)
	}

	param.Value = data[0]
	return nil
}

func (param *PEFConfigParam_NumberOfEventFilters) Pack() []byte {
	return []byte{param.Value}
}

func (param *PEFConfigParam_NumberOfEventFilters) Format() string {
	return fmt.Sprintf("Number of Event Filters: %d\n", param.Value)
}

type PEFConfigParam_EventFilter struct {
	EntryNumber uint8
	Entry       *PEFEventFilter
}

func (*PEFConfigParam_EventFilter) PEFConfigParamSelector() PEFConfigParamSelector {
	return PEFConfigParamSelector_EventFilter
}

func (param *PEFConfigParam_EventFilter) Unpack(data []byte) error {
	if len(data) < 21 {
		return ErrUnpackedDataTooShortWith(len(data), 21)
	}

	param.EntryNumber = data[0]

	entry := &PEFEventFilter{}
	if err := entry.Unpack(data[1:21]); err != nil {
		return fmt.Errorf("unpack entry failed, err: %s", err)
	}
	param.Entry = entry

	return nil
}

func (param *PEFConfigParam_EventFilter) Pack() []byte {
	entryData := param.Entry.Pack()
	out := make([]byte, len(entryData))

	out[0] = param.EntryNumber
	packBytes(entryData, out, 1)

	return out
}

func (param *PEFConfigParam_EventFilter) Format() string {
	return fmt.Sprintf(`
EntryNumber:   %d
Entry:
%v
`, param.EntryNumber, param.Entry.Format())
}

type PEFConfigParam_EventFilterData1 struct {
	FilterNumber uint8
	Data1        byte
}

func (*PEFConfigParam_EventFilterData1) PEFConfigParamSelector() PEFConfigParamSelector {
	return PEFConfigParamSelector_EventFilterData1
}

func (param *PEFConfigParam_EventFilterData1) Unpack(data []byte) error {
	if len(data) < 2 {
		return ErrUnpackedDataTooShortWith(len(data), 21)
	}

	param.FilterNumber = data[0]
	param.Data1 = data[1]

	return nil
}

func (param *PEFConfigParam_EventFilterData1) Pack() []byte {
	out := make([]byte, 21)

	out[0] = param.FilterNumber
	out[1] = param.Data1

	return out
}

func (param *PEFConfigParam_EventFilterData1) Format() string {
	return fmt.Sprintf(`
FilterNumber:   %d
Data1:    %v
`, param.FilterNumber, param.Data1)
}

type PEFConfigParam_NumberOfAlertPolicies struct {
	Value uint8
}

func (*PEFConfigParam_NumberOfAlertPolicies) PEFConfigParamSelector() PEFConfigParamSelector {
	return PEFConfigParamSelector_NumberOfAlertPolicyEntries
}

func (param *PEFConfigParam_NumberOfAlertPolicies) Unpack(data []byte) error {
	if len(data) < 1 {
		return ErrUnpackedDataTooShortWith(len(data), 1)
	}

	param.Value = data[0]

	return nil
}

func (param *PEFConfigParam_NumberOfAlertPolicies) Pack() []byte {
	return []byte{param.Value}
}

func (param *PEFConfigParam_NumberOfAlertPolicies) Format() string {
	return fmt.Sprintf("Number of Alert Policy Entries: %d\n", param.Value)
}

type PEFConfigParam_AlertPolicy struct {
	EntryNumber uint8
	Entry       *PEFAlertPolicy
}

func (*PEFConfigParam_AlertPolicy) PEFConfigParamSelector() PEFConfigParamSelector {
	return PEFConfigParamSelector_AlertPolicy
}

func (param *PEFConfigParam_AlertPolicy) Unpack(data []byte) error {
	if len(data) < 4 {
		return ErrUnpackedDataTooShortWith(len(data), 4)
	}

	param.EntryNumber = data[0]

	b := &PEFAlertPolicy{}
	if err := b.Unpack(data[1:]); err != nil {
		return err
	}
	param.Entry = b

	return nil
}

func (param *PEFConfigParam_AlertPolicy) Pack() []byte {
	entryData := param.Entry.Pack()

	out := make([]byte, 1+len(entryData))

	out[0] = param.EntryNumber
	packBytes(entryData, out, 1)

	return out
}

func (param *PEFConfigParam_AlertPolicy) Format() string {
	return fmt.Sprintf(`
EntryNumber:   %d
Entry:
%v
`, param.EntryNumber, param.Entry.Format())
}

// Used to fill in the GUID field in a PET Trap.
type PEFConfigParam_SystemGUID struct {
	// [7:1] - reserved
	// [0]
	//	1b = BMC uses following value in PET Trap.
	//	0b = BMC ignores following value and uses value returned from Get System GUID command instead.
	UseGUID bool
	GUID    [16]byte
}

func (*PEFConfigParam_SystemGUID) PEFConfigParamSelector() PEFConfigParamSelector {
	return PEFConfigParamSelector_SystemGUID
}

func (param *PEFConfigParam_SystemGUID) Unpack(configData []byte) error {
	if len(configData) < 17 {
		return ErrUnpackedDataTooShortWith(len(configData), 17)
	}

	param.UseGUID = isBit0Set(configData[0])
	param.GUID = array16(configData[1:17])
	return nil
}

func (param *PEFConfigParam_SystemGUID) Pack() []byte {
	out := make([]byte, 17)

	out[0] = setOrClearBit0(0x00, param.UseGUID)
	copy(out[1:], param.GUID[:])

	return out
}

func (param *PEFConfigParam_SystemGUID) Format() string {
	u, err := ParseGUID(param.GUID[:], GUIDModeSMBIOS)
	if err != nil {
		return fmt.Sprintf("<invalid UUID bytes> (%s)", err)
	}

	out := ""
	out += fmt.Sprintf("UseGUID:   %v\n", param.UseGUID)
	out += fmt.Sprintf("GUID:      %s\n", u.String())
	return out
}

type PEFConfigParam_NumberOfAlertStrings struct {
	Value uint8
}

func (*PEFConfigParam_NumberOfAlertStrings) PEFConfigParamSelector() PEFConfigParamSelector {
	return PEFConfigParamSelector_NumberOfAlertStrings
}

func (param *PEFConfigParam_NumberOfAlertStrings) Unpack(data []byte) error {
	if len(data) < 1 {
		return ErrUnpackedDataTooShortWith(len(data), 1)
	}

	param.Value = data[0]

	return nil
}

func (param *PEFConfigParam_NumberOfAlertStrings) Pack() []byte {
	return []byte{param.Value}
}

func (param *PEFConfigParam_NumberOfAlertStrings) Format() string {
	return fmt.Sprintf("Number of Alert Strings: %d\n", *param)
}

type PEFConfigParam_AlertStringKeys struct {
	AlertStringSelector uint8
	EventFilterNumber   uint8
	AlertStringSet      uint8
}

func (*PEFConfigParam_AlertStringKeys) PEFConfigParamSelector() PEFConfigParamSelector {
	return PEFConfigParamSelector_AlertStringKeys
}

func (param *PEFConfigParam_AlertStringKeys) Unpack(data []byte) error {
	if len(data) < 3 {
		return ErrUnpackedDataTooShortWith(len(data), 3)
	}

	param.AlertStringSelector = data[0]
	param.EventFilterNumber = data[1]
	param.AlertStringSet = data[2]

	return nil
}

func (param *PEFConfigParam_AlertStringKeys) Pack() []byte {
	return []byte{param.AlertStringSelector, param.EventFilterNumber, param.AlertStringSet}
}

func (param *PEFConfigParam_AlertStringKeys) Format() string {
	return fmt.Sprintf(`
	AlertStringSelector    : %d
	EventFilterNumber      : %d
	AlertStringSet         : %d
`, param.AlertStringSelector, param.EventFilterNumber, param.AlertStringSet)
}

type PEFConfigParam_AlertStrings struct {
	// Set Selector = string selector.
	// [7] - reserved.
	// [6:0] - string selector.
	// 0 = selects volatile string
	// 01h-7Fh = non-volatile string selectors
	StringSelector uint8

	// Block Selector = string block number to set, 1 based. Blocks are 16 bytes.
	BlockSelector uint8

	StringData []byte
}

func (*PEFConfigParam_AlertStrings) PEFConfigParamSelector() PEFConfigParamSelector {
	return PEFConfigParamSelector_AlertStrings
}

func (param *PEFConfigParam_AlertStrings) Unpack(data []byte) error {
	if len(data) < 3 {
		return ErrUnpackedDataTooShortWith(len(data), 3)
	}

	param.StringSelector = data[0]
	param.BlockSelector = data[1]
	param.StringData, _, _ = unpackBytes(data, 2, len(data)-2)

	return nil
}

func (param *PEFConfigParam_AlertStrings) Pack() []byte {
	out := make([]byte, 2+len(param.StringData))

	out[0] = param.StringSelector
	out[1] = param.BlockSelector
	packBytes(param.StringData, out, 2)

	return out
}

func (param *PEFConfigParam_AlertStrings) Format() string {
	return fmt.Sprintf(`AlertStringSelector:   %d
	BlockSelector: %d
	StringData:    %s
`, param.StringSelector, param.BlockSelector, string(param.StringData))
}

type PEFConfigParam_NumberOfGroupControlTableEntries struct {
	Value uint8
}

func (*PEFConfigParam_NumberOfGroupControlTableEntries) PEFConfigParamSelector() PEFConfigParamSelector {
	return PEFConfigParamSelector_NumberOfGroupControlTableEntries
}

func (param *PEFConfigParam_NumberOfGroupControlTableEntries) Unpack(data []byte) error {
	if len(data) < 1 {
		return ErrUnpackedDataTooShortWith(len(data), 1)
	}

	param.Value = data[0]

	return nil
}

func (param *PEFConfigParam_NumberOfGroupControlTableEntries) Pack() []byte {
	return []byte{param.Value}
}

func (param *PEFConfigParam_NumberOfGroupControlTableEntries) Format() string {
	return fmt.Sprintf("Number of Group Control Table Entries: %d\n", *param)
}

type PEFConfigParam_GroupControlTable struct {
	EntrySelector uint8

	ForceControlOperation bool
	DelayedControl        bool
	ChannelNumber         uint8

	GroupID0              uint8
	MemberID0             uint8
	DisableMemberID0Check bool

	GroupID1              uint8
	MemberID1             uint8
	DisableMemberID1Check bool

	GroupID2              uint8
	MemberID2             uint8
	DisableMemberID2Check bool

	GroupID3              uint8
	MemberID3             uint8
	DisableMemberID3Check bool

	RetryCount uint8

	Operation uint8
}

func (*PEFConfigParam_GroupControlTable) PEFConfigParamSelector() PEFConfigParamSelector {
	return PEFConfigParamSelector_GroupControlTable
}

func (param *PEFConfigParam_GroupControlTable) Unpack(data []byte) error {
	if len(data) < 11 {
		return ErrUnpackedDataTooShortWith(len(data), 11)
	}

	param.EntrySelector = data[0]

	param.ForceControlOperation = isBit5Set(data[1])
	param.DelayedControl = isBit4Set(data[1])
	param.ChannelNumber = data[1] & 0x0F

	param.GroupID0 = data[2]
	param.MemberID0 = data[3] & 0x0F
	param.DisableMemberID0Check = isBit4Set(data[3])

	param.GroupID1 = data[4]
	param.MemberID1 = data[5] & 0x0F
	param.DisableMemberID1Check = isBit4Set(data[5])

	param.GroupID2 = data[6]
	param.MemberID2 = data[7] & 0x0F
	param.DisableMemberID2Check = isBit4Set(data[7])

	param.GroupID3 = data[8]
	param.MemberID3 = data[9] & 0x0F
	param.DisableMemberID3Check = isBit4Set(data[9])

	// data 11: - Retries and Operation
	// [7] - reserved
	// [6:4] - number of times to retry sending the command to perform
	// the group operation [For ICMB, the BMC broadcasts a
	// Group Chassis Control command] (1-based)
	param.RetryCount = (data[10] & 0x7F) >> 4
	param.Operation = data[10] & 0x0F
	return nil
}

func (param *PEFConfigParam_GroupControlTable) Pack() []byte {
	var b uint8

	out := make([]byte, 11)
	out[0] = param.EntrySelector

	b = param.ChannelNumber & 0x0F
	b = setOrClearBit5(b, param.ForceControlOperation)
	b = setOrClearBit4(b, param.DelayedControl)
	out[1] = b

	out[2] = param.GroupID0
	b = param.MemberID0 & 0x0F
	b = setOrClearBit4(b, param.DisableMemberID0Check)
	out[3] = b

	out[4] = param.GroupID1
	b = param.MemberID1 & 0x0F
	b = setOrClearBit4(b, param.DisableMemberID1Check)
	out[5] = b

	out[6] = param.GroupID2
	b = param.MemberID2 & 0x0F
	b = setOrClearBit4(b, param.DisableMemberID2Check)
	out[7] = b

	out[8] = param.GroupID3
	b = param.MemberID3 & 0x0F
	b = setOrClearBit4(b, param.DisableMemberID3Check)
	out[9] = b

	b = param.RetryCount << 4
	b |= param.Operation
	out[10] = b

	return out
}

func (param *PEFConfigParam_GroupControlTable) Format() string {
	return fmt.Sprintf(`EntrySelector:           %d
	ForceControlOperation:  %v
	DelayedControl:         %v
	ChannelNumber:          %d
	GroupID0:               %d
	MemberID0:              %d
	DisableMemberID0Check:  %v
	GroupID1:               %d
	MemberID1:              %d
	DisableMemberID1Check:  %v
	GroupID2:               %d
	MemberID2:              %d
	DisableMemberID2Check:  %v
	GroupID3:               %d
	MemberID3:              %d
	DisableMemberID3Check:  %v
	RetryCount:             %d
	Operation:              %d`,
		param.EntrySelector,
		param.ForceControlOperation,
		param.DelayedControl,
		param.ChannelNumber,
		param.GroupID0,
		param.MemberID0,
		param.DisableMemberID0Check,
		param.GroupID1,
		param.MemberID1,
		param.DisableMemberID1Check,
		param.GroupID2,
		param.MemberID2,
		param.DisableMemberID2Check,
		param.GroupID3,
		param.MemberID3,
		param.DisableMemberID3Check,
		param.RetryCount,
		param.Operation)
}
