package ipcclient

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"net"
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

// NewIPCClient creates a new IPC client and returns it.
// The name is the name of the module, and the socketPath is the path to the UNIX domain socket.
func NewIPCClient(name string, identifier string, serverId string) *IPCClient {
	upper := strings.ToUpper(serverId)
	ipc.SetIPCID([]byte(upper)) // What server to communicate with. Used for requests and responses
	var identifierBytes [4]byte
	copy(identifierBytes[:], identifier)
	c := &IPCClient{
		Name:       name,
		Identifier: identifierBytes, // Set the identifier of the client
	}
	c.SetSocket(ipc.DefaultSock(serverId)) // Lowercase serverId
	return c
}

// Connect to the IPC server (UNIX domain socket)
// Parameter is the name of the module
// And identifier is the identifier of the server
func (c *IPCClient) Connect() error {
	c.SetDescf("IPC client for %s", c.Name)

	fmt.Println("IPC client connecting to", c.Sock)

	conn, err := net.Dial("unix", c.Sock)
	if err != nil {
		fmt.Println("Dial error:", err)
		return err
	}
	c.conn = conn
	// c.Identifier = ipc.IDENTIFIERS[identifier]

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
		// pynezzentials.PrintError("socket (" + defaultSocketPath() + ") not found")
		pynezzentials.PrintError("socket not found!")
		pynezzentials.PrintColorUnderline(pynezzentials.DarkYellow, "Retry? [Y/n]")
		var response string
		fmt.Scanln(&response)
		if len(response) > 0 {
			if response[0] == 'n' {
				return false, fmt.Errorf("socket not found")
			}
		}
		return true, nil
	}
	return false, nil
}

func (c *IPCClient) SetSocket(socketPath string) error {
	if socketPath == "" {
		socketPath = ipc.DefaultSock(ipc.GetIPCStrID())
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

// userRetry asks the user if they want to retry connecting to the IPC server.
func userRetry() bool {
	fmt.Println("IPCClient not connected\nDid you forget to call Connect()?")
	pynezzentials.PrintWarning("Do you want to retry? [Y/n]")

	var retry string
	fmt.Scanln(&retry)
	if len(retry) == 0 { // If the user doesn't enter anything,
		return true // we assume they hit enter and want to retry
	}
	return retry[0] != 'n' // If the user doesn't want to retry, return false
}

func (c *IPCClient) AwaitResponse() (ipc.IPCMessage, error) {
	var err error
	var response ipc.IPCMessage

	if c.conn == nil {
		pynezzentials.PrintError("Connection not established")
	}

	req, err := parseConnection(c.conn)
	if err != nil {
		if err.Error() == "EOF" {
			return response, fmt.Errorf("client disconnected")
		}
		pynezzentials.PrintError("Error parsing the connection")
		return response, err
	}
	// pynezzentials.PrintSuccess("Received response from server: " + req.Message.StringData)

	if string(req.Message.Data) == "OK" {
		pynezzentials.PrintColorf(pynezzentials.LightCyan, "Message type: %v\n", req.Header.MessageType)
		pynezzentials.PrintSuccess("Checksums match")
	} else {
		pynezzentials.PrintError("Checksums do not match")
		return response, fmt.Errorf("checksums do not match")
	}

	return response, nil
}

// ClientListen listens for a message from the server and returns the data.
// GenericData is a generic map for data (map[string]interface{}). It can be used to store any data type.
func (c *IPCClient) ClientListen() ipc.IPCResponse {
	var err error

	response := ipc.IPCResponse{}

	if c.conn == nil {
		pynezzentials.PrintError("Connection not established")
		return response
	}

	res, err := parseConnection(c.conn)
	if err != nil {
		response.Success = false
		if err.Error() == "EOF" {
			pynezzentials.PrintWarning("Client disconnected")
			return response
		}
		pynezzentials.PrintError("Error parsing the connection")
		return response
	}

	response = ipc.IPCResponse{
		Request:    res,
		Success:    true,
		Message:    res.Message.StringData,
		Checksum32: res.Checksum32,
	}

	pynezzentials.PrintSuccess("Received message from server: " + response.Message)

	// TODO: This will not work properly. Time is too constrained atm. Fix this later.
	if string(res.Message.Data) == "OK" {
		pynezzentials.PrintColorf(pynezzentials.LightCyan, "Message type: %v\n", res.Header.MessageType)
		pynezzentials.PrintSuccess("Checksums match")
	} else {
		pynezzentials.PrintError("Checksums do not match")
	}

	return response
}

// SendIPCMessage sends an IPC message to the server.
// To get the response, you can pass a function that will be called after the message is sent.
//
// Example:
//
//	err := client.SendIPCMessage(req, func() (ipc.IPCMessage, error) {
//		return client.ParseResponse()
//	})
func (c *IPCClient) SendIPCMessage(msg *ipc.IPCRequest, then ...func() (ipc.IPCMessage, error)) (ipc.IPCMessage, error) {
	var bBuffer bytes.Buffer
	var response ipc.IPCMessage

	encoder := gob.NewEncoder(&bBuffer)
	err := encoder.Encode(msg)
	if err != nil {
		return response, err
	}

	if c.conn == nil {
		if !userRetry() {
			return response, fmt.Errorf("connection not established")
		} else {
			c.Connect() // Get the name of the IPC identifier from the socket path
		}
	}

	pynezzentials.PrintItalic("Sending encoded message to server...")
	_, err = c.conn.Write(bBuffer.Bytes())
	if err != nil {
		fmt.Println("Write error:", err)
		return response, err
	}
	pynezzentials.PrintSuccess("Message sent: " + msg.Message.StringData)

	pynezzentials.PrintDebug("Awaiting response...")

	// next is a function that will be called after the message is sent
	next := func() (ipc.IPCMessage, error) {
		pynezzentials.PrintDebug("Awaiting response...")
		return c.AwaitResponse()
	}

	// funcArr := []func() (ipc.IPCMessage, error){}

	// // If there are any functions in the then array, add them to the funcArr
	// if len(then) > 0 {
	// 	for _, f := range then[:len(then)-1] {
	// 		funcArr = append(funcArr, f)
	// 	}
	// 	funcArr = append(funcArr, next)
	// } else {
	// 	response, err = next()
	// }
	response, err = next()

	if err != nil {
		pynezzentials.PrintError("Error receiving response from server")
		fmt.Println(err)
	}

	return response, nil
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
