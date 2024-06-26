package ipcserver

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"io"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/pynezz/pynezzentials"
	"github.com/pynezz/pynezzentials/ansi"
	"github.com/pynezz/pynezzentials/fsutil"
	"github.com/pynezz/pynezzentials/ipc"
	"gopkg.in/yaml.v3"
)

/* CONSTANTS
 * The following constants are used for the IPC communication between the connector and the other modules.
 */
const (
	AF_UNIX  = "unix"     // UNIX domain sockets
	AF_DGRAM = "unixgram" // UNIX domain datagram sockets as specified in net package
)

/* IDENTIFIERS
 * To identify the module, the client will send a 4 byte identifier as part of the header.
 */
var MODULEIDENTIFIERS map[string][]byte

var MSGTYPE map[string]byte

var SERVERIDENTIFIER [4]byte

var IPCID []byte

/* TYPES
 * Types for the IPC communication between the connector and the other modules.
 */
type IPCServer struct {
	path       string
	identifier string
	conn       net.Listener
}

func init() {
	MODULEIDENTIFIERS = map[string][]byte{}
}

func LoadModules(path string) {
	if !fsutil.FileExists(path) {
		ansi.PrintError("LoadModules(): File does not exist: " + path)
	}
	ansi.PrintSuccess("File exists: " + path)
	f, err := os.Open(path)
	if err != nil {
		ansi.PrintError("LoadModules(): " + err.Error())
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Println(line)
		parts := strings.Split(line, " ")
		if len(parts[0]) < 1 {
			// empty line
			continue
		}
		firstChar := parts[0][0]
		if firstChar == '#' || firstChar == ' ' || firstChar == '\t' || firstChar == '/' || firstChar == '*' || firstChar == '\n' || firstChar == '\r' {
			// comment
			continue
		}

		for i, part := range parts {
			if part == "\t" || part == " " {
				// skip
				continue
			}
			parts[i] = strings.TrimSpace(part)
		}

		AddModule(parts[0], []byte(parts[1])) // Add module to the server
		ansi.PrintColorf(ansi.LightCyan, "Loaded module: %s", parts[0])
	}
}

// NewIPCServer creates a new IPC server and returns it.
func NewIPCServer(name string, identifier string) *IPCServer {
	path := ipc.DefaultSock(name)
	IPCID = []byte(identifier)
	ipc.SetIPCID(IPCID)
	SetServerIdentifier(IPCID)

	ansi.PrintColorf(ansi.LightCyan, "[🔌SOCKETS] IPC server path: %s", path)

	return &IPCServer{
		path:       path,
		identifier: identifier,
		conn:       nil,
	}
}

// Add a new module identifier to the map
func AddModule(identifier string, id []byte) {
	if len(id) > 4 {
		ansi.PrintError("AddModule(): Identifier length must be 4 bytes")
		ansi.PrintInfo("Truncating the identifier to 4 bytes")
		id = id[:4]
	}
	MODULEIDENTIFIERS[identifier] = id

	ansi.PrintSuccess("added module:" + string(MODULEIDENTIFIERS[identifier]))
}

// Set the server identifier to the SERVERIDENTIFIER variable
func SetServerIdentifier(id []byte) {
	if len(id) > 4 {
		ansi.PrintError("SetServerIdentifier(): Identifier length must be 4 bytes")
		ansi.PrintInfo("Truncating the identifier to 4 bytes")
		id = id[:4]
	}
	SERVERIDENTIFIER = [4]byte(id) // Convert the slice to an array
	ansi.PrintSuccess("Set server identifier: " + string(SERVERIDENTIFIER[:]))
}

// Write a socket file and add it to the map
func (s *IPCServer) InitServerSocket() bool {
	// Making sure the socket is clean before starting
	if err := os.RemoveAll(s.path); err != nil {
		ansi.PrintError("InitServerSocket(): Failed to remove old socket: " + err.Error())
		return false
	}

	return true
}

// Creates a new listener on the socket path (which should be set in the config in the future)
func (s *IPCServer) Listen() {
	ansi.PrintColorBold(ansi.DarkGreen, "🎉 IPC server running!")
	s.conn, _ = net.Listen(AF_UNIX, s.path)
	ansi.PrintColorf(ansi.LightCyan, "[🔌SOCKETS] Starting listener on %s", s.path)

	for {
		ansi.PrintDebug("Waiting for connection...")
		conn, err := s.conn.Accept()
		ansi.PrintColorf(ansi.LightCyan, "[🔌SOCKETS]: New connection from %s", conn.LocalAddr().String())

		if err != nil {
			ansi.PrintError("Listen(): " + err.Error())
			continue
		}

		s.handleConnection(conn)
	}
}

func NewIPCID(identifier string, id []byte) {
	if len(id) > 4 {
		ansi.PrintError("NewIPCID(): Identifier length must be 4 bytes")
		ansi.PrintInfo("Truncating the identifier to 4 bytes")
		id = id[:4]
	}
	ipc.SetIPCID(id)
}

func Cleanup() {
	// for _, server := range connections.sockets {
	// 	ansi.PrintItalic("Cleaning up IPC server: " + server.path)
	// 	err := os.Remove(server.path)
	// 	if err != nil {
	// 		ansi.PrintError("Cleanup(): " + err.Error())
	// 	}

	// 	ansi.PrintItalic("Closing connection: " + server.conn.Addr().String())
	// 	server.conn.Close()
	// }
	ansi.PrintItalic("\t... IPC server cleanup complete.")
}

func crc(b []byte) uint32 {
	return crc32.ChecksumIEEE(b)
}

// Function to create a new IPCMessage based on the identifier key
func NewIPCMessage(identifierKey string, messageType byte, data []byte) (*ipc.IPCRequest, error) {
	identifier, ok := MODULEIDENTIFIERS[identifierKey]
	if !ok {
		AddModule(identifierKey, []byte(identifierKey))
		return nil, fmt.Errorf("invalid identifier key: %s", identifierKey)
	}

	var id [4]byte
	copy(id[:], identifier[:4]) // Ensure no out of bounds panic

	crcsum32 := crc(data)

	message := ipc.IPCMessage{
		Data:       data,
		StringData: string(data),
	}

	return &ipc.IPCRequest{
		Header: ipc.IPCHeader{
			Identifier:  id,
			MessageType: messageType,
		},
		Message:    message,
		Timestamp:  pynezzentials.UnixNanoTimestamp(),
		Checksum32: int(crcsum32),
	}, nil
}

// Return the parsed IPCRequest object
func parseConnection(c net.Conn) (ipc.IPCRequest, error) {
	var request ipc.IPCRequest
	// var reqBuffer bytes.Buffer

	ansi.PrintDebug("Trying to decode the bytes to a request struct...")
	ansi.PrintColorf(ansi.LightCyan, "Decoding the bytes to a request struct... %v", c)

	decoder := gob.NewDecoder(c)
	err := decoder.Decode(&request)
	if err != nil {
		ansi.PrintWarning("parseConnection: Error decoding the request: \n > " + err.Error())
		return request, err
	}
	d := parseData(&request.Message)
	if d == nil {
		fmt.Println("Data is nil")
		return request, nil
	}
	if parseMetadata(d) {
		fmt.Println("Method: ", parseVerb(d))
	}

	fmt.Println(request.Stringify())
	ansi.PrintDebug("--------------------")
	ansi.PrintSuccess("[ipcserver.go] Parsed the message signature!")
	fmt.Printf("Message ID: %s\n", string(request.MessageSignature))

	return request, nil
}

func parseMetadata(msg ipc.GenericData) bool {

	// TODO: Might be reasonable to implement the Metadata struct here (ipc.Metadata)
	metadata := msg["metadata"]
	if metadata == nil {
		return false
	}
	source := metadata.(map[string]interface{})["source"]
	destination := metadata.(map[string]interface{})["destination"].(map[string]interface{})["destination"]
	destinationId := destination.(map[string]interface{})["id"]
	destinationName := destination.(map[string]interface{})["name"]
	destinationInfo := destination.(map[string]interface{})["info"]

	method := metadata.(map[string]interface{})["method"]

	v := ""
	if method == "POST" {
		v = "send data to"
	} else if method == "GET" {
		v = "get data from"
	} else if method == "PUT" {
		v = "update data in"
	} else if method == "DELETE" {
		v = "delete data from"
	} else {
		v = "???"
	}

	sentence := fmt.Sprintf("\n %s wants to %s %s with id %s \n", source, v, destinationName, destinationId)
	ansi.PrintBold(sentence)
	ansi.PrintItalic("Additional info: " + destinationInfo.(string))
	return metadata != nil
}

// Parse the method/verb from the message
func parseVerb(msg ipc.GenericData) string {
	v := msg["metadata"].(map[string]interface{})["method"].(string)
	if v == "" {
		return "nil"
	}
	ansi.PrintSuccess("Verb: " + v)
	return v
}

func parseData(msg *ipc.IPCMessage) ipc.GenericData {
	var data ipc.GenericData

	switch msg.Datatype {
	case ipc.DATA_TEXT:
		fmt.Println("Data is string")
	case ipc.DATA_INT:
		// Parse the integer data
		// Might be a disconnect message
		fmt.Println("Data is integer")

	case ipc.DATA_JSON:
		// Parse the JSON data
		fmt.Println("Data is json / generic data")

		// var temp interface{}
		err := json.Unmarshal(msg.Data, &data)
		if err != nil {
			fmt.Println("Error unmarshaling JSON data:", err)
		} else {
			fmt.Printf("Data: %v\n", data)
		}

		handleGenericData(data)

	case ipc.DATA_YAML:
		// Parse the YAML data
		fmt.Println("Data is YAML / generic data")
		err := yaml.Unmarshal(msg.Data, &data)
		if err != nil {
			fmt.Println("Error unmarshaling YAML data:", err)
		}

		handleGenericData(data)

	case ipc.DATA_BIN:
		// Parse the binary data
		fmt.Println("Data is binary / generic data")

		handleGenericData(data)

	default:
		// Default to generic data
		fmt.Println("Data is generic")
		gob.NewDecoder(bytes.NewReader(msg.Data)).Decode(&data)
		handleGenericData(data)

	}

	if data == nil {
		fmt.Println("Data is nil")
	}

	return data
}

func handleGenericData(data ipc.GenericData) {
	// Handle the data
}

// Calculate the response time
func responseTime(reqTime int64) {
	currTime := pynezzentials.UnixNanoTimestamp()
	diff := currTime - reqTime
	fmt.Printf("Response time: %d\n", diff)
	fmt.Printf("Seconds: %f\n", float64(diff)/1e9)
	fmt.Printf("Milliseconds: %f\n", float64(diff)/1e6)
	fmt.Printf("Microseconds: %f\n", float64(diff)/1e3)
}

// handleConnection handles the incoming connection
func (s *IPCServer) handleConnection(c net.Conn) {
	defer c.Close()

	ansi.PrintColorf(ansi.LightCyan, "[🔌SOCKETS] Handling connection...")

	for {
		request, err := parseConnection(c)
		if err != nil {
			if err == io.EOF {
				ansi.PrintDebug("Connection closed by client")
				break
			}
			ansi.PrintError("Error parsing request: " + err.Error())
			break
		}

		ansi.PrintDebug("Request parsed: " + strconv.Itoa(request.Checksum32))

		// Process the request...
		ansi.PrintColorf(ansi.BgGreen, "Received: %+v\n", request)

		// Finally, respond to the client
		err = s.respond(c, request)
		if err != nil {
			ansi.PrintError("handleConnection: " + err.Error())
			break
		}
	}

}

type ReturnData struct {
	Metadata ipc.Metadata `json:"metadata"`
	Data     interface{}  `json:"data"`
}

type DataSource struct {
}

// TODO: Where should we tell the server where to get the logs from?
func (d *ReturnData) GetLogs(path string, filter string) {
	// Read the logs from the file
	if fsutil.FileExists(path) {

	}

	// Not sure if this fits here. Might rather implement it in the internal package
}

// c is the connection to the client
// req is the request from the client
func (s *IPCServer) respond(c net.Conn, req ipc.IPCRequest) error {
	if req.Checksum32 == 0 {
		ansi.PrintWarning("Checksum is 0, skipping response")
		return nil
	}

	ansi.PrintDebug("Responding to the client...")
	moduleId := string(req.Header.Identifier[:])
	var response *ipc.IPCRequest
	var err error
	if req.Checksum32 == int(crc(req.Message.Data)) {
		// TODO: Refactor. Not very pretty. (the identifier key part)
		response, err = NewIPCMessage(moduleId, ipc.MSG_ACK, []byte("OK"))
	} else {
		fmt.Printf("Request checksum: %v\nCalculated checksum: %v\n", req.Checksum32, crc(req.Message.Data))
		response, err = NewIPCMessage(moduleId, ipc.MSG_ERROR, []byte("CHKSUM ERROR"))
	}
	if err != nil {
		return err
	}

	var responseBuffer bytes.Buffer
	encoder := gob.NewEncoder(&responseBuffer)
	err = encoder.Encode(response)
	if err != nil {
		return err
	}

	_, err = c.Write(responseBuffer.Bytes())
	if err != nil {
		return err
	}
	ansi.PrintColor(ansi.BgGreen, "🚀 Response sent!")

	responseTime(req.Timestamp)

	return nil
}
