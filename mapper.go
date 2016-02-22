package main

import (
	"encoding/json"
	"fmt"
)

type Mapper struct {
	tags        chan *Tag
	connection  *MudConnection
	rooms       []*Room
	roomsByID   map[int]*Room
	currentRoom *Room
	nextID      int
}

func NewMapper(conn *MudConnection) *Mapper {
	seedRoom := NewRoom(1)
	tags := make(chan *Tag, 1000)
	conn.TagChans = append(conn.TagChans, tags)

	return &Mapper{
		tags:       tags,
		connection: conn,
		rooms:      []*Room{seedRoom},
		roomsByID: map[int]*Room{
			1: seedRoom,
		},
		currentRoom: seedRoom,
		nextID:      2,
	}
}

func (self *Mapper) Start() {
	for {
		tag := <-self.tags
	TagSwitch:
		switch tag.Name {
		case "movement":
			if dir, ok := tag.Attributes["dir"]; ok {
				// This means we've moved.
				exitDir, err := ExitDirFromString(dir)
				if err != nil {
					fmt.Printf("\n[[ %v ]]\n", err)
					break TagSwitch
				}

				reverseDir, err := exitDir.Reverse()
				if err != nil {
					fmt.Printf("\n[[ %v ]]\n", err)
					break TagSwitch
				}

				nextRoom := self.currentRoom.ExitPointers[exitDir]
				if nextRoom == nil {
					nextRoom = NewRoom(self.nextID)
					self.rooms = append(self.rooms, nextRoom)
					self.roomsByID[nextRoom.ID] = nextRoom
					self.currentRoom.ExitPointers[exitDir] = nextRoom
					self.nextID++
				}
				nextRoom.ExitPointers[reverseDir] = self.currentRoom
				self.currentRoom = nextRoom
				fmt.Printf("\n[[ ENTERING ROOM: %v (#%v) ]]\n", self.currentRoom.Name, self.currentRoom.ID)
			} else {
				fmt.Printf("\n[[ UNHANDLED MOVEMENT TAG: %v '%v' ]]\n", tag.Attributes, tag.TextContent)
			}
		case "room":
			if nameTag := tag.FirstChild("name"); nameTag != nil {
				self.currentRoom.Name = nameTag.TextContent
			}
			if descriptionTag := tag.FirstChild("description"); descriptionTag != nil {
				self.currentRoom.Description = descriptionTag.TextContent
			}
		case "exits":
			for _, exit := range NewExits(tag.TextContent) {
				self.currentRoom.Exits[exit.Direction] = exit
			}
		}
	}
}

func (self *Mapper) PrintData() {
	roomBytes, err := json.MarshalIndent(self.rooms, "", "    ")
	if err != nil {
		fmt.Printf("Error marshalling rooms: %v\n", err)
	} else {
		fmt.Println(string(roomBytes))
	}
}
