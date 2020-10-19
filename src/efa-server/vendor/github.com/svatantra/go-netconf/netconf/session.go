package netconf

import (
	"encoding/xml"
	"strings"
	"unicode"
)

// Session defines the necessary components for a Netconf session
type Session struct {
	Transport          Transport
	SessionID          int
	ServerCapabilities []string
	ErrOnWarning       bool
}

// Close is used to close and end a transport session
func (s *Session) Close() error {
	return s.Transport.Close()
}

// Exec is used to execute an RPC method or methods
func (s *Session) Exec(methods ...RPCMethod) (*RPCReply, error) {
	rpc := NewRPCMessage(methods)

	request, err := xml.Marshal(rpc)
	if err != nil {
		return nil, err
	}

	header := []byte(xml.Header)
	request = append(header, request...)

	log.Debugf("REQUEST: %s\n", request)

	err = s.Transport.Send(request)
	if err != nil {
		return nil, err
	}

	rawXML, err := s.Transport.Receive()
	if err != nil {
		return nil, err
	}
	log.Debugf("REPLY: %s\n", rawXML)

	//https://blog.zikes.me/post/cleaning-xml-files-before-unmarshaling-in-go/
	printOnly := func(r rune) rune {
		if unicode.IsPrint(r) {
			return r
		}
		return -1
	}
	t := strings.Map(printOnly, string(rawXML))


	reply := &RPCReply{}
	reply.RawReply = t

	if err := xml.Unmarshal([]byte(t), reply); err != nil {
    	    return nil, err
	}

	if reply.Errors != nil {
		// We have errors, lets see if it's a warning or an error.
		for _, rpcErr := range reply.Errors {
			if rpcErr.Severity == "error" || s.ErrOnWarning {
				return reply, &rpcErr
			}
		}

	}

	return reply, nil
}

// NewSession creates a new NetConf session using the provided transport layer.
func NewSession(t Transport) *Session {
	s := new(Session)
	s.Transport = t

	// Receive Servers Hello message
	serverHello, _ := t.ReceiveHello()
	s.SessionID = serverHello.SessionID
	s.ServerCapabilities = serverHello.Capabilities

	// Send our hello using default capabilities.
	t.SendHello(&HelloMessage{XMLNms: "urn:ietf:params:xml:ns:netconf:base:1.0", Capabilities: DefaultCapabilities})

	return s
}
