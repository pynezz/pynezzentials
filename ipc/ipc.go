/*
  IPC Package provides utility functions for Unix Domain Socket IPC communication.
*/

package ipc

import (
	"encoding/gob"
	"os"
	"path"
	"time"

	"github.com/pynezz/pynezzentials/ansi"
)

const (
	AF_UNIX  = "unix"     // UNIX domain sockets
	AF_DGRAM = "unixgram" // UNIX domain datagram sockets as specified in net package

	STREAM = "SOCK_STREAM" // Stream socket 		(like TCP)
	DGRAM  = "SOCK_DGRAM"  // Datagram socket 		(like UDP)

	// Network values if applicable
	Network = "tcp"
	Address = "localhost:50052"
	Timeout = 1 * time.Second
)

var IPCID []byte // Identifier of the IPC communication

type IPCMessageId []byte // Identifier of the message

func SetIPCID(id []byte) {
	if IPCID == nil {
		IPCID = id
		ansi.PrintSuccess("Set IPC ID to " + string(IPCID))
	} else {
		ansi.PrintWarning("IPC ID already set to " + string(IPCID))
	}
}

func GetIPCStrID() string {
	return string(IPCID)
}

func DefaultSock(name string) string {
	tmpDir := os.TempDir()                     // Temporary directory (eg. /tmp)
	subTmpDir := path.Join(tmpDir, name)       // Subdirectory in the temporary directory (eg. /tmp/<subTmpDir>)
	sock := path.Join(subTmpDir, name+".sock") // Socket file path (eg. /tmp/<subTmpDir>/<name>)
	sock = path.Clean(sock)                    // Clean the path

	return sock
}

func init() {

	gob.Register(IPCRequest{})
	gob.Register(IPCMessage{})
	gob.Register(IPCHeader{})
	gob.Register(IPCMessageId{})
	gob.Register(IPCResponse{})
}
