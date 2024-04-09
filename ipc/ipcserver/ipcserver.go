package ipcserver

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"io"
	"net"
	"os"
	"strconv"

	"github.com/pynezz/pynezzentials"
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
	path string
	conn net.Listener
}

func init() {
	MODULEIDENTIFIERS = map[string][]byte{}

	gob.Register(ipc.IPCRequest{})
	gob.Register(ipc.IPCMessage{})
	gob.Register(ipc.IPCHeader{})
	gob.Register(ipc.IPCMessageId{})
	gob.Register(&ipc.IPCResponse{})
}

// NewIPCServer creates a new IPC server and returns it.
func NewIPCServer(name string, identifier string) *IPCServer {
	path := ipc.DefaultSock(name)
	IPCID = []byte(identifier)
	ipc.SetIPCID(IPCID)

	pynezzentials.PrintColorf(pynezzentials.LightCyan, "[🔌SOCKETS] IPC server path: %s", path)

	return &IPCServer{
		path: path,
	}
}

// Add a new module identifier to the map
func AddModule(identifier string, id []byte) {
	if len(id) > 4 {
		pynezzentials.PrintError("AddModule(): Identifier length must be 4 bytes")
		pynezzentials.PrintInfo("Truncating the identifier to 4 bytes")
		id = id[:4]
	}
	MODULEIDENTIFIERS[identifier] = id
}

// Set the server identifier to the SERVERIDENTIFIER variable
func SetServerIdentifier(id []byte) {
	if len(id) > 4 {
		pynezzentials.PrintError("SetServerIdentifier(): Identifier length must be 4 bytes")
		pynezzentials.PrintInfo("Truncating the identifier to 4 bytes")
		id = id[:4]
	}
	SERVERIDENTIFIER = [4]byte(id) // Convert the slice to an array
}

// Write a socket file and add it to the map
func (s *IPCServer) InitServerSocket() bool {
	// Making sure the socket is clean before starting
	if err := os.RemoveAll(s.path); err != nil {
		pynezzentials.PrintError("InitServerSocket(): Failed to remove old socket: " + err.Error())
		return false
	}

	return true
}

// Creates a new listener on the socket path (which should be set in the config in the future)
func (s *IPCServer) Listen() {
	pynezzentials.PrintColorBold(pynezzentials.DarkGreen, "🎉 IPC server running!")
	s.conn, _ = net.Listen(AF_UNIX, s.path)
	pynezzentials.PrintColorf(pynezzentials.LightCyan, "[🔌SOCKETS] Starting listener on %s", s.path)

	for {
		pynezzentials.PrintDebug("Waiting for connection...")
		conn, err := s.conn.Accept()
		pynezzentials.PrintColorf(pynezzentials.LightCyan, "[🔌SOCKETS]: New connection from %s", conn.LocalAddr().String())

		if err != nil {
			pynezzentials.PrintError("Listen(): " + err.Error())
			continue
		}

		handleConnection(conn)
	}
}

func NewIPCID(identifier string, id []byte) {
	if len(id) > 4 {
		pynezzentials.PrintError("NewIPCID(): Identifier length must be 4 bytes")
		pynezzentials.PrintInfo("Truncating the identifier to 4 bytes")
		id = id[:4]
	}
	ipc.SetIPCID(id)
}

func Cleanup() {
	// for _, server := range connections.sockets {
	// 	pynezzentials.PrintItalic("Cleaning up IPC server: " + server.path)
	// 	err := os.Remove(server.path)
	// 	if err != nil {
	// 		pynezzentials.PrintError("Cleanup(): " + err.Error())
	// 	}

	// 	pynezzentials.PrintItalic("Closing connection: " + server.conn.Addr().String())
	// 	server.conn.Close()
	// }
	pynezzentials.PrintItalic("\t... IPC server cleanup complete.")
}

func crc(b []byte) uint32 {
	return crc32.ChecksumIEEE(b)
}

// Function to create a new IPCMessage based on the identifier key
func NewIPCMessage(identifierKey string, messageType byte, data []byte) (*ipc.IPCRequest, error) {
	identifier, ok := MODULEIDENTIFIERS[identifierKey]
	if !ok {
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

	pynezzentials.PrintDebug("Trying to decode the bytes to a request struct...")
	pynezzentials.PrintColorf(pynezzentials.LightCyan, "Decoding the bytes to a request struct... %v", c)

	decoder := gob.NewDecoder(c)
	err := decoder.Decode(&request)
	if err != nil {
		pynezzentials.PrintWarning("parseConnection: Error decoding the request: \n > " + err.Error())
		return request, err
	}
	d := parseData(&request.Message)
	fmt.Println("Vendor: ", d["vendor"])

	fmt.Println(request.Stringify())
	pynezzentials.PrintDebug("--------------------")
	pynezzentials.PrintSuccess("[ipcserver.go] Parsed the message signature!")
	fmt.Printf("Message ID: %s\n", string(request.MessageSignature))

	return request, nil
}

func parseData(msg *ipc.IPCMessage) ipc.GenericData {
	var data ipc.GenericData
	// var dataType ipc.DataType

	switch msg.Datatype {
	case ipc.DATA_TEXT:
		// Parse the integer data
		fmt.Println("Data is string")
	case ipc.DATA_INT:
		// Parse the JSON data
		fmt.Println("Data is integer")
		// data = ipc.JSONData(msg.Data)
	case ipc.DATA_JSON:
		// Parse the string data
		fmt.Println("Data is json / generic data")
		// json.Unmarshal(msg.Data, &data)

		// var temp interface{}
		err := json.Unmarshal(msg.Data, &data)
		if err != nil {
			fmt.Println("Error unmarshaling JSON data:", err)
		} else {
			// fmt.Println("Temporary data:", temp)
			// data = temp.(map[string]interface{})
			fmt.Printf("Data: %v\n", data)
		}

	case ipc.DATA_YAML:
		// Parse the YAML data
		fmt.Println("Data is YAML / generic data")
		err := yaml.Unmarshal(msg.Data, &data)
		if err != nil {
			fmt.Println("Error unmarshaling YAML data:", err)
		}
	case ipc.DATA_BIN:
		// Parse the binary data
		fmt.Println("Data is binary / generic data")
	default:
		// Default to generic data
		fmt.Println("Data is generic")
		gob.NewDecoder(bytes.NewReader(msg.Data)).Decode(&data)
	}

	if data == nil {
		fmt.Println("Data is nil")
	}

	return data
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
func handleConnection(c net.Conn) {
	defer c.Close()

	pynezzentials.PrintColorf(pynezzentials.LightCyan, "[🔌SOCKETS] Handling connection...")

	for {
		request, err := parseConnection(c)
		if err != nil {
			if err == io.EOF {
				pynezzentials.PrintDebug("Connection closed by client")
				break
			}
			pynezzentials.PrintError("Error parsing request: " + err.Error())
			break
		}

		pynezzentials.PrintDebug("Request parsed: " + strconv.Itoa(request.Checksum32))

		// Process the request...
		pynezzentials.PrintColorf(pynezzentials.BgGreen, "Received: %+v\n", request)

		// Finally, respond to the client
		err = respond(c, request)
		if err != nil {
			pynezzentials.PrintError("handleConnection: " + err.Error())
			break
		}
	}

}

func respond(c net.Conn, req ipc.IPCRequest) error {
	pynezzentials.PrintDebug("Responding to the client...")
	var response *ipc.IPCRequest
	var err error
	if req.Checksum32 == int(crc(req.Message.Data)) {
		response, err = NewIPCMessage("ipc", ipc.MSG_ACK, []byte("OK"))
	} else {
		fmt.Printf("Request checksum: %v\nCalculated checksum: %v\n", req.Checksum32, crc(req.Message.Data))
		response, err = NewIPCMessage("ipc", ipc.MSG_ERROR, []byte("CHKSUM ERROR"))
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
	pynezzentials.PrintColor(pynezzentials.BgGreen, "🚀 Response sent!")

	responseTime(req.Timestamp)

	return nil
}
