# tempserver
Simple web + rest server to provide data from templogger: https://github.com/perolo/templogger

## How to use
Reads a properties file tempserver.properties, override with --prop filename.properties

* dbfile - string: Name and path to sqlite file database
* port - string: Port number

## Build
`
go build tempserver.go
`
## Start
`
nohup ./tempserver &
`
## API:s

### http://localhost:8081/

### http://localhost:8081/start/{id}
Retrieve the 50 items in database starting from {id}

### http://localhost:8081/last
Retrieve the last 10 items in database

### http://localhost:8081/quit
Terminate server


## Using TempToExcel
Provides a way to retrieve data through rest API and present data in Excelsheet:
https://github.com/perolo/temptoexcel

Gui graphs very rudimetal - work remains...


