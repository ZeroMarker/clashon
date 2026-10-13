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
	f, err := os.Open("/proc/self/status")
	if err != nil {
		return err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if v, ok := strings.CutPrefix(sc.Text(), "CapEff:"); ok {
			caps, err := strconv.ParseUint(strings.TrimSpace(v), 16, 64)
			if err == nil && caps&(1<<capNetAdmin) != 0 {
				return nil
			}
		}
	}
	return ErrMissing
}

// grant sets file capabilities so non-root runs can manage TUN and bind :53.
// Capabilities are lost whenever the binary is replaced, e.g. on update.
func grant(exe string) error {
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
