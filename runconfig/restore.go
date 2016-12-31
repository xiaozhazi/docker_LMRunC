package runconfig

type CriuConfig struct {
	ImagesDirectory         string
	WorkDirectory           string
	PrevImagesDir           string
	PreDump                 bool
	TrackMem                bool
	LeaveRunning            bool
	TcpEstablished          bool
	ExternalUnixConnections bool
	ShellJob                bool
}

type RestoreConfig struct {
	CriuOpts     CriuConfig
	ForceRestore bool
}
