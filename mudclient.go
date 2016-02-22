package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

var (
	escapedXML = map[string]string{
		"lt":   `<`,
		"gt":   `>`,
		"amp":  `&`,
		"apos": `'`,
		"quot": `"`,
	}
)

type MudConnection struct {
	Conn     net.Conn
	Mapper   *Mapper
	Debug    bool
	XMLMode  bool
	OutChans []chan string
	InChans  []chan string
	TagChans []chan *Tag
}

func main() {
	conn, err := net.Dial("tcp", "mume.org:4242")
	if err != nil {
		log.Fatalf("Error connecting: %s", err.Error())
	}

	_, err = conn.Write([]byte("~$#EX1\n3\n"))
	if err != nil {
		log.Fatalf("Error sending XML mode: %v", err)
	}

	mudConn := &MudConnection{
		Conn: conn,
	}
	mudConn.Mapper = NewMapper(mudConn)

	mudConn.startScanners()
}

func (self *MudConnection) startScanners() {
	go self.startInputScanner()
	go self.Mapper.Start()
	self.startOutputScanner()
}

func (self *MudConnection) startOutputScanner() {
	scanner := bufio.NewScanner(self.Conn)
	scanner.Split(bufio.ScanBytes)
	scannerChar := ""
	lineBuf := ""

	inEscaped := false
	escapedBuf := ""

	inTag := false
	tagBuf := ""
	var currentTag *Tag

	for scanner.Scan() {
		if scanner.Err() != nil {
			log.Fatalf("Error reading from server: %s", scanner.Err().Error())
		} else {
			scannerChar = scanner.Text()

			// Handle XML tags - <room ...>
			if !inTag {
				if scannerChar == "<" {
					inTag = true
				}
			} else {
				if scannerChar == ">" {
					if !strings.HasPrefix(tagBuf, "/") {
						// A beginning tag (<tag> or <tag />) - set currentTag and deal with tag contents
						currentTag = &Tag{
							Parent:     currentTag,
							Attributes: make(map[string]string),
						}

						tagParts := strings.Fields(strings.Trim(tagBuf, "/ "))
						currentTag.Name = tagParts[0]
						for i := 1; i < len(tagParts); i++ {
							attributeParts := strings.SplitN(tagParts[i], "=", 2)
							if len(attributeParts) == 2 {
								currentTag.Attributes[attributeParts[0]] = strings.Trim(attributeParts[1], `'"`)
							}
						}
					}

					// Figure out if we're closing a tag. (Either </tag> or <tag />)
					if strings.HasPrefix(tagBuf, "/") || strings.HasSuffix(tagBuf, "/") {
						// Publish closed tag to channels
						for _, tagChan := range self.TagChans {
							tagChan <- currentTag
						}
						if currentTag.Parent != nil {
							currentTag.Parent.Children = append(currentTag.Parent.Children, currentTag)
						}
						currentTag = currentTag.Parent
					}

					scannerChar = ""
					tagBuf = ""
					inTag = false
				} else {
					tagBuf += scannerChar
				}
			}

			// Handle escaped characters - &gt; and the like.
			if !inEscaped {
				if scannerChar == "&" {
					inEscaped = true
				}
			} else {
				if scannerChar == ";" {
					if char, ok := escapedXML[escapedBuf]; ok {
						scannerChar = char
					} else {
						log.Printf("UNKNOWN XML ESCAPE SEQ: %v", escapedBuf)
					}
					escapedBuf = ""
					inEscaped = false
				} else {
					escapedBuf += scannerChar
				}
			}

			if !inEscaped && !inTag {
				fmt.Print(scannerChar)
				if currentTag != nil {
					currentTag.TextContent += scannerChar
				}
			}

			if len(scannerChar) > 0 && scannerChar[0] == 13 {
				if len(lineBuf) > 0 {
					for _, outChan := range self.OutChans {
						outChan <- lineBuf
					}
					lineBuf = ""
				}
			} else {
				lineBuf += scannerChar
			}
		}
	}
}

func (self *MudConnection) startInputScanner() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		commandLine := scanner.Text()
		if !self.interceptInput(commandLine) {
			self.sendLine(commandLine)
		}
	}
}

func (self *MudConnection) sendLine(line string) {
	_, err := self.Conn.Write([]byte(fmt.Sprintf("%s\n\r", line)))
	if err != nil {
		log.Fatalf("Error writing: %s", err.Error())
	}
	for _, inChan := range self.InChans {
		inChan <- line
	}
}

func (self *MudConnection) interceptInput(line string) bool {
	if len(strings.TrimSpace(line)) == 0 {
		return false
	}

	command := strings.Fields(line)[0]
	// args := strings.Fields(line)[1:]
	intercepted := true
	switch command {
	case "debug":
		self.Debug = !self.Debug
		fmt.Printf("[[ Debug mode: %v ]]\n", self.Debug)
	case "printmap":
		fmt.Printf("\n[[ Map data coming up... ]]\n")
		self.Mapper.PrintData()
	default:
		intercepted = false
	}

	return intercepted
}
