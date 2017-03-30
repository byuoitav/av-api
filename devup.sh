#!/usr/bin/env bash

# Build binaries
echo ---------- av-api ---------- 
go build -v -o av-api
echo ---------- telnet-microservice ---------- 
go build -v -o ../telnet-microservice/telnet-microservice ../telnet-microservice
echo ---------- crestron-control-microservice ---------- 
go build -v -o ../crestron-control-microservice/crestron-control-microservice ../crestron-control-microservice
echo ---------- pjlink-microservice ---------- 
go build -v -o ../pjlink-microservice/pjlink-microservice ../pjlink-microservice
echo ---------- configuration-database-microservice ---------- 
go build -v -o ../configuration-database-microservice/configuration-database-microservice ../configuration-database-microservice
echo ---------- sony-control-microservice ---------- 
go build -v -o ../sony-control-microservice/sony-control-microservice ../sony-control-microservice
echo ---------- london-audio-microservice ---------- 
go build -v -o ../london-audio-microservice/london-audio-microservice ../london-audio-microservice
echo ---------- cgi-microservice ---------- 
go build -v -o ../cgi-microservice/cgi-microservice ../cgi-microservice
echo ---------- pulse-eight-neo-microservice ---------- 
go build -v -o ../pulse-eight-neo-microservice/pulse-eight-neo-microservice ../pulse-eight-neo-microservice
echo ---------- adcp-control-microservice ---------- 
go build -v -o ../adcp-control-microservice/adcp-control-microservice ../adcp-control-microservice
echo ---------- av-api-rpc ---------- 
go build -v -o ../av-api-rpc/av-api-rpc ../av-api-rpc
echo ---------- touchpanel-ui-microservice ---------- 
go build -v -o ../touchpanel-ui-microservice/touchpanel-ui-microservice ../touchpanel-ui-microservice
echo ---------- configuration-database-tool ---------- 
go build -v -o ../configuration-database-tool/configuration-database-tool ../configuration-database-tool

# Build Docker containers
docker-compose -f docker-compose-dev.yml build
docker-compose -f docker-compose-dev.yml up -d
