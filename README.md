# General API "Documentation"

[![Run in Postman](https://run.pstmn.io/button.svg)](https://app.getpostman.com/run-collection/535e1aa2df75f17c63f0)

This file is meant to be general notes for getting thoughts down on paper. Actual API documentation is done in Swagger and is visible [here](https://byuoitav.github.io/av-api/).

## Requirements
- Updates should post to ServiceNow CMDB
- HATEOS compliant (context/authorization sensitive)

## Route
`Client` -> `WSO2 -> Go/Node/.NET` -> `Fusion/EMS/DMPS`

## Resources
**Physical Systems**  
- Control Processors
- Touchpanels
- Projectors
- Flatscreens
- Observation System
  - Server
  - Cameras

**Status**  
- Availability of rooms from EMS/GetSignal of DMPS
- Room configuration

## Endpoints
`[GET, POST] /room` View and manage all rooms  
`[GET, PUT, DELETE] /room/{room}` View and manage a single room  
`[GET, POST] /manage/{building}/{room}` View and manage all signals  
`[GET, PUT, DELETE] /manage/{building}/{room}/{signal}` View and manage a single signal  

## Response Models
```
links: [{
  rel: "self",
  href: "http://[root]"
}, {
  rel: "next",
  href: "http://[root]?page=2"
}]
```
```
room: {
  name: "Test",
  ID: 123-456-789,
  roomID: "123A",
  description: "This room rocks",
  available: true
}
```
```
asset: {
  name: "Test",
  ID: 123-456-789,
  make: "Crestron",
  model: "TSS-123",
  serial: "123456789"
}
```
```
processor: {
  name: "Main",
  ID: 123-456-789,
  address: "10.6.25.415",
  modified: 2012-04-23T18:25:43.511Z,
  signals: [signal, signal]
}
```
```
signal: {
  name: "Apple TV",
  ID: 987-654-321,
  type: 2,
  modified: 2012-04-23T18:25:43.511Z
}
```

## HTTP Response Codes
`200` OK (Everything went well)  
`400` Bad Request (The HTTP request that was sent to the server has invalid syntax)  
`401` Unauthorized (You are unauthorized to perform the requested action or view the requested data)  
`404` Not Found (The service couldn't find what you requested)  
`500` Internal Server Error (There was a problem with the server on our end)  
`503` Service Unavailable (The server is overloaded, is under maintenance, or is otherwise unavailable)  

## References
[http://timelessrepo.com/haters-gonna-hateoas](http://timelessrepo.com/haters-gonna-hateoas)  
[http://martinfowler.com/articles/richardsonMaturityModel.html](http://martinfowler.com/articles/richardsonMaturityModel.html)  
[https://en.wikipedia.org/wiki/Representational_state_transfer#Architectural_constraints](https://en.wikipedia.org/wiki/Representational_state_transfer#Architectural_constraints)  
[http://www.ics.uci.edu/~fielding/pubs/dissertation/rest_arch_style.htm](http://www.ics.uci.edu/~fielding/pubs/dissertation/rest_arch_style.htm)
[http://www.restapitutorial.com/httpstatuscodes.html](http://www.restapitutorial.com/httpstatuscodes.html)
