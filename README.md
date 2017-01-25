# bf-ia-broker

A broker for image archives in support of Beachfront. This component generally stands between a UI (e.g., bf-ui) and one or more imagery providers (e.g., Planet Labs).

### Building
This component is built in Go. To build it, clone the repository in your Go source directory and issue the following command:
```
go install
```

### Environment Variables

|Variable|Description|Default|
|---------|-----------|------|
|BF_TIDE_PREDICTION_URL|Location of the tide prediction service|https://bf-tideprediction.int.geointservices.io/tides |
|PL_API_URL|Location of Planet Labs API|https://api.planet.com/ |
|PL_API_KEY|Planet Labs API Key|N/A|

### Running
To run as a web server, use the following command:
```
bf-ia-broker serve
```
Additional command line options may be available at a later time.

### Using
In [handlers.go](planet/handlers.go) there are some REST handlers.

|Endpoint|Command|Description|
|-------|--------|------------|
|/planet/discover/{itemType}|GET|Discover (search), as a GeoJSON feature collection|
|/planet/{itemType}/{id}|GET|Metadata for an ID, as a GeoJSON feature|
|/planet/activate/{itemType}/{id}|POST|Activate a resource|

See the Swagger docs or the source for details on using those handlers.