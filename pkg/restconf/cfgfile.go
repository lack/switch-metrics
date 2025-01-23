package restconf

import (
	"fmt"
	"os"
	"strings"

	"github.com/adrg/xdg"
	yaml "sigs.k8s.io/yaml/goyaml.v3"
)

var DefaultFilename = "switches.yaml"

var SearchPaths = []string{
	xdg.ConfigHome,
	".",
}

type Cfgfile struct {
	Switches []struct {
		Ip       string `yaml:"ip"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"switches"`
}

func LoadSwitches() ([]Switch, error) {
	var cfgpath string
	for _, p := range SearchPaths {
		cfgpath = strings.Join([]string{p, DefaultFilename}, string(os.PathSeparator))
		_, err := os.Stat(cfgpath)
		if err == nil {
			break
		}
		cfgpath = ""
	}

	if cfgpath == "" {
		return nil, fmt.Errorf("could not find %s in any of: %v", DefaultFilename, SearchPaths)
	}

	data, err := os.ReadFile(cfgpath)
	if err != nil {
		return nil, err
	}

	var cfgfile Cfgfile
	err = yaml.Unmarshal(data, &cfgfile)
	if err != nil {
		return nil, err
	}

	result := make([]Switch, len(cfgfile.Switches))
	for i, sw := range cfgfile.Switches {
		result[i] = Switch{
			Ip:       sw.Ip,
			Username: sw.Username,
			Password: sw.Password,
		}
	}
	return result, nil
}
