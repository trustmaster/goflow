#!/bin/bash
go install -buildmode=plugin test_plugins/Adder_goplug.go
go install -buildmode=plugin test_plugins/Plug1_goplug.go
go install -buildmode=plugin test_plugins/NGen_goplug.go
