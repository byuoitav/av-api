# av-api
[![CircleCI](https://img.shields.io/circleci/project/byuoitav/av-api.svg)](https://circleci.com/gh/byuoitav/av-api) [![Apache 2 License](https://img.shields.io/hexpm/l/plug.svg)](https://raw.githubusercontent.com/byuoitav/av-api/master/LICENSE)

[![Run in Postman](https://run.pstmn.io/button.svg)](https://app.getpostman.com/run-collection/dd1b2c873b3eff5a4ca7) [![View in Swagger](http://jessemillar.github.io/view-in-swagger-button/button.svg)](http://byuoitav.github.io/swagger-ui/?url=https://raw.githubusercontent.com/byuoitav/av-api/master/swagger.json)

## Setup
`CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS` needs to be set to enable communication with the [configuration-database-microservice](https://github.com/byuoitav/common). The `EMS_API_USERNAME` and `EMS_API_PASSWORD` environment variables need to be set in order to retrieve room availability data from the [Event Management System](https://emsweb.byu.edu/VirtualEMS/BrowseForSpace.aspx).

## Example Usage
Perform a PUT on `http://localhost:8000/buildings/ITB/rooms/1001D` with the following body:
```
{
	"currentVideoInput": "AppleTV",
	"displays": [{
		"name": "D1",
		"power": "on",
		"blanked": false
	}],
	"audioDevices": [{
		"volume": 10
	}]
}
```

## Docker Development
For Docker development via `docker-compose` utilize the following commands depending on your use case:

### Development Testing (Build Containers Locally)
```
docker-compose -f docker-compose-dev.yml up
```

### Production Testing (Pull Containers)
```
docker-compose -f docker-compose-prod.yml up
```
