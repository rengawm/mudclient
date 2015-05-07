package main 

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

type MudConnection struct {
	Conn net.Conn
	RunningScripts []Script
}

func main() {
	conn, err := net.Dial("tcp", "legendsofthejedi.com:5656")
//	conn, err := net.Dial("tcp", "arkmud.org:4200")
	if err != nil { log.Fatalf("Error connecting: %s", err.Error()) }
	
	mudConn := new(MudConnection)
	mudConn.Conn = conn
	mudConn.RunningScripts = make([]Script, 0)
	mudConn.startScanners()
}

func (mudConn *MudConnection) startScanners() {
	go mudConn.startInputScanner()
	mudConn.startOutputScanner()
}

func (mudConn *MudConnection) startOutputScanner() {
	scanner := bufio.NewScanner(mudConn.Conn)
	scanner.Split(bufio.ScanBytes)
	scannerChar := ""
	lineBuf := ""
	
	for scanner.Scan() {
		if scanner.Err() != nil {
			log.Fatalf("Error reading from server: %s", scanner.Err().Error())
		} else {
			scannerChar = scanner.Text()
			fmt.Print(scannerChar)
			
			if scannerChar[0] == 13 {
				if len(lineBuf) > 0 {
					mudConn.checkOutputForTriggers(lineBuf)
					lineBuf = ""
				}
			} else {
				lineBuf = fmt.Sprintf("%s%s", lineBuf, scannerChar)
			}
		}
	}
}

func (mudConn *MudConnection) startInputScanner() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		commandLine := scanner.Text()
		if !mudConn.interceptInput(commandLine) {
			mudConn.sendLine(commandLine)
		}
	}
}

func (mudConn *MudConnection) sendLine(line string) {
	fmt.Printf("<< %s\n", line)
	_, err := mudConn.Conn.Write([]byte(fmt.Sprintf("%s\n\r", line)))
	if err != nil { log.Fatalf("Error writing: %s", err.Error()) }
}

func (mudConn *MudConnection) checkOutputForTriggers(line string) {
	// Sanitize the line - remove all color codes, newline, etc.
	escaping := false
	cleanLine := ""
	for i := 0; i < len(line); i++ {
		if line[i] == 0x1B {
			escaping = true
		} else {
			if !escaping {
				cleanLine = fmt.Sprintf("%s%c", cleanLine, line[i])
			} else {
				if line[i] == 'm' {
					escaping = false
				}
			}
		}
	}
	
	line = strings.TrimSpace(cleanLine)
	
	for _, script := range mudConn.RunningScripts {
		script.SendOutput(line)
	}
}

func (mudConn *MudConnection) interceptInput(line string) bool {
	if len(strings.TrimSpace(line)) == 0 {
		return false
	}

	args := strings.Fields(line)[1:]
	intercepted := true
	switch line {
		case "autoresearch":
			script := &ResearchScript{&BaseScript{}}
			script.Execute(args, script, mudConn)
		case "autoponder":
			script := &PonderScript{&BaseScript{}}
			script.Execute(args, script, mudConn)
		case "stopscripts":
			mudConn.RunningScripts = make([]Script, 0)
			fmt.Println("All scripts aborted.")
		case "listscripts":
			fmt.Printf("Running scripts: %q\n", mudConn.RunningScripts)
		default:
			intercepted = false
	}
	
	return intercepted
}