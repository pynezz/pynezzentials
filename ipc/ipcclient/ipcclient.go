package ipcclient

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"net"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/pynezz/pynezzentials"
	"github.com/pynezz/pynezzentials/fsutil"
	"github.com/pynezz/pynezzentials/ipc"
)

type IPCClient struct {
	Name string // Name of the module
	Desc string // Description of the module

	Identifier [4]byte // Identifier of the module

	Sock string   // Path to the UNIX domain socket
	conn net.Conn // Connection to the IPC server (UNIX domain socket)
}

func init() {
	gob.Register(&ipc.IPCRequest{})
	gob.Register(&ipc.IPCMessage{})
	gob.Register(&ipc.IPCHeader{})
	gob.Register(&ipc.IPCMessageId{})
	gob.Register(&ipc.IPCResponse{})
}

// NewIPCClient creates a new IPC client and returns it.
// The name is the name of the module, and the socketPath is the path to the UNIX domain socket.
func NewIPCClient(name string, serverId string) *IPCClient {
	ipc.SetIPCID([]byte(serverId)) // What server to communicate with. Used for requests and responses

	c := &IPCClient{
		Name: name,
	}
	tmpDir := os.TempDir() // Get the temporary directory, OS agnostic
	ipcTmpDir := path.Join(tmpDir, serverId)
	c.SetSocket(path.Join(ipcTmpDir, serverId+".sock"))
	// c.SetSocket(serverId) // Default the socket path to the server id name (e.g. "connector" -> "/tmp/connector/connector.sock")
	return c
}

// Connect to the IPC server (UNIX domain socket)
// Parameter is the name of the module
// And identifier is the identifier of the server
func (c *IPCClient) Connect(module string, identifier string) error {
	c.SetDescf("IPC client for %s", module)
	c.Name = module

	// Check if the socket exists
	if c.Sock == "" {
		err := c.SetSocket(defaultSocketPath())
		if err != nil { // The socket did not exist and the user did not want to retry
			return err // Return the error
		}
	}
	conn, err := net.Dial("unix", c.Sock)
	if err != nil {
		fmt.Println("Dial error:", err)
		return err
	}
	c.conn = conn
	c.Identifier = ipc.IDENTIFIERS[identifier]

	pynezzentials.PrintColorAndBg(pynezzentials.BgGray, pynezzentials.BgCyan, "Connected to "+c.Sock)

	// Print box with client info
	pynezzentials.PrintColor(pynezzentials.Cyan, c.Stringify())

	return nil
}

// Set description with format string for easier type conversion
func (c *IPCClient) SetDescf(desc string, args ...interface{}) {
	c.Desc = fmt.Sprintf(desc, args...)
}

func (c *IPCClient) Stringify() string {
	if c.Name == "" {
		pynezzentials.PrintWarning("No name set for IPCClient")
		c.Name = "IPCClient"
	}
	if c.Desc == "" {
		pynezzentials.PrintWarning("No description set for IPCClient")
		c.Desc = "IPC testing client"
	}
	if c.Identifier == [4]byte{} {
		pynezzentials.PrintWarning("No identifier set for IPCClient")
		c.Identifier = ipc.IDENTIFIERS["test_client"]
	}

	stringified := fmt.Sprintln("IPCCLIENT")
	stringified += fmt.Sprintln("-----------")
	stringified += fmt.Sprintf("Name:        %s\n", c.Name)
	stringified += fmt.Sprintf("Description: %s\n", c.Desc)
	stringified += fmt.Sprintf("Identifier:  %s\n", c.Identifier)

	return pynezzentials.FormatRoundedBox(stringified)
}

// returns a bool (retry) and an error
func existHandler(exist bool) (bool, error) {
	if !exist {
		pynezzentials.PrintError("socket (" + defaultSocketPath() + ") not found")
		pynezzentials.PrintColorUnderline(pynezzentials.DarkYellow, "Retry? [Y/n]")
		var response string
		fmt.Scanln(&response)
		if response[0] == 'n' {
			return false, fmt.Errorf("socket not found")
		}
		return true, nil
	}
	return false, nil
}

func (c *IPCClient) SetSocket(socketPath string) error {
	if socketPath == "" {
		socketPath = defaultSocketPath()
	}
	c.Sock = socketPath

	retry, err := existHandler(socketExists(socketPath))
	if err != nil {
		return err
	}
	if retry {
		c.SetSocket(socketPath)
	}
	return err
}

// Get the default socket path (UNIX domain socket path, /tmp/ipc/ipc.sock)
func defaultSocketPath() string {
	tmpDir := os.TempDir() // Get the temporary directory, OS agnostic
	ipcTmpDir := path.Join(tmpDir, "ipc")
	return path.Join(ipcTmpDir, "ipc.sock")
}

// userRetry asks the user if they want to retry connecting to the IPC server.
func userRetry() bool {
	fmt.Println("IPCClient not connected\nDid you forget to call Connect()?")
	pynezzentials.PrintWarning("Do you want to retry? [Y/n]")

	var retry string
	fmt.Scanln(&retry)
	return retry[0] != 'n' // If the user doesn't want to retry, return false
}

func (c *IPCClient) AwaitResponse() error {
	var err error

	if c.conn == nil {
		pynezzentials.PrintError("Connection not established")
	}

	req, err := parseConnection(c.conn)
	if err != nil {
		if err.Error() == "EOF" {
			pynezzentials.PrintWarning("Client disconnected")
			return err
		}
		pynezzentials.PrintError("Error parsing the connection")
		return err
	}
	pynezzentials.PrintSuccess("Received response from server: " + req.Message.StringData)

	if string(req.Message.Data) == "OK" {
		pynezzentials.PrintColorf(pynezzentials.LightCyan, "Message type: %v\n", req.Header.MessageType)
		pynezzentials.PrintSuccess("Checksums match")
	} else {
		pynezzentials.PrintError("Checksums do not match")
		return fmt.Errorf("checksums do not match")
	}

	return nil
}

// SendIPCMessage sends an IPC message to the server.
func (c *IPCClient) SendIPCMessage(msg *ipc.IPCRequest) error {
	var bBuffer bytes.Buffer
	encoder := gob.NewEncoder(&bBuffer)
	err := encoder.Encode(msg)
	if err != nil {
		return err
	}

	if c.conn == nil {
		if !userRetry() {
			return fmt.Errorf("connection not established")
		} else {
			c.Connect(c.Name, strings.Split(path.Base(c.Sock), ".")[0]) // Get the name of the IPC identifier from the socket path
		}
	}

	pynezzentials.PrintItalic("Sending encoded message to server...")
	_, err = c.conn.Write(bBuffer.Bytes())
	if err != nil {
		fmt.Println("Write error:", err)
		return err
	}
	pynezzentials.PrintSuccess("Message sent: " + msg.Message.StringData)

	pynezzentials.PrintDebug("Awaiting response...")
	err = c.AwaitResponse()
	if err != nil {
		pynezzentials.PrintError("Error receiving response from server")
		fmt.Println(err)
		return err
	}

	return nil
}

// NewMessage creates a new IPC message.
func (c *IPCClient) CreateReq(message string, t ipc.MsgType, dataType ipc.DataType) *ipc.IPCRequest {
	checksum := crc32.ChecksumIEEE([]byte(message))
	pynezzentials.PrintDebug("Created IPC checksum: " + strconv.Itoa(int(checksum)))

	return &ipc.IPCRequest{
		MessageSignature: ipc.IPCID,
		Header: ipc.IPCHeader{
			Identifier:  c.Identifier,
			MessageType: byte(t),
		},
		Message: ipc.IPCMessage{
			Datatype:   dataType,
			Data:       []byte(message),
			StringData: message,
		},
		Timestamp:  pynezzentials.UnixNanoTimestamp(),
		Checksum32: int(checksum),
	}
}

func (c *IPCClient) CreateGenericReq(message interface{}, t ipc.MsgType, dataType ipc.DataType) *ipc.IPCRequest {
	pynezzentials.PrintDebug("[CLIENT] Creating a generic IPC request...")
	var data []byte
	var err error

	switch dataType {
	case ipc.DATA_TEXT:
		data = []byte(message.(string))
	case ipc.DATA_INT:
		data = []byte(strconv.Itoa(message.(int)))
	case ipc.DATA_JSON:
		data, err = json.Marshal(message)
		if err != nil {
			// Handle the error
			pynezzentials.PrintError("[CLIENT] Error marshaling JSON data:" + err.Error())
			return nil
		}
		pynezzentials.PrintDebug("[CLIENT] Marshaling JSON data...")

	case ipc.DATA_YAML:
		fmt.Println("[CLIENT] Marshaling YAML data...")
		data, err = yaml.Marshal(message)
		if err != nil {
			pynezzentials.PrintError("[CLIENT] Error marshaling YAML data:" + err.Error())
			return nil
		}
	case ipc.DATA_BIN:
		data = message.([]byte)
	}

	checksum := crc32.ChecksumIEEE(data)
	pynezzentials.PrintDebug("[CLIENT] Created IPC checksum: " + strconv.Itoa(int(checksum)))

	return &ipc.IPCRequest{
		MessageSignature: ipc.IPCID,
		Header: ipc.IPCHeader{
			Identifier:  c.Identifier,
			MessageType: byte(t),
		},
		Message: ipc.IPCMessage{
			Datatype:   dataType,
			Data:       data,
			StringData: fmt.Sprintf("%v", message),
		},
		Timestamp:  pynezzentials.UnixNanoTimestamp(),
		Checksum32: int(checksum),
	}
}

// Return the parsed IPCRequest object
func parseConnection(c net.Conn) (ipc.IPCRequest, error) {
	var request ipc.IPCRequest
	// var reqBuffer bytes.Buffer

	pynezzentials.PrintDebug("[CLIENT] Trying to decode the bytes to a request struct...")
	pynezzentials.PrintColorf(pynezzentials.LightCyan, "[CLIENT] Decoding the bytes to a request struct... %v", c)

	decoder := gob.NewDecoder(c)
	err := decoder.Decode(&request)
	if err != nil {
		if err.Error() == "EOF" {
			pynezzentials.PrintWarning("parseConnection: EOF error, connection closed")
			return request, err
		}
		pynezzentials.PrintWarning("parseConnection: Error decoding the request \n > " + err.Error())
		return request, err
	}

	pynezzentials.PrintDebug("Trying to encode the bytes to a request struct...")
	fmt.Println(request.Stringify())
	pynezzentials.PrintDebug("--------------------")

	pynezzentials.PrintSuccess("[ipcclient.go] Parsed the message signature!")
	fmt.Printf("Message ID: %v\n", request.MessageSignature)

	return request, nil
}

// Close the connection
func (c *IPCClient) Close() {
	c.conn.Close()
}

func countDown(secLeft int) { // i--
	pynezzentials.PrintInfo(pynezzentials.Overwrite + strconv.Itoa(secLeft) + " seconds left" + pynezzentials.Backspace)
	time.Sleep(time.Second)
	if secLeft > 0 {
		countDown(secLeft - 1)
	}
}

func socketExists(socketPath string) bool {
	if !fsutil.FileExists(socketPath) {
		pynezzentials.PrintError("The UNIX domain socket does not exist")
		pynezzentials.PrintInfo("Retrying in 5 seconds...")
		countDown(5)
		return false
	}
	return true
}
