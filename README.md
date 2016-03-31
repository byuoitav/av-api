# Brigham Young University Audiovisual API  

[![Run in Postman](https://run.pstmn.io/button.svg)](https://app.getpostman.com/run-collection/535e1aa2df75f17c63f0)

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
