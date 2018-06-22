package models

// A logFile database model
//
// This is used to save a logFile in the boltdb
type LogFile struct {
	ID        int    `storm:"id,increment"`
	Name      string `json:"name"`
	Timestamp int64  `json:"timestamp"`
	HasNote   bool   `json:"hasNote"`
}

func NewLogFile(name string, timestamp int64, hashNote bool) *LogFile {
	logFile := new(LogFile)
	logFile.Name = name
	logFile.Timestamp = timestamp
	logFile.HasNote = hashNote
	return logFile
}
