#!/bin/bash

go build wiki.go
mv wiki build/
./build/wiki