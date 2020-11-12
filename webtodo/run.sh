#!/bin/bash

go build
export dbuser="dbuser"
export dbpass="ololo"
export dbname="todoDB"
./webtodo
