#!/usr/bin/env bash
go-bindata -prefix "frontend/" -pkg main -nomemcopy -nocompress `find frontend/ -type d`
