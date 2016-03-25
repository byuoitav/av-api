# General API "Documentation"

These are meant to be general notes for getting thoughts down on paper. Actual API documentation is done in Swagger and is visible [here](https://byuoitav.github.io/av-api/).

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
- Aggregate of EMS/GetSignal of DMPS Online
- Room configuration

## Endpoints
`[GET, POST, etc.] /room` View and manage all rooms  
`[GET, POST, etc.] /configuration/{building}/{room}` Get and manage the configuration of a room  
`[GET, POST, etc.] /manage/{building}/{room}/{system}` Get status and manage attributes/signals of specific system  

## Response Models
```
links: [{
  rel: "self",
  href: "http://[root]"
}]
```
```
room: {
  guid: 123
}
```
```
system: {
  guid: 123
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
[http://martinfowler.com/articles/richardsonMaturityModel.html](http://martinfowler.com/articles/richardsonMaturityModel.html)  
[http://timelessrepo.com/haters-gonna-hateoas](http://timelessrepo.com/haters-gonna-hateoas)  
[https://en.wikipedia.org/wiki/Representational_state_transfer#Architectural_constraints](https://en.wikipedia.org/wiki/Representational_state_transfer#Architectural_constraints)  
[http://www.ics.uci.edu/~fielding/pubs/dissertation/rest_arch_style.htm](http://www.ics.uci.edu/~fielding/pubs/dissertation/rest_arch_style.htm)  
