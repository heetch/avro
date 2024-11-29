module dummy

go 1.23

require (
	github.com/google/uuid v1.6.0
	github.com/heetch/avro v0.4.6-0.20241128170218-20b562ce498f
	github.com/heetch/galaxy-go/v2 v2.24.0
)

require github.com/actgardner/gogen-avro/v10 v10.2.1 // indirect

replace (
	github.com/heetch/avro => ../../../..
)
