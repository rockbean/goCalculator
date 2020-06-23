@echo off
go get github.com/zserge/lorca
go generate
go build -o goCalc.exe main.go assets.go
