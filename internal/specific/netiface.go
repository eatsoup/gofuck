package specific

import "net"

// EnumerateInterfaces is the seam tests swap. The default implementation calls
// net.Interfaces and returns each interface's Name. Upstream thefuck shells
// out to `ifconfig -a` and scrapes section headers; Go can answer the same
// question without a subprocess.
var EnumerateInterfaces = defaultEnumerateInterfaces

func defaultEnumerateInterfaces() []string {
	ifs, err := net.Interfaces()
	if err != nil {
		return nil
	}
	out := make([]string, 0, len(ifs))
	for _, i := range ifs {
		out = append(out, i.Name)
	}
	return out
}
