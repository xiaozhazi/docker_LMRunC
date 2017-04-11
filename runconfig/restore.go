package runconfig
import(
 "github.com/opencontainers/runc/libcontainer"
 )

type CriuPageServerInfo struct {
	Address string
	Port    int32
}

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
	AutoDedup               bool
	PageServer              libcontainer.CriuPageServerInfo
}

type RestoreConfig struct {
	CriuOpts     CriuConfig
	ForceRestore bool
}
