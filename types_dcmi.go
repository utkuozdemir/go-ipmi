package ipmi

// [DCMI specification v1.5]: 6.6.3 Set Power Limit
// Exception Actions, taken if the Power Limit is exceeded and cannot be controlled within the Correction Time Limit
type DCMIExceptionAction uint8

const (
	DCMIExceptionAction_NoAction          DCMIExceptionAction = 0x00
	DCMIExceptionAction_PowerOffAndLogSEL DCMIExceptionAction = 0x01 // Hard Power Off system and log events to SEL
	DCMIExceptionAction_LogSELOnly        DCMIExceptionAction = 0x11
)

func (a DCMIExceptionAction) String() string {
	m := map[DCMIExceptionAction]string{
		0x00: "No Action",
		0x01: "Hard Power Off system and log events to SEL",
		0x11: "Log event to SEL only",
	}
	s, ok := m[a]
	if ok {
		return s
	}
	return "unknown"
}
