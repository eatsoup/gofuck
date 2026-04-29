package rules

import (
	"reflect"
	"testing"

	"github.com/eatsoup/gofuck/internal/specific"
)

// withInterfaces swaps specific.EnumerateInterfaces for one block, mirroring
// upstream's BytesIO mock of `ifconfig -a` stdout.
func withInterfaces(t *testing.T, names []string, fn func()) {
	t.Helper()
	prev := specific.EnumerateInterfaces
	t.Cleanup(func() { specific.EnumerateInterfaces = prev })
	specific.EnumerateInterfaces = func() []string { return names }
	fn()
}

func TestIfconfigDeviceNotFoundMatch(t *testing.T) {
	output := "wlan0: error fetching interface information: Device not found"
	cases := []struct {
		script, output string
	}{
		{"ifconfig wlan0", output},
		{"ifconfig -s eth0", "eth0: error fetching interface information: Device not found"},
	}
	for _, tc := range cases {
		withInterfaces(t, []string{"wlp2s0"}, func() {
			assertMatch(t, "ifconfig_device_not_found", cmd(tc.script, tc.output), true)
		})
	}
}

func TestIfconfigDeviceNotFoundNotMatch(t *testing.T) {
	cases := []struct {
		script, output string
	}{
		{"config wlan0", "wlan0: error fetching interface information: Device not found"},
		{"ifconfig eth0", ""},
	}
	for _, tc := range cases {
		withInterfaces(t, []string{"wlp2s0"}, func() {
			assertMatch(t, "ifconfig_device_not_found", cmd(tc.script, tc.output), false)
		})
	}
}

func TestIfconfigDeviceNotFoundNewCommand(t *testing.T) {
	output := "wlan0: error fetching interface information: Device not found"
	cases := []struct {
		script string
		want   []string
	}{
		{"ifconfig wlan0", []string{"ifconfig wlp2s0"}},
		{"ifconfig -s wlan0", []string{"ifconfig -s wlp2s0"}},
	}
	for _, tc := range cases {
		withInterfaces(t, []string{"wlp2s0"}, func() {
			got := mustRule(t, "ifconfig_device_not_found").GetNewCommand(cmd(tc.script, output))
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("ifconfig_device_not_found: GetNewCommand(%q) = %v, want %v", tc.script, got, tc.want)
			}
		})
	}
}
