package client

import (
	"golang.org/x/crypto/ssh"
	"io"
	"log"
	"net"
	"strings"
)

//SSHClient contains info needed to establish, maintain and close a SSH session.
type SSHClient struct {
	Host     string
	User     string
	Password string
	Session  *ssh.Session
	Client   *ssh.Client
	Stdin    io.WriteCloser
	Stdout   io.Reader
}

//Login will be used to SSH login to the switch.
func (n *SSHClient) Login() error {

	config := &ssh.ClientConfig{
		User: n.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(n.Password),
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	// Connect to the remote server and perform the SSH handshake.
	client, err := ssh.Dial("tcp", n.Host+":22", config)

	if err != nil {
		log.Fatal("Failed to Dial: ", err)
		return err
	}
	n.Client = client

	session, err := client.NewSession()
	if err != nil {
		log.Fatal("Failed to create session: ", err)
		return err
	}

	stdin, err := session.StdinPipe()
	if err != nil {
		log.Fatal(err)
		return err
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	/*
			 * To display output of the command
		    session.Stdout = os.Stdout
		    session.Stderr = os.Stderr
	*/

	if err := session.Shell(); err != nil {
		log.Fatal(err)
		return err
	}
	n.Session = session
	n.Stdin = stdin
	n.Stdout = stdout

	return err
}

//ExecuteOperationalCommand will be used to execute the operational command on the switch.
func (n *SSHClient) ExecuteOperationalCommand(command string) string {

	//stdin.Write([]byte("terminal length 0" + "\n"))
	suffix := "EXTREME SLX-CLI"
	n.Stdin.Write([]byte(command + " ; oscmd echo \"" + suffix + "\" \n"))

	buf := make([]byte, 1000)
	num, err := n.Stdout.Read(buf) //this reads the ssh terminal welcome message
	loadStr := ""
	if err == nil {
		loadStr = string(buf[:num])
	}
	for (err == nil) && (!strings.Contains(loadStr, suffix)) {
		num, err = n.Stdout.Read(buf)
		loadStr += string(buf[:num])
	}

	loadStr = loadStr[:len(loadStr)-len(suffix)-1]
	return loadStr
}

//ExecuteConfigCommand will be used to execute the config command on the switch.
func (n *SSHClient) ExecuteConfigCommand(command string) string {

	//stdin.Write([]byte("terminal length 0" + "\n"))
	suffix := "EXTREME SLX-CLI"
	n.Stdin.Write([]byte("configure terminal" + "\n"))
	n.Stdin.Write([]byte(command + " ; do oscmd echo \"" + suffix + "\" \n"))

	buf := make([]byte, 1000)
	num, err := n.Stdout.Read(buf) //this reads the ssh terminal welcome message
	loadStr := ""
	if err == nil {
		loadStr = string(buf[:num])
	}
	for (err == nil) && (!strings.Contains(loadStr, suffix)) {
		num, err = n.Stdout.Read(buf)
		loadStr += string(buf[:num])
	}
	loadStr = loadStr[:len(loadStr)-len(suffix)-1]

	n.Stdin.Write([]byte("top" + "\n"))
	n.Stdin.Write([]byte("exit" + "\n"))
	return loadStr

}

//Close will be used to close the SSH session to the switch.
func (n *SSHClient) Close() {
	n.Client.Close()
	n.Session.Close()
	n.Session.Wait()
	log.Println("SshClient session close for the host : ", n.Host)
}
