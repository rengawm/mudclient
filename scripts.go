package main

import (
	"fmt"
	"strconv"
	"strings"
)


func GetOutputUntilPrompt(mudOutput chan string) (cmdOutput string) {
	cmdOutput = ""
	LineLoop:
		for {
			select {
				case line := <-mudOutput:
					if strings.HasPrefix(line, "HP: ") && strings.HasSuffix(line, ">") {
						break LineLoop
					}
					cmdOutput = cmdOutput+"\n"+line
			}
		}

	return cmdOutput
}

type Script interface {
	Execute([]string, Script, *MudConnection)
	innerExecute([]string)
	SendOutput(string)
}

type BaseScript struct {
	Script
	MudOutput chan string
	MudConn *MudConnection
}

func (script *BaseScript) Execute(args []string, outerScript Script, mudConn *MudConnection) {
	script.MudOutput = make(chan string)
	script.MudConn = mudConn

	mudConn.RunningScripts = append(mudConn.RunningScripts, outerScript)
	go script.invokeInnerExecute(args, outerScript)
}

func (script *BaseScript) invokeInnerExecute(args []string, outerScript Script) {
	outerScript.innerExecute(args)
	s := script.MudConn.RunningScripts
	
	for i := 0; i < len(s); i++ {
		if outerScript == s[i] {
		    if i < len(s)-1 {
		        script.MudConn.RunningScripts = append(s[:i], s[i+1:]...)
		    } else {
		        script.MudConn.RunningScripts = s[:i]
		    }
			return
		}
	}
	fmt.Errorf("Could not find running script in scripts collection")
}

func (script *BaseScript) SendOutput(line string) {
	script.MudOutput <- line
}


type ResearchScript struct {
	*BaseScript
}

func (script *ResearchScript) innerExecute(args []string) {
	researchQueue := make([]string, 0)

	script.MudConn.sendLine("practice")
	practiceOutput := GetOutputUntilPrompt(script.MudOutput)
	for _, skillLine := range strings.Split(practiceOutput, "\n") {
		if strings.Contains(skillLine, "%") {
			for _, skill := range strings.Split(skillLine, "%") {
				skill = strings.TrimSpace(skill)
				if len(skill) > 0 {
					skillParts := strings.Split(skill, "  ")
					skillLevel, err := strconv.ParseInt(skillParts[1], 0, 0)
					if err != nil { fmt.Errorf("Error in parseInt: %s", err.Error()) }
					
					if (skillLevel < 76) {
						researchQueue = append(researchQueue, skillParts[0])
					}
				}
			}
		}
	}
	
	for _, skill := range researchQueue {
		script.MudConn.sendLine(fmt.Sprintf("research %s", skill))

		ResearchLoop:
		for {
			select {
				case line := <-script.MudOutput:
					if strings.Contains(line, "You can't learn ") {
						break ResearchLoop
					} else if line == "You study for hours on end, but fail to gather any knowledge." || line == "You finish your studies and feel much more skilled." {
						script.MudConn.sendLine(fmt.Sprintf("research %s", skill))
					}
			}
		}
	}
}


type PonderScript struct {
	*BaseScript
}

func (script *PonderScript) innerExecute(args []string) {
	script.MudConn.sendLine("ponder")
	
	for {
		select {
			case line := <-script.MudOutput:
				if strings.Contains(line, "You ponder for some time, but fail to figure anything out") {
					script.MudConn.sendLine("ponder")
				} else if strings.Contains(line, "You ponder for some time, and things seem clearer.") {
					script.MudConn.sendLine("ponder")
				}
		}
	}
}
