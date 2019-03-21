package models

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/3devo/feconnector/utils"
)

// A logFile database model
//
// This is used to save a logFile in the boltdb
type LogFile struct {
	UUID      string `storm:"id" json:"uuid"`
	Name      string `json:"name"`
	FileName  string `json:"filename"`
	Timestamp int64  `json:"timestamp"`
	HasNote   bool   `json:"hasNote"`
}

// CreateLogFile creates a new logfile in the database
// It also creates a log file and a note file if available on the system
func CreateLogFile(uuid string, name string, note string, env *utils.Env) error {
	logFile := new(LogFile)
	logFile.UUID = uuid
	logFile.Name = name
	logFile.Timestamp = time.Now().Unix()
	if note != "" {
		logFile.HasNote = true
	}
	f, err := os.Create(filepath.Join(env.DataDir, "logs", logFile.GetFileName()))
	f.Close()
	if err != nil {
		return err
	}
	if logFile.HasNote {
		err := ioutil.WriteFile(filepath.Join(env.DataDir, "notes", logFile.GetFileName()), []byte(note), os.ModePerm)
		if err != nil {
			return err
		}
	}
	if err := env.Db.Save(logFile); err != nil {
		return err
	}
	return nil
}

// UpdateLogFile updates the name and notes
func (logFile *LogFile) UpdateLogFile(name string, note string, env *utils.Env) error {
	logFile.Name = name
	if note != "" {
		ioutil.WriteFile(filepath.Join(env.DataDir, "notes", logFile.GetFileName()), []byte(note), os.ModePerm)
	}
	err := env.Db.Update(logFile)
	if err != nil {
		return err
	}
	return nil
}

// AppendLog appends new information to an already existing log file.
func (logFile *LogFile) AppendLog(logData string, env *utils.Env) error {
	if logData != "" {
		f, err := os.OpenFile(filepath.Join(env.DataDir, "logs", logFile.GetFileName()), os.O_APPEND|os.O_WRONLY, os.ModePerm)
		if err == nil {
			f.WriteString(logData)
			f.Sync()
		}
		f.Close()
		return err
	}
	return nil
}

// DeleteLogFile
func (logFile *LogFile) DeleteLogFile(env *utils.Env) error {
	err := os.Remove(filepath.Join(env.DataDir, "logs", logFile.GetFileName()))

	if err != nil {
		return err
	}

	if logFile.HasNote {
		err = os.Remove(filepath.Join(env.DataDir, "notes", logFile.GetFileName()))
		if err != nil {
			return err
		}
	}
	err = env.Db.DeleteStruct(logFile)
	if err != nil {
		return err
	}
	return nil
}

// GetFileName returns the filename to use
func (logFile *LogFile) GetFileName() string {
	// Generate and store the filename on first use
	if logFile.FileName == "" {
		logFile.FileName := logFile.Name + "-" + time.Unix(logFile.Timestamp, 0).Format("2006-01-02-15-04-05") + ".txt"
	}
	return logFile.FileName
}
