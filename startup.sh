#! /usr/bin/env bash

if [ -f booking ]
then
    echo "Found booking"
    rm -f booking
fi
go build -o booking app/web/*.go
./booking
