// +build experimental

package client

import (
	"fmt"

	Cli "github.com/docker/docker/cli"
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/docker/docker/runconfig"
	 "github.com/opencontainers/runc/libcontainer"
)


func (cli *DockerCli) CmdCheckpoint(args ...string) error {
	cmd := Cli.Subcmd("checkpoint", []string{"CONTAINER [CONTAINER...]"}, "Checkpoint one or more running containers", true)
	cmd.Require(flag.Min, 1)

	var (
		flImgDir        = cmd.String([]string{"-image-dir"}, "", "directory for storing checkpoint image files")
		flWorkDir       = cmd.String([]string{"-work-dir"}, "", "directory for storing log file")
		flLeaveRunning  = cmd.Bool([]string{"-leave-running"}, false, "leave the container running after checkpoint")
		flCheckTcp      = cmd.Bool([]string{"-allow-tcp"}, false, "allow checkpointing tcp connections")
		flExtUnix       = cmd.Bool([]string{"-allow-ext-unix"}, false, "allow checkpointing external unix connections")
		flShell         = cmd.Bool([]string{"-allow-shell"}, false, "allow checkpointing shell jobs")
		flPrevImagesDir = cmd.String([]string{"-prev-images-dir"},"","directory for storing the pre dump memory files")
		flPreDump       = cmd.Bool([]string{"-pre-dump"}, false, "allow checkpoint by pre dump")
		flTrackMem      = cmd.Bool([]string{"-track-mem"}, false, "allow turn on the memory track in kernel")
		flAutoDedup     = cmd.Bool([]string{"-auto-dedup"}, false, "allow open the image directory and punches hole in parent images")
		flPs            = cmd.Bool([]string{"-page-server"}, false, "allow turn on the page server service,senf pages to page server")
		flAddress       = cmd.String([]string{"-address"}, "", "address of server or service")
		flPort          = cmd.Int([]string{"-port"},0, "port of page server")
	)

	if err := cmd.ParseFlags(args, true); err != nil {
		return err
	}

	if cmd.NArg() < 1 {
		cmd.Usage()
		return nil
	}


	criuOpts := &runconfig.CriuConfig{
		ImagesDirectory:         *flImgDir,
		WorkDirectory:           *flWorkDir,
		LeaveRunning:            *flLeaveRunning,
		TcpEstablished:          *flCheckTcp,
		ExternalUnixConnections: *flExtUnix,
		ShellJob:                *flShell,
		PrevImagesDir:           *flPrevImagesDir,
		PreDump:                 *flPreDump,
		TrackMem:	             *flTrackMem,
		AutoDedup:               *flAutoDedup,
	}
	if *flPs==true && *flAddress !="" && *flPort !=0 {
		fmt.Printf("Page server turn on")
		criuOpts.PageServer = libcontainer.CriuPageServerInfo{
			Address:    *flAddress,
			Port:       int32(*flPort),
		}
	}

	var encounteredError error
	for _, name := range cmd.Args() {
		_, _, err := readBody(cli.call("POST", "/containers/"+name+"/checkpoint", criuOpts, nil))
		if err != nil {
			fmt.Fprintf(cli.err, "%s\n", err)
			encounteredError = fmt.Errorf("Error: failed to checkpoint one or more containers")
		} else {
			fmt.Fprintf(cli.out, "%s\n", name)
		}
	}
	return encounteredError
}
