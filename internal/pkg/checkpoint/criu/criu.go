package criu

import (
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/apptainer/apptainer/internal/pkg/buildcfg"
	"github.com/apptainer/apptainer/internal/pkg/checkpoint"
	"github.com/apptainer/apptainer/internal/pkg/util/paths"
	"github.com/apptainer/apptainer/pkg/sylog"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Bins []string `yaml:"bins"`
	Libs []string `yaml:"libs"`
}

const (
	ContainerStatePath = "/.checkpoint"
	PidFile = "cmd.pid"
	BackSuffix = ".back"
	// portFile           = "coord.port"
	// logFile            = "coord.log"
)

const (
	criuPath = "criu"
)

func criuDir() string {
	return filepath.Join(checkpoint.StatePath(), criuPath)
}

func parseConfig() (*Config, error) {
	confPath := filepath.Join(buildcfg.APPTAINER_CONFDIR, "criu-conf.yaml")
	buf, err := ioutil.ReadFile(confPath)
	if err != nil {
		return nil, err
	}

	var c Config
	err = yaml.Unmarshal(buf, &c)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

func GetPaths() ([]string, []string, error) {
	conf, err := parseConfig()
	if err != nil {
		return nil, nil, err
	}

	libs, bins, err := paths.Resolve(append(conf.Bins, conf.Libs...))
	if err != nil {
		return nil, nil, err
	}

	var usrBins []string
	for _, bin := range bins {
		usrBin := filepath.Join("/usr/bin", filepath.Base(bin))
		usrBins = append(usrBins, strings.Join([]string{bin, usrBin}, ":"))
	}

	return usrBins, libs, nil
}

// QuickInstallationCheck is a quick smoke test to see if criu is installed
// on the host for injection by checking for one of the well known criu
// executables in the PATH. If not found a warning is emitted.
func QuickInstallationCheck() {
	_, err := exec.LookPath("criu")
	if err == nil {
		return
	}

	sylog.Warningf("Unable to locate a criu installation, some functionality may not work as expected. Please ensure a criu installation exists or install it following instructions here: https://github.com/checkpoint-restore/criu/blob/criu-dev/INSTALL.md")
}
