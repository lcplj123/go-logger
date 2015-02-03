package logger

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"
)

const DATEFORMAT string = "2006-01-02"

//版本信息
const (
	VERSION string = "1.0.0"
)

//日志的level定义
const (
	ALL int = iota
	DEBUG
	INFO
	WARN
	ERROR
	FATAL
	OFF
)

//流量单位定义
const (
	_        = iota
	KB int64 = 1 << (iota * 10)
	MB
	GB
	TB
)

//日志轮转的方式
const (
	_NULL int = iota //不论转日志
	_DATE            //按日期轮转
	_FILE            //按文件大小轮转
)

//logger的结构定义(日志对象的定义)
type Logger struct { //包级私有
	level        int     //日志的等级
	maxFileSize  int64   //日志的最大文件
	maxFileCount uint    //日志文件的最大个数
	rollDay      uint    //每隔多少天轮转日志
	rollWay      int     //日志轮转的方式
	console      bool    //是否控制台输出
	logObj       *__FILE //日志输出文件对象
}

//文件对象定义
type __FILE struct {
	dir     string        //文件目录
	fname   string        //文件名字
	suffix  int           //文件后缀,用数字
	iscover bool          //是否覆盖
	date    *time.Time    //时间
	mu      *sync.RWMutex //锁
	logf    *os.File      //文件句柄
	lg      *log.Logger
}

//logger的接口信息
type ILogger interface {
	SetConsole(b bool)                                                        //设置是否控制台输出
	SetLevel(l int)                                                           //设置输出级别
	SetRollFile(dir, name string, maxfilesize, maxfilecount uint, unit int64) //按照文件大小轮转日志
	SetRollDate(dir, name string)                                             //按照日期轮转日志
	Console(s ...interface{})                                                 //输出到控制台
	Debug(s ...interface{})                                                   //debug输出
	Info(s ...interface{})                                                    //info输出
	Warn(s ...interface{})                                                    //Warn输出
	Error(s ...interface{})                                                   //error输出
	Fatal(s ...interface{})                                                   //fatal输出
	catchError()                                                              //捕获错误
}

//生成新的log对象
//默认不论转日志，控制台输出
func NewLogger() *Logger {
	return &Logger{
		level:   ERROR,
		rollWay: _NULL,
		console: true,
		logObj:  nil,
	}
}

//实现ILogger接口
func (this *Logger) SetConsole(b bool) {
	this.console = b
}

func (this *Logger) SetLevel(level int) {
	if level < ALL || level > OFF {
		panic("logger: set level error!")
		return
	}
	this.level = level
}
func (this *Logger) SetRollFile(dir, name string, maxfilesize, maxfilecount uint, unit int64) {
	if dir == "" || name == "" || maxfilesize == 0 || maxfilecount == 0 {
		panic("Logger: SetRollFile error!")
		return
	}
	if this.rollWay != _NULL {
		return
	}
	this.maxFileCount = maxfilecount
	this.maxFileSize = int64(maxfilesize) * unit
	this.rollWay = _FILE
	this.logObj = &__FILE{
		dir:     dir,
		fname:   name,
		iscover: false,
		mu:      new(sync.RWMutex),
	}
	this.logObj.mu.Lock()
	defer this.logObj.mu.Unlock()
	if !isExist(dir) {
		os.Mkdir(dir, os.ModeDir|os.ModePerm)
	}
	if !this.isMustRename() {
		this.logObj.logf, _ = os.OpenFile(dir+"/"+name, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0)
		this.logObj.lg = log.New(this.logObj.logf, "\n", log.Ldate|log.Ltime|log.Lshortfile)
	} else {
		this.rename()
	}
	//监测文件
	go this.fileMonitor()
}

//目前限定只能是每隔一天
func (this *Logger) SetRollDate(dir, name string, interval uint) {
	if interval == 0 || dir == "" || name == "" {
		panic("Logger: SetRollData error!")
		return
	}
	if this.rollWay != _NULL {
		return
	}
	this.rollWay = _DATE
	this.rollDay = interval
	t, _ := time.Parse(DATEFORMAT, time.Now().Add(24*time.Hour).Format(DATEFORMAT))
	this.logObj = &__FILE{
		dir:     dir,
		fname:   name,
		iscover: false,
		date:    &t,
		mu:      new(sync.RWMutex),
	}
	this.logObj.mu.Lock()
	defer this.logObj.mu.Unlock()
	if !isExist(dir) {
		os.Mkdir(dir, os.ModeDir|os.ModePerm)
	}
	if !this.isMustRename() {
		this.logObj.logf, _ = os.OpenFile(dir+"/"+name, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0)
		this.logObj.lg = log.New(this.logObj.logf, "\n", log.Ldate|log.Ltime|log.Lshortfile)
	} else {
		this.rename()
	}
	//监测文件
	go this.fileMonitor()
}

func (this *Logger) Console(s ...interface{}) {
	if this.console {
		_, file, line, _ := runtime.Caller(2)
		short := file
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				short = file[i+1:]
				break
			}
		}
		file = short
		log.Println(file+":"+strconv.Itoa(line), s)
	}
}

func (this *Logger) Debug(s ...interface{}) {
	if this.level > DEBUG {
		return
	}
	defer catchError()
	this.logObj.mu.Lock()
	defer this.logObj.mu.Unlock()
	this.logObj.lg.Output(2, fmt.Sprintln("debug:", s))
	this.Console("debug:", s)
}

func (this *Logger) Info(s ...interface{}) {
	if this.level > INFO {
		return
	}
	defer catchError()
	this.logObj.mu.Lock()
	defer this.logObj.mu.Unlock()
	this.logObj.lg.Output(2, fmt.Sprintln("info:", s))
	this.Console("info:", s)
}

func (this *Logger) Warn(s ...interface{}) {
	if this.level > WARN {
		return
	}
	defer catchError()
	this.logObj.mu.Lock()
	defer this.logObj.mu.Unlock()
	this.logObj.lg.Output(2, fmt.Sprintln("warn:", s))
	this.Console("warn:", s)
}

func (this *Logger) Error(s ...interface{}) {
	if this.level > ERROR {
		return
	}
	defer catchError()
	this.logObj.mu.Lock()
	defer this.logObj.mu.Unlock()
	this.logObj.lg.Output(2, fmt.Sprintln("error:", s))
	this.Console("error:", s)
}

func (this *Logger) Fatal(s ...interface{}) {
	if this.level > FATAL {
		return
	}
	defer catchError()
	this.logObj.mu.Lock()
	defer this.logObj.mu.Unlock()
	this.logObj.lg.Output(2, fmt.Sprintln("fatal:", s))
	this.Console("fatal:", s)
}

func isExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

func (this *Logger) isMustRename() bool {
	if this.rollWay == _DATE {
		t, _ := time.Parse(DATEFORMAT, time.Now().Format(DATEFORMAT))
		if t.After(*this.logObj.date) {
			return true
		}

	} else if this.rollWay == _FILE {
		if getFileSize(this.logObj.dir+"/"+this.logObj.fname) >= this.maxFileSize {
			return true
		}
	}
	return false
}

func (this *Logger) rename() {
	if this.rollWay == _DATE {
		newname := this.logObj.dir + "/" + this.logObj.fname + "." + this.logObj.date.Add(-24*time.Hour).Format(DATEFORMAT)
		if !isExist(newname) { //如果新生成的日志不存在
			if this.logObj.logf != nil {
				this.logObj.logf.Sync()
				this.logObj.logf.Close()
			}
			if this.maxFileCount == 1 {
				err := os.Remove(this.logObj.dir + "/" + this.logObj.fname)
				if err != nil {
					this.logObj.lg.Println("Logger: rename error", err.Error())
				}
			} else {
				err := os.Rename(this.logObj.dir+"/"+this.logObj.fname, newname)
				if err != nil {
					this.logObj.lg.Println("Logger: rename error", err.Error())
				}
			}

			t, _ := time.Parse(DATEFORMAT, time.Now().Format(DATEFORMAT))
			this.logObj.date = &t
			this.logObj.logf, _ = os.Create(this.logObj.dir + "/" + this.logObj.fname)
			this.logObj.lg = log.New(this.logObj.logf, "\n", log.Ldate|log.Ltime|log.Lshortfile)
		} else {
		}
	} else if this.rollWay == _FILE {
		if this.maxFileCount == 1 {
			if this.logObj.logf != nil {
				this.logObj.logf.Sync()
				this.logObj.logf.Close()
				os.Remove(this.logObj.dir + "/" + this.logObj.fname)
			}
		} else {
			for i := this.maxFileCount; i >= 1; i-- {
				oldname := this.logObj.dir + "/" + this.logObj.fname + "." + strconv.Itoa(int(i))
				if i == this.maxFileCount {
					if isExist(oldname) {
						os.Remove(oldname)
					}
				} else {
					if isExist(oldname) {
						os.Rename(oldname, this.logObj.dir+"/"+this.logObj.fname+"."+strconv.Itoa(int(i+1)))
					}
				}
			}
			this.logObj.logf.Sync()
			this.logObj.logf.Close()
			os.Rename(this.logObj.dir+"/"+this.logObj.fname, this.logObj.dir+"/"+this.logObj.fname+"."+strconv.Itoa(1))
		}
		this.logObj.logf, _ = os.Create(this.logObj.dir + "/" + this.logObj.fname)
		this.logObj.lg = log.New(this.logObj.logf, "\n", log.Ldate|log.Ltime|log.Lshortfile)
	}
}

func (this *Logger) fileMonitor() {
	timer := time.NewTicker(2 * time.Second)
	for {
		select {
		case <-timer.C:
			this.fileCheck()
		}
	}
}

func (this *Logger) fileCheck() {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
		}
	}()
	if this.logObj != nil && this.isMustRename() {
		this.logObj.mu.Lock()
		defer this.logObj.mu.Unlock()
		this.rename()
	}

}
func getFileSize(path string) int64 {
	f, e := os.Stat(path)
	if e != nil {
		return 0
	}
	return f.Size()
}

func catchError() {
	if err := recover(); err != nil {
		//log.Println("catchError:", err)
		fmt.Println("catchError:", err)
	}
}
