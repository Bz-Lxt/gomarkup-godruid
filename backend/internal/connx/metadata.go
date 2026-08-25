package connx

// Metadata is safe to expose on the control plane. It must never contain
// passwords, tokens, or a full DSN.
type Metadata struct {
	Kind     string `json:"kind"`
	Remote   string `json:"remote,omitempty"`
	Local    string `json:"local,omitempty"`
	Note     string `json:"note,omitempty"`
	Isolated bool   `json:"isolated"`
}

func SanitizeRemote(hostport string) string {
	if hostport == "" {
		return ""
	}
	return hostport
}
