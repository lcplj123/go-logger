# go-logger
用法：
import "github.com/lcplj123/go-logger/logger"
log := logger.NewLogger()
log.SetLevel(logger.DEBUG) //设置输出级别，默认是ERROR
log.SetConsole(false)  //设置是否是控制台输出，默认是true

log.SetRollFile(...)  //设置文件轮转
log.SetRollDate(...)  //设置日期轮转

日志轮转和日期轮转只能设置一个。

