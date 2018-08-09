#!/bin/bash

# making checks to see that named is answering
dig +time=1 +tries=1 +short server.example.com @127.0.0.1 >/dev/null 2>&1

if [ "$?" -eq "0" ]; then
        exit 0
else
        exit 1
fi
