package client

import (
	"errors"
	"fmt"
	"log"
	"net"

	"github.com/beevik/etree"
	"github.com/svatantra/go-netconf/netconf"
	"golang.org/x/crypto/ssh"
	"strings"
)

const (
	//NetConfError error String
	NetConfError = "netconf rpc [error] "
)

//NetconfClient contains the info needed to establish, maintain and close the Netconf session to the switch.
type NetconfClient struct {
	Host     string
	User     string
	Password string
	Session  *netconf.Session
}

//Login will be used to login to the Netconf session to the switch.
func (n *NetconfClient) Login() error {
	sshConfig := &ssh.ClientConfig{
		Config: ssh.Config{
			Ciphers: []string{"aes128-cbc", "hmac-sha1"},
		},
		User: n.User,
		Auth: []ssh.AuthMethod{ssh.Password(n.Password)},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	s, err := netconf.DialSSH(n.Host, sshConfig)

	if err != nil {
		log.Println("Failed to Login", err)
	}
	n.Session = s
	return err
}

//GetConfig will be used to get the "running-config" from the switch.
func (n *NetconfClient) GetConfig(data string) (string, error) {
	config := `<get-config>
					<source>
						<running></running>
					</source>
					<filter type="xpath" select="%s"></filter>
			   </get-config>`
	request := fmt.Sprintf(config, data)
	//fmt.Println(request)

	reply, err := n.Session.Exec(netconf.RawMethod(request))
	//t.Println(reply)
	if err != nil {
		return "", err
	}

	return reply.Data, nil
}

//EditConfig will be used to edit the "running-config" on the switch.
func (n *NetconfClient) EditConfig(data string) (string, error) {
	preConfig :=
		`<edit-config>
		<target>
		  	<running></running>
		</target>`

	postConfig :=
		`</edit-config>`

	request := preConfig + data + postConfig
	reply, respErr := n.Session.Exec(netconf.RawMethod(request))

	if respErr != nil {
		respMessage := respErr.Error()
		if respMessage == "netconf rpc [error] ''" || (respMessage == "netconf rpc [warning] ''") {
			doc := etree.NewDocument()
			if err := doc.ReadFromBytes([]byte(reply.Data)); err != nil {
				fmt.Println("Error in edit-config response")
				return "", respErr
			}

			de := doc.FindElement("//rpc-error")
			if de != nil {
				de = doc.FindElement("//error-tag")
				if de != nil {
					if de.Text() == "access-denied" {
						return "", errors.New("%%Error: User is not authorized to perform this operation")
					}
					return "", errors.New(de.Text())
				}
			}
		} else if strings.Contains(respMessage, NetConfError) {
			return "", errors.New(extractMessage(respMessage, NetConfError))
		} else {
			return "", respErr
		}
	}

	return reply.Data, nil
}

func extractMessage(respMessage string, subString string) string {
	pos := strings.LastIndex(respMessage, subString)
	if pos == -1 {
		return respMessage
	}
	adjustedPos := pos + len(subString)
	if adjustedPos >= len(respMessage) {
		return respMessage
	}
	return respMessage[adjustedPos:len(respMessage)]
}

//ExecuteRPC will be used to execute the netconf rpc on the switch.
func (n *NetconfClient) ExecuteRPC(data string) (string, error) {

	reply, err := n.Session.Exec(netconf.RawMethod(data))

	if err != nil {
		return "", err
	}
	//fmt.Printf(reply.Data)
	return reply.Data, nil

}

//Close will be used to close the Netconf session to the switch.
func (n *NetconfClient) Close() error {
	return n.Session.Close()
}
