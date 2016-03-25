# General API "Documentation"

This file is meant to be general notes for getting thoughts down on paper. Actual API documentation is done in Swagger and is visible [here](https://byuoitav.github.io/av-api/).

## Requirements
- Updates should post to ServiceNow CMDB
- HATEOS compliant (context/authorization sensitive)

## Route
`Client` -> `WSO2 -> Go/Node/.NET` -> `Fusion/EMS/DMPS`

## Resources
**Physical Systems**  
- Control Processors
  - Any attribute, GET, PUT
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
`[GET, POST, etc.] /room` View and manage all rooms  
`[GET, POST, etc.] /configuration/{building}/{room}` Get and manage the configuration of a room  
`[GET, POST, etc.] /manage/{building}/{room}/{symbol}` Get status and manage attributes/signals of specific symbol  

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
  roomID: "123A",
  address: "10.6.25.415",
  modified: 2012-04-23T18:25:43.511Z,
  symbols: [symbol, symbol]
}
```
```
symbol: {
  name: "Apple TV",
  ID: 123-456-789,
  signalID: 987-654-321,
  roomID: 159-753-852,
  type: 2,
  modified: 2012-04-23T18:25:43.511Z
}
```

## Response Codes
`200` Everything went well  
`400` The HTTP request that was sent to the server has invalid syntax  
`401` You are unauthorized to perform that action or view that data  
`404` Resource not found  
`500` There was a problem with the server on our end  
`503` The server is overloaded or under maintenance  

## References
[http://timelessrepo.com/haters-gonna-hateoas](http://timelessrepo.com/haters-gonna-hateoas)  
[http://martinfowler.com/articles/richardsonMaturityModel.html](http://martinfowler.com/articles/richardsonMaturityModel.html)  
[https://en.wikipedia.org/wiki/Representational_state_transfer#Architectural_constraints](https://en.wikipedia.org/wiki/Representational_state_transfer#Architectural_constraints)  
[http://www.ics.uci.edu/~fielding/pubs/dissertation/rest_arch_style.htm](http://www.ics.uci.edu/~fielding/pubs/dissertation/rest_arch_style.htm)  
