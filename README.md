# General API "Documentation"

## Layout
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

## Possible Endpoints
`[GET, POST, etc.] /rooms` View all rooms  
`[GET, POST, etc.] /configuration/{building}/{room}` Get and manage the configuration of a room  
`[GET, POST, etc.] /manage/{building}/{room}/{system}` Get status and manage attributes/signals of specific system  

## Notes
- Updates should post to ServiceNow CMDB

## References
[http://martinfowler.com/articles/richardsonMaturityModel.html](http://martinfowler.com/articles/richardsonMaturityModel.html)  
[http://timelessrepo.com/haters-gonna-hateoas](http://timelessrepo.com/haters-gonna-hateoas)  
[https://en.wikipedia.org/wiki/Representational_state_transfer#Architectural_constraints](https://en.wikipedia.org/wiki/Representational_state_transfer#Architectural_constraints)  
[http://www.ics.uci.edu/~fielding/pubs/dissertation/rest_arch_style.htm](http://www.ics.uci.edu/~fielding/pubs/dissertation/rest_arch_style.htm)  
