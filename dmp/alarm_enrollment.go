package dmp

import "strings"

type EventParameters struct {
	TimeDelay     string
	ACCExpression string
	Limit         string
	Normal        string
}

func ParseEventParameters(s string) EventParameters {
	var ep EventParameters
	for _, line := range strings.Split(s, "\n") {
		k, v, ok := split(line)
		if !ok {
			continue
		}
		switch k {
		case "timeDelay":
			ep.TimeDelay = v
		case "ACCExpression":
			ep.ACCExpression = v
		case "Limit":
			ep.Limit = v
		case "Normal":
			ep.Normal = v
		}
	}
	return ep
}
