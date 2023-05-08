package registry

type Registration struct {
	ServerName            ServerName
	ServerURL             string
	ServerRequireServices []ServerName
	ServerUpdateURL       string
	HeartbeatUrl          string
}

type ServerName string

const (
	LogService    = ServerName("Log service")
	GradeService  = ServerName("Grade service")
	PortalService = ServerName("Portal")
)

type patchEntry struct {
	Name ServerName
	URL  string
}

type patch struct {
	Added   []patchEntry
	Removed []patchEntry
}
