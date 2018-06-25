package models

import (
	"io/ioutil"
	"os"
	"time"

	"github.com/3devo/feconnector/utils"
)

// A logFile database model
//
// This is used to save a logFile in the boltdb
type LogFile struct {
	UUID      string `storm:"id" json:"uuid"`
	Name      string `json:"name"`
	Timestamp int64  `json:"timestamp"`
	HasNote   bool   `json:"hasNote"`
}

// CreateLogFile creates a new logfile in the database
// It also creates a log file and a note file if available on the system
func CreateLogFile(uuid string, name string, note string) (*LogFile, error) {
	logFile := new(LogFile)
	logFile.UUID = uuid
	logFile.Name = name
	logFile.Timestamp = time.Now().Unix()
	if note != "" {
		logFile.HasNote = true
	}
	logName := logFile.Name + "-" + time.Unix(logFile.Timestamp, 0).Format("2006-01-02-15-04-05") + ".txt"
	f, err := os.Create("./logs/" + logName)
	f.Close()
	if err != nil {
		return nil, err
	}
	if logFile.HasNote {
		err := ioutil.WriteFile("./notes/"+logName, []byte(note), os.ModePerm)
		if err != nil {
			return nil, err
		}
	}
	return logFile, nil
}

// UpdateLogFile updates the name and notes
func (logFile *LogFile) UpdateLogFile(name string, note string) {
	logName := logFile.Name + "-" + time.Unix(logFile.Timestamp, 0).Format("2006-01-02-15-04-05") + ".txt"
	logFile.Name = name
	if note != "" {
		ioutil.WriteFile("./notes/"+logName, []byte(note), os.ModePerm)
	}
}

// AppendLog appends new information to an already existing log file.
func (logFile *LogFile) AppendLog(log string) {
	logName := logFile.Name + "-" + time.Unix(logFile.Timestamp, 0).Format("2006-01-02-15-04-05") + ".txt"

	if log != "" {
		f, err := os.OpenFile("./logs/"+logName, os.O_APPEND, os.ModePerm)
		if err == nil {
			f.WriteString(log)
		}
		f.Close()
	}
}

// DeleteLogFile
func (logFile *LogFile) DeleteLogFile(env *utils.Env) error {
	logName := logFile.Name + "-" + time.Unix(logFile.Timestamp, 0).Format("2006-01-02-15-04-05") + ".txt"
	err := os.Remove("./logs/" + logName)

	if err != nil {
		return err
	}

	err = os.Remove("./notes/" + logName)
	if err != nil {
		return err
	}
	env.Db.DeleteStruct(logFile)
	return nil
}
