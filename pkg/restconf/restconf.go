package restconf

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

var DumpRawJson = false

type Switch struct {
	Ip       string
	Username string
	Password string
}

type SyncState int

const (
	SYNC_STATE_LOCKED = iota
	SYNC_STATE_AQUIRING
	SYNC_STATE_HOLDOVER
	SYNC_STATE_UNKNOWN
)

type PtpStatus struct {
	SyncState SyncState
	Offset    int
	Hops      int
	GmId      string
	LocalId   string
}

func (s SyncState) String() string {
	switch s {
	case SYNC_STATE_LOCKED:
		return "locked"
	case SYNC_STATE_AQUIRING:
		return "aquiring"
	case SYNC_STATE_HOLDOVER:
		return "holdover"
	default:
		return "unknown"
	}
}

func (s *Switch) GetPtpStatus() (PtpStatus, error) {
	return PtpStatus{}, nil
}

var tlsConfig = tls.Config{
	InsecureSkipVerify: true,
}

var transport = http.Transport{
	TLSClientConfig: &tlsConfig,
}

var client = http.Client{
	Transport: &transport,
}

func (s *Switch) fetch(path string) ([]byte, error) {
	sep := "/"
	if path[0] == '/' || path[0] == '&' {
		sep = ""
	}
	url := fmt.Sprintf("https://%s/restconf/data%s%s", s.Ip, sep, path)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(s.Username, s.Password)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if DumpRawJson {
		fmt.Fprintf(os.Stderr, "%v\n", string(body))
	}
	return body, err
}

type PtpCfgMode int

const (
	PTP_CFG_UNKNOWN = iota
	PTP_CFG_DISABLED
	PTP_CFG_ENABLED_RX
	PTP_CFG_ENABLED_TX
)

type Iface struct {
	Name   string
	Desc   string
	PtpCfg PtpCfgMode
}

type BasicInterface struct {
	IetfName string `json:"name"`
	IetfDesc string `json:"description"`
}

func (s *Switch) fetchAndUnwrap(path, toplevel string, inner any) error {
	data, err := s.fetch(path)
	if err != nil {
		return err
	}
	var raw map[string]any
	err = json.Unmarshal(data, &raw)
	if err != nil {
		return err
	}
	innerData, err := json.Marshal(raw[toplevel])
	if err != nil {
		return err
	}
	err = json.Unmarshal(innerData, inner)
	return err
}

func (s *Switch) interfaces(interfaces any) error {
	return s.fetchAndUnwrap("ietf-interfaces:interfaces/interface",
		"ietf-interfaces:interface",
		&interfaces)
}

func (s *Switch) Interfaces() ([]Iface, error) {
	var iflist []BasicInterface
	err := s.interfaces(&iflist)
	returnList := make([]Iface, len(iflist))
	for idx, i := range iflist {
		returnList[idx] = Iface{
			Name:   i.IetfName,
			Desc:   i.IetfDesc,
			PtpCfg: PTP_CFG_UNKNOWN,
		}
	}
	return returnList, err
}

type SwitchInfo struct {
	Ip        string
	Hostname  string
	Vendor    string
	Model     string
	SwVersion string
}
