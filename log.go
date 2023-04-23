package gobatis

import (
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"sync"
)

var logs *logrus.Logger

// Log 自定义Log需要实现的借口
type Log interface {
	Info(...interface{})
	Error(...interface{})
	Debug(...interface{})
	Panic(...interface{})
	Warn(...interface{})
}

func init() {
	logs = logrus.New()
	f := &Formatter{ProjectName: "GoBatis"}
	f.Buf = &sync.Pool{New: func() any {
		return new(bytes.Buffer)
	}}
	logs.SetFormatter(f)
	logs.Out = os.Stdout
	logs.Level = logrus.DebugLevel
}

func Level(level logrus.Level) {
	logs.SetLevel(level)
}

type Formatter struct {
	ProjectName string
	Buf         *sync.Pool
	*logrus.TextFormatter
}

func Info(msg ...any) {
	logs.Infoln(msg...)
}

func Debug(msg ...any) {
	logs.Debugln(msg...)
}

func Error(err ...any) {
	logs.Errorln(err...)
}

func Panic(v ...any) {
	logs.Panicln(v...)
}

func (format *Formatter) Format(entry *logrus.Entry) ([]byte, error) {
	buf := format.Buf.Get().(*bytes.Buffer)
	defer format.Buf.Put(buf)
	defer buf.Reset()
	var f = ""
	t := entry.Time.Format("2006-01-02 15:04:05")
	if entry.Data != nil && len(entry.Data) > 0 {
		f = fmt.Sprintf("[%s] %s [%s] [%s] -> %s\n", format.ProjectName, t, entry.Level, entry.Data["type"], entry.Message)
	} else {
		f = fmt.Sprintf("[%s] %s [%s]  -> %s\n", format.ProjectName, t, entry.Level, entry.Message)
	}
	buf.WriteString(f)
	return buf.Bytes(), nil
}
