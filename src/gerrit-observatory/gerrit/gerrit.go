package gerrit

import (
	"bufio"
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"os"
	"time"
)

var (
	gerritCommand = "gerrit stream-events"
)

type retriableError error

// EventStream struct
type EventStream struct {
	Channel chan map[string]interface{}
	config  *ssh.ClientConfig
	addr    string
	deamon  bool
}

// NewEventStream EventStream initialize
func NewEventStream(port int, user string, hostname string, privateKey string) (eventStream *EventStream, err error) {
	signer, err := makeSigner(privateKey)
	if err != nil {
		return
	}

	sshConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{ssh.PublicKeys(signer)},
	}
	sshConfig.SetDefaults()

	eventStream = &EventStream{
		addr:    fmt.Sprintf("%s:%d", hostname, port),
		Channel: make(chan map[string]interface{}, 100),
		config:  sshConfig,
	}
	return
}

// SetDeamon recover es.Run if loop exit
func (es *EventStream) SetDeamon() {
	es.deamon = true
}

// Run eventStream loop
func (es *EventStream) Run() {
	defer func() {
		if e := recover(); e != nil {
			fmt.Printf("panic in EventStream: %s \n", e)
			if _, ok := e.(retriableError); ok && es.deamon {
				// log here
				time.Sleep(time.Second * 2)
				go es.Run()
			}
		}
	}()

	client, err := ssh.Dial("tcp", es.addr, es.config)
	if err != nil {
		panic(err)
	}
	defer client.Close()
	go es.createSession(client)

	err = client.Conn.Wait()
	if err != nil {
		panic(err)
	}
}

func (es *EventStream) createSession(client *ssh.Client) {
	for {
		session, err := client.NewSession()
		if err != nil {
			panic(err)
		}
		defer session.Close()
		session.Stdout = nil
		session.Stderr = nil
		stdout, err := session.StdoutPipe()
		if err != nil {
			panic(err)
		}
		stderr, err := session.StderrPipe()
		if err != nil {
			panic(err)
		}
		go es.stdoutParser(stdout)
		go es.stderrParser(stderr)
		session.Start(gerritCommand)
		session.Wait()
	}
}

func (es *EventStream) stdoutParser(stdout io.Reader) {
	var raw = make(map[string]interface{})
	decoder := json.NewDecoder(stdout)
	decoder.UseNumber()
	for {
		err := decoder.Decode(&raw)
		if err != nil {
			// log here
			fmt.Printf("%s\n", err)
			return
		}
		es.Channel <- raw
	}
}

func (es *EventStream) stderrParser(stderr io.Reader) {
	scanner := bufio.NewScanner(stderr)
	for scanner.Scan() {
		//log here
		fmt.Printf("stderr: %s \n", scanner.Text())
	}
}

func makeSigner(keyname string) (signer ssh.Signer, err error) {
	fp, err := os.Open(keyname)
	if err != nil {
		return
	}
	defer fp.Close()

	buf, err := ioutil.ReadAll(fp)
	if err != nil {
		return
	}
	signer, err = ssh.ParsePrivateKey(buf)
	return
}
