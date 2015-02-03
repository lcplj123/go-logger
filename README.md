# go-logger
用法：
import "github.com/lcplj123/go-logger/logger"
log := logger.NewLogger()
log.SetLevel(logger.DEBUG) //设置输出级别，默认是ERROR
log.SetConsole(false)  //设置是否是控制台输出，默认是true

//设置文件大小轮转
参数一： 日志的目录
参数二： 日志文件名
参数三： 日志的大小（其单位用最后一个参数标示出来）
参数四： 最大日志文件个数

log.SetRollFile(`./logs`,`mylog.log`,15,10,logger.MB)  //设置文件大小轮转

//日志文件日期轮转
参数一： 日志的目录
参数二： 日志文件名
参数三：每隔几天轮转 目前此参数无效，默认设置为1即可。
log.SetRollDate(`./logs`,`mylog.log`,1)  //设置日期轮转

日志轮转和日期轮转只能设置一个。如果设置两个，以第一次设置为准。

可以通过NewLogger函数多次生成log对象，针对不同的日志分类输出到不通文件。

