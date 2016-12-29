package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/docker/pkg/version"
	"github.com/docker/docker/runconfig"
)

func (s *Server) getExecByID(version version.Version, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if vars == nil {
		return fmt.Errorf("Missing parameter 'id'")
	}

	eConfig, err := s.daemon.ContainerExecInspect(vars["id"])
	if err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, eConfig)
}

func (s *Server) postContainerExecCreate(version version.Version, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if err := parseForm(r); err != nil {
		return err
	}
	if err := checkForJSON(r); err != nil {
		return err
	}
	name := vars["name"]

	execConfig := &runconfig.ExecConfig{}
	if err := json.NewDecoder(r.Body).Decode(execConfig); err != nil {
		return err
	}
	execConfig.Container = name

	if len(execConfig.Cmd) == 0 {
		return fmt.Errorf("No exec command specified")
	}

	// Register an instance of Exec in container.
	id, err := s.daemon.ContainerExecCreate(execConfig)
	if err != nil {
		logrus.Errorf("Error setting up exec command in container %s: %s", name, err)
		return err
	}

	return writeJSON(w, http.StatusCreated, &types.ContainerExecCreateResponse{
		ID: id,
	})
}

// TODO(vishh): Refactor the code to avoid having to specify stream config as part of both create and start.
func (s *Server) postContainerExecStart(version version.Version, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if err := parseForm(r); err != nil {
		return err
	}
	var (
		execName = vars["name"]
		stdin    io.ReadCloser
		stdout   io.Writer
		stderr   io.Writer
	)

	execStartCheck := &types.ExecStartCheck{}
	if err := json.NewDecoder(r.Body).Decode(execStartCheck); err != nil {
		return err
	}

	if !execStartCheck.Detach {
		// Setting up the streaming http interface.
		inStream, outStream, err := hijackServer(w)
		if err != nil {
			return err
		}
		defer closeStreams(inStream, outStream)

		var errStream io.Writer

		if _, ok := r.Header["Upgrade"]; ok {
			fmt.Fprintf(outStream, "HTTP/1.1 101 UPGRADED\r\nContent-Type: application/vnd.docker.raw-stream\r\nConnection: Upgrade\r\nUpgrade: tcp\r\n\r\n")
		} else {
			fmt.Fprintf(outStream, "HTTP/1.1 200 OK\r\nContent-Type: application/vnd.docker.raw-stream\r\n\r\n")
		}

		if !execStartCheck.Tty {
			errStream = stdcopy.NewStdWriter(outStream, stdcopy.Stderr)
			outStream = stdcopy.NewStdWriter(outStream, stdcopy.Stdout)
		}

		stdin = inStream
		stdout = outStream
		stderr = errStream
	}
	// Now run the user process in container.

	if err := s.daemon.ContainerExecStart(execName, stdin, stdout, stderr); err != nil {
		logrus.Errorf("Error starting exec command in container %s: %s", execName, err)
		return err
	}
	w.WriteHeader(http.StatusNoContent)

	return nil
}

func (s *Server) postContainerExecResize(version version.Version, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if err := parseForm(r); err != nil {
		return err
	}
	if vars == nil {
		return fmt.Errorf("Missing parameter")
	}

	height, err := strconv.Atoi(r.Form.Get("h"))
	if err != nil {
		return err
	}
	width, err := strconv.Atoi(r.Form.Get("w"))
	if err != nil {
		return err
	}

	return s.daemon.ContainerExecResize(vars["name"], height, width)
}
