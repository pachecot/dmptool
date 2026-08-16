package dmp

import "strings"

type EEEventParameters struct {
	TimeDelay    string
	StateValues  string // ChangeOfState: comma-joined triggering values from listOfValues
	COVIncrement string // ChangeOfValue: increment threshold from covCriteria
	LowLimit     string // OutOfRange
	HighLimit    string // OutOfRange
	Deadband     string // OutOfRange
}

func ParseEEEventParameters(s string) EEEventParameters {
	var ep EEEventParameters
	var currentBlock string
	var stateValues []string
	for _, line := range strings.Split(s, "\n") {
		k, v, ok := split(line)
		if !ok {
			continue
		}
		switch k {
		case "timeDelay":
			ep.TimeDelay = v
		case "lowLimit":
			ep.LowLimit = v
		case "highLimit":
			ep.HighLimit = v
		case "deadband":
			ep.Deadband = v
		case "covCriteria":
			currentBlock = "covCriteria"
		case "listOfValues":
			currentBlock = "listOfValues"
		case "Value":
			if currentBlock == "covCriteria" {
				ep.COVIncrement = v
			}
		case "value":
			if currentBlock == "listOfValues" {
				stateValues = append(stateValues, v)
			}
		}
	}
	ep.StateValues = strings.Join(stateValues, ",")
	return ep
}

func ParseObjPropRef(s string) string {
	for _, line := range strings.Split(s, "\n") {
		k, v, ok := split(line)
		if !ok {
			continue
		}
		if k == "objectIdentifier" {
			return v
		}
	}
	return ""
}
