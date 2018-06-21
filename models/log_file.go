package models

type LogFile struct {
	ID        int `storm:"id,increment"`
	Name      string
	Timestamp int64
	HasNote   bool
}

func NewLogFile(name string, timestamp int64, hashNote bool) *LogFile {
	logFile := new(LogFile)
	logFile.Name = name
	logFile.Timestamp = timestamp
	logFile.HasNote = hashNote
	return logFile
}
