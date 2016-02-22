package main

import (
	"fmt"
	"strings"
)

type ExitDir string

var (
	DIR_NORTH = ExitDir("north")
	DIR_EAST  = ExitDir("east")
	DIR_SOUTH = ExitDir("south")
	DIR_WEST  = ExitDir("west")
	DIR_UP    = ExitDir("up")
	DIR_DOWN  = ExitDir("down")
)

func ExitDirFromString(dirName string) (ExitDir, error) {
	dirName = strings.Trim(dirName, ",. ")
	switch dirName {
	case "north":
		return DIR_NORTH, nil
	case "east":
		return DIR_EAST, nil
	case "south":
		return DIR_SOUTH, nil
	case "west":
		return DIR_WEST, nil
	case "up":
		return DIR_UP, nil
	case "down":
		return DIR_DOWN, nil
	default:
		return ExitDir(""), fmt.Errorf("UNKNOWN EXIT DIRECTION: %v", dirName)
	}
}

func (self ExitDir) Reverse() (ExitDir, error) {
	switch self {
	case DIR_NORTH:
		return DIR_SOUTH, nil
	case DIR_EAST:
		return DIR_WEST, nil
	case DIR_SOUTH:
		return DIR_NORTH, nil
	case DIR_WEST:
		return DIR_EAST, nil
	case DIR_UP:
		return DIR_DOWN, nil
	case DIR_DOWN:
		return DIR_UP, nil
	}
	return ExitDir(""), fmt.Errorf("REVERSE: UNKNOWN EXIT DIRECTION: %v", self)
}

type ExitFlag string

var (
	FLAG_DOOR      = ExitFlag("door")       // [dir], (dir), or #dir#
	FLAG_PORTAL    = ExitFlag("portal")     // {dir}
	FLAG_UPCLIMB   = ExitFlag("up_climb")   // /dir\
	FLAG_DOWNCLIMB = ExitFlag("down_climb") // \dir/
	FLAG_ROAD      = ExitFlag("road")       // =dir=
	FLAG_TRAIL     = ExitFlag("trail")      // -dir-
	FLAG_SWIMMING  = ExitFlag("swimming")   // ~dir~
)

type Exit struct {
	Direction ExitDir
	Flags     []ExitFlag
}

func NewExits(exitString string) (exits []*Exit) {
	exitString = strings.TrimSpace(exitString)
	exitString = strings.TrimPrefix(exitString, "Exits: ")
	exitString = strings.TrimSuffix(exitString, ".")

	for _, exitName := range strings.Fields(exitString) {
		exit := &Exit{}
	ExitCharLoop:
		for _, char := range exitName {
			switch char {
			case '[':
				exit.Flags = append(exit.Flags, FLAG_DOOR)
			case '(':
				exit.Flags = append(exit.Flags, FLAG_DOOR)
			case '#':
				exit.Flags = append(exit.Flags, FLAG_DOOR)
			case '{':
				exit.Flags = append(exit.Flags, FLAG_PORTAL)
			case '/':
				exit.Flags = append(exit.Flags, FLAG_UPCLIMB)
			case '\\':
				exit.Flags = append(exit.Flags, FLAG_DOWNCLIMB)
			case '=':
				exit.Flags = append(exit.Flags, FLAG_ROAD)
			case '-':
				exit.Flags = append(exit.Flags, FLAG_TRAIL)
			case '~':
				exit.Flags = append(exit.Flags, FLAG_SWIMMING)
			case 'n':
				exit.Direction = DIR_NORTH
				break ExitCharLoop
			case 'e':
				exit.Direction = DIR_EAST
				break ExitCharLoop
			case 's':
				exit.Direction = DIR_SOUTH
				break ExitCharLoop
			case 'w':
				exit.Direction = DIR_WEST
				break ExitCharLoop
			case 'u':
				exit.Direction = DIR_UP
				break ExitCharLoop
			case 'd':
				exit.Direction = DIR_DOWN
				break ExitCharLoop
			default:
				fmt.Printf("\n[[ UNRECOGNIZED EXIT STRING: '%v' (char:%s) ]]\n", exitName, char)
				break ExitCharLoop
			}
		}
		exits = append(exits, exit)
	}

	return
}

type Room struct {
	ID           int
	Name         string
	Description  string
	Exits        map[ExitDir]*Exit
	ExitPointers map[ExitDir]*Room `json:"-"`
}

func NewRoom(id int) *Room {
	return &Room{
		ID:           id,
		Exits:        make(map[ExitDir]*Exit),
		ExitPointers: make(map[ExitDir]*Room),
	}
}
