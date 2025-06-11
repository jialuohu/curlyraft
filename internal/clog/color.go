package clog

import "github.com/fatih/color"

var CRed = color.New(color.FgRed).Add(color.Bold).SprintfFunc()
var CBlue = color.New(color.FgBlue).Add(color.Bold).SprintfFunc()
var CGreen = color.New(color.FgGreen).Add(color.Bold).SprintfFunc()

var CRedRc = func(funcName string) string {
	return CRed("[RaftCore/%s]", funcName)
}
var CBlueRc = func(funcName string) string {
	return CBlue("[RaftCore/%s]", funcName)
}
var CGreenRc = func(funcName string) string {
	return CGreen("[RaftCore/%s]", funcName)
}

var CRedNode = func(funcName string) string {
	return CRed("[Node/%s]", funcName)
}
var CBlueNode = func(funcName string) string {
	return CBlue("[Node/%s]", funcName)
}
var CGreenNode = func(funcName string) string {
	return CGreen("[Node/%s]", funcName)
}

var CRedRg = func(funcName string) string {
	return CRed("[RaftGateway/%s]", funcName)
}
var CBlueRg = func(funcName string) string {
	return CBlue("[RaftGateway/%s]", funcName)
}
var CGreenRg = func(funcName string) string {
	return CGreen("[RaftGateway/%s]", funcName)
}
