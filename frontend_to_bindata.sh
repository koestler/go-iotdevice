#!/usr/bin/env bash
go-bindata -prefix "../frontend/" -pkg httpServer -nomemcopy -nocompress `find ../frontend/ -type d`
