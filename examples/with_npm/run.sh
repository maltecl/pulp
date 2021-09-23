#!/usr/bin/sh

browserify web/index.js > web/bundle.js && gen -in main.go -out build/main.go && go run build/main.go 
