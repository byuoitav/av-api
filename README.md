# Brigham Young University Audiovisual API  

[![Run in Postman](https://run.pstmn.io/button.svg)](https://app.getpostman.com/run-collection/6a1964154481b8b2114c) [![View in Swagger](http://www.jessemillar.com/view-in-swagger-button/button.svg)](https://byuoitav.github.io/av-api/)

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

## References
[http://timelessrepo.com/haters-gonna-hateoas](http://timelessrepo.com/haters-gonna-hateoas)  
[http://martinfowler.com/articles/richardsonMaturityModel.html](http://martinfowler.com/articles/richardsonMaturityModel.html)  
[https://en.wikipedia.org/wiki/Representational_state_transfer#Architectural_constraints](https://en.wikipedia.org/wiki/Representational_state_transfer#Architectural_constraints)  
[http://www.ics.uci.edu/~fielding/pubs/dissertation/rest_arch_style.htm](http://www.ics.uci.edu/~fielding/pubs/dissertation/rest_arch_style.htm)
[http://www.restapitutorial.com/httpstatuscodes.html](http://www.restapitutorial.com/httpstatuscodes.html)
