package restconf

import "strconv"

type DellSwitch struct {
	Switch
}

type DellPtpCfg struct {
	Role   string `json:"role"`
	Enable bool   `json:"enable"`
}

type DellInterface struct {
	BasicInterface
	DellPtpCfg *DellPtpCfg `json:"dell-ptp:ptp-port-config,optional"`
}

func (i DellInterface) Name() string {
	return i.IetfName
}

func (i DellInterface) Desc() string {
	return i.IetfDesc
}

func (i DellInterface) PtpCfg() PtpCfgMode {
	if i.DellPtpCfg == nil || !i.DellPtpCfg.Enable {
		return PTP_CFG_DISABLED
	}
	if i.DellPtpCfg.Role == "master" {
		return PTP_CFG_ENABLED_TX
	}
	if i.DellPtpCfg.Role == "slave" {
		return PTP_CFG_ENABLED_RX
	}
	return PTP_CFG_UNKNOWN
}

func (s *DellSwitch) Interfaces() ([]Iface, error) {
	var iflist []DellInterface
	err := s.interfaces(&iflist)
	returnList := make([]Iface, len(iflist))
	for idx, i := range iflist {
		returnList[idx] = Iface{
			Name:   i.IetfName,
			Desc:   i.IetfDesc,
			PtpCfg: i.PtpCfg(),
		}
	}
	return returnList, err
}

type DefaultDs struct {
	ClockId string `json:"clock-identity"`
}

type CurrentDs struct {
	Offset string `json:"offset-from-master"`
	Steps  int    `json:"steps-removed"`
}

type ParentDs struct {
	GrandmasterId string `json:"grandmaster-identity"`
}

type ServoStatus struct {
	State  string `json:"servo-state"`
	Status string `json:"lock-status"`
}

type ClockDs struct {
	Local   DefaultDs `json:"default-ds"`
	Current CurrentDs `json:"current-ds"`
}

type DellPtpState struct {
	ClockDs ClockDs     `json:"clock-ds"`
	Parent  ParentDs    `json:"parent-ds"`
	Servo   ServoStatus `json:"servo-status"`
}

func (s *DellSwitch) GetPtpStatus() (PtpStatus, error) {
	var ptpState DellPtpState
	err := s.fetchAndUnwrap("dell-ptp:ptp-ds", "dell-ptp:ptp-ds", &ptpState)
	if err != nil {
		return PtpStatus{}, err
	}
	var syncState SyncState
	switch ptpState.Servo.State {
	case "locked":
		syncState = SYNC_STATE_LOCKED
	default:
		syncState = SYNC_STATE_UNKNOWN
	}

	offset, err := strconv.Atoi(ptpState.ClockDs.Current.Offset)
	if err != nil {
		return PtpStatus{}, err
	}
	return PtpStatus{
		GmId:      ptpState.Parent.GrandmasterId,
		Offset:    offset,
		Hops:      ptpState.ClockDs.Current.Steps,
		LocalId:   ptpState.ClockDs.Local.ClockId,
		SyncState: syncState,
	}, nil
}

type DellSystem struct {
	Hostname string `json:"hostname"`
}

func (s *DellSwitch) Info() (SwitchInfo, error) {
	var sys DellSystem
	err := s.fetchAndUnwrap("dell-system:system", "dell-system:system", &sys)
	if err != nil {
		return SwitchInfo{}, err
	}
	return SwitchInfo{
		Ip:       s.Ip,
		Hostname: sys.Hostname,
	}, nil
}
