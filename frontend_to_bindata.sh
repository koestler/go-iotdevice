#!/usr/bin/env bash

cp -r ~/git/react-ve-sensor/build/* ./frontend

go-bindata -o httpServer/frontend-bindata.go -prefix "./frontend/" -pkg httpServer -nomemcopy -nocompress `find ./frontend/ -type d`
