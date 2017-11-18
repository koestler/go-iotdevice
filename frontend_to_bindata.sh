#!/usr/bin/env bash
go-bindata -prefix "../frontend/" -pkg webserver -nomemcopy -nocompress `find ../frontend/ -type d`
