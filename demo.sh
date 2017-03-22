#!/bin/bash

#prompt user for ip address, buidling, and room
export ADDRESS
echo input IP address of API host
read ADDRESS

export BUILDING
echo enter building
read BUILDING

export ROOM
echo enter room
read ROOM

#power room on and off
echo Powering room on and off
echo
curl -H "Content-Type: application/json" -X PUT -d '{"power":"on"}' $ADDRESS:8000/buildings/$BUILDING/rooms/$ROOM
echo
sleep 20ss

curl -H "Content-Type: application/json" -X PUT -d '{"power":"standby"}' $ADDRESS:8000/buildings/$BUILDING/rooms/$ROOM
echo
sleep 10s

#switching video inputs
echo Switching video inputs
echo
curl -H "Content-Type: application/json" -X PUT -d '{"power":"on", "currentVideoInput":"HDMIIn"}' $ADDRESS:8000/buildings/$BUILDING/rooms/$ROOM
echo 
sleep 10s

#blanking and unblanking
echo blanking and unblanking
echo
curl -H "Content-Type: application/json" -X PUT -d '{"blanked":true}' $ADDRESS:8000/buildings/$BUILDING/rooms/$ROOM
echo
sleep 10s

curl -H "Content-Type: application/json" -X PUT -d '{"blanked":false}' $ADDRESS:8000/buildings/$BUILDING/rooms/$ROOM 
echo
sleep 10s

#muting and unmuting
echo muting and unmuting
echo
curl -H "Content-Type: application/json" -X PUT -d '{"muted":true}' $ADDRESS:8000/buildings/$BUILDING/rooms/$ROOM 
sleep 10s

curl -H "Content-Type: application/json" -X PUT -d '{"muted":false}' $ADDRESS:8000/buildings/$BUILDING/rooms/$ROOM 
echo
sleep 10s

#powering off
echo shutting down
curl -H "Content-Type: application/json" -X PUT -d '{"power":"standby"}' $ADDRESS:8000/buildings/$BUILDING/rooms/$ROOM 
