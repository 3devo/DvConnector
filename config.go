package main

const (
	DevoUsbPID = "0C5B"
	DevoUsbVID = "16D0"
)

type Column struct {
	Validator string
}

// ApplicationConfiguration is the configuration that contains application specific logic like.
// The incoming header format
var KnownColumns = map[string]Column{
	"Time": Column{
		Validator: `-?[0-9]\d*(\.\d+)?`,
	},
	"SetT1": Column{
		Validator: `-?[0-9]\d*(\.\d+)?`,
	},
	"Temp1": Column{
		Validator: `-?[0-9]\d*(\.\d+)?`,
	},
	"dc1": Column{
		Validator: `-?[0-9]\d*(\.\d+)?`,
	},
	"Err1": Column{
		Validator: `-?[0-9]\d*(\.\d+)?`,
	},
	"SetT2": Column{
		Validator: `-?[0-9]\d*(\.\d+)?`,
	},
	"Temp2": Column{
		Validator: `-?[0-9]\d*(\.\d+)?`,
	},
	"dc2": Column{
		Validator: `-?[0-9]\d*(\.\d+)?`,
	},
	"Err2": Column{
		Validator: `-?[0-9]\d*(\.\d+)?`,
	},
	"SetT3": Column{
		Validator: `-?[0-9]\d*(\.\d+)?`,
	},
	"Temp3": Column{
		Validator: `-?[0-9]\d*(\.\d+)?`,
	},
	"dc3": Column{
		Validator: `-?[0-9]\d*(\.\d+)?`,
	},
	"Err3": Column{
		Validator: `-?[0-9]\d*(\.\d+)?`,
	},
	"SetT4": Column{
		Validator: `-?[0-9]\d*(\.\d+)?`,
	},
	"Temp4": Column{
		Validator: `-?[0-9]\d*(\.\d+)?`,
	},
	"dc4": Column{
		Validator: `-?[0-9]\d*(\.\d+)?`,
	},
	"Err4": Column{
		Validator: `-?[0-9]\d*(\.\d+)?`,
	},
	"intT4": Column{
		Validator: `-?[0-9]\d*(\.\d+)?`,
	},
	"ExtCur": Column{
		Validator: `-?[0-9]\d*(\.\d+)?`,
	},
	"ExtPWM": Column{
		Validator: `-?[0-9]\d*(\.\d+)?`,
	},
	"ExtTmp": Column{
		Validator: `-?[0-9]\d*(\.\d+)?`,
	},
	"Overht": Column{
		Validator: `-?[0-9]\d*(\.\d+)?`,
	},
	"FAULT": Column{
		Validator: `-?[0-9]\d*(\.\d+)?`,
	},
	"SetRPM": Column{
		Validator: `-?[0-9]\d*(\.\d+)?`,
	},
	"RPM": Column{
		Validator: `-?[0-9]\d*(\.\d+)?`,
	},
	"FT": Column{
		Validator: `-?[0-9]\d*(\.\d+)?`,
	},
	"FTAVG": Column{
		Validator: `-?[0-9]\d*(\.\d+)?`,
	},
	"Puller": Column{
		Validator: `-?[0-9]\d*(\.\d+)?`,
	},
	"MemFree": Column{
		Validator: `-?[0-9]\d*(\.\d+)?`,
	},
	"Status": Column{
		Validator: `[a-zA-Z]+`,
	},
	"WndrSpd": Column{
		Validator: `-?[0-9]\d*(\.\d+)?`,
	},
	"PosSpd": Column{
		Validator: `-?[0-9]\d*(\.\d+)?`,
	},
	"Length": Column{
		Validator: `-?[0-9]\d*(\.\d+)?`,
	},
	"Volume": Column{
		Validator: `-?[0-9]\d*(\.\d+)?`,
	},
	"SpDia": Column{
		Validator: `-?[0-9]\d*(\.\d+)?`,
	},
	"SpFill": Column{
		Validator: `-?[0-9]\d*(\.\d+)?`,
	},
}
