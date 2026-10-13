package privilege

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

const capNetAdmin = 12

func check() error {
	if os.Geteuid() == 0 {
		return nil
	}
	if caps, err := capSet("CapEff"); err == nil && caps&(1<<capNetAdmin) != 0 {
		return nil
	}
	return ErrMissing
}

// capSet reads one capability mask (CapEff, CapBnd, ...) of this process.
func capSet(name string) (uint64, error) {
	f, err := os.Open("/proc/self/status")
	if err != nil {
		return 0, err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if v, ok := strings.CutPrefix(sc.Text(), name+":"); ok {
			return strconv.ParseUint(strings.TrimSpace(v), 16, 64)
		}
	}
	return 0, fmt.Errorf("%s not found", name)
}

// grant sets file capabilities so non-root runs can manage TUN and bind :53.
// Capabilities are lost whenever the binary is replaced, e.g. on update.
func grant(exe string) error {
	// A file capability outside the bounding set makes the binary fail to
	// execute at all (EPERM), e.g. inside most containers.
	if bnd, err := capSet("CapBnd"); err != nil || bnd&(1<<capNetAdmin) == 0 {
		return fmt.Errorf("CAP_NET_ADMIN is not available in this environment (container?); TUN mode cannot be used here")
	}
	return setcap("cap_net_admin,cap_net_bind_service=+ep", exe)
}

func revoke(exe string) error { return setcap("-r", exe) }

func setcap(arg, exe string) error {
	if os.Geteuid() != 0 {
		return fmt.Errorf("must run as root (sudo)")
	}
	out, err := exec.Command("setcap", arg, exe).CombinedOutput()
	if err != nil {
		return fmt.Errorf("setcap: %v: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}
