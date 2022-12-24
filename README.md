# timestampService
A JSON/HTTP service, in golang, that returns the matching timestamps of a periodic task.

Run instructions:
1) go run main.go <port>

2) To test the service open another terminal and run:
go run test.go <port> <period> <tz> <t1> <t2>

e.g.

go run test.go 8080 1h Europe/Athens 20210714T204603Z 20210715T123456Z

go run test.go 8080 2d Europe/Athens 20211010T204603Z 20211112T123456Z

go run test.go 8080 1mo Europe/Athens 20210214T204603Z 20211115T123456Z

go run test.go 8080 5y Europe/Athens 19850214T204603Z 20201115T123456Z

go run test.go 8080 1w Europe/Athens 20180214T204603Z 20211115T123456Z


OR
open a browser and put a URL like

http://localhost:8080/ptlist?period=1h&tz=Europe/Athens&t1=20210714T204603Z&t2=20210715T123456Z

http://localhost:8080/ptlist?period=2d&tz=Europe/Athens&t1=20211010T204603Z&t2=20211112T123456Z

http://localhost:8080/ptlist?period=1mo&tz=Europe/Athens&t1=20210214T204603Z&t2=20211115T123456Z

http://localhost:8080/ptlist?period=5y&tz=Europe/Athens&t1=19850214T204603Z&t2=20201115T123456Z

http://localhost:8080/ptlist?period=1w&tz=Europe/Athens&t1=20180214T204603Z&t2=20211115T123456Z
