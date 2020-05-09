package main

import (
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/toolkits/pkg/file"
	"github.com/toolkits/pkg/runner"
	"main/config"
	"main/redisc"
	"main/util"
	"main/util/logger"
	"os"
	"os/signal"
	"path"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var (
	conf      *string
	content   string
	vers      *bool
	help      *bool
	test      *string
	urlStr    string
	semaphore chan int
	sidEtime  map[int64]int64
)

func init() {
	vers = flag.Bool("v", false, "display the version.")
	help = flag.Bool("h", false, "print this help.")
	conf = flag.String("f", "", "specify configuration file.")
	test = flag.String("t", "", "test smtp configuration.")
	flag.Parse()

	if *vers {
		fmt.Println("version:", config.Version)
		os.Exit(0)
	}

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	runner.Init()
	fmt.Println("runner.cwd:", runner.Cwd)
	fmt.Println("runner.hostname:", runner.Hostname)

	sidEtime=make(map[int64]int64)
}

func main() {
	aconf()
	pconf()
	cfg := config.Get()
	redisc.InitRedis()
	urlStr = initurl(cfg)

	gosendMessages()

	ending()
}

//自动添加配置文件
func aconf() {
	if *conf != "" && file.IsExist(*conf) {
		return
	}
	dir, err := os.Getwd()
	if err != nil {
		log.WithError(err).Info("获取当前工作目录失败")
		return
	}
	*conf = path.Join(dir, "config", "message-sender.local.yml")
	if file.IsExist(*conf) {
		return
	}

	*conf = path.Join(dir, "config", "message-sender.yml")
	if file.IsExist(*conf) {
		return
	}

	fmt.Println("no configuration file for sender")
	os.Exit(1)
}

//读取配置文件
func pconf() {
	if err := config.ParseConfig(*conf); err != nil {
		log.WithFields(logger.Weblog).Errorln("配置文件读取失败")
		fmt.Println("cannot parse configuration file:", err)
		os.Exit(1)
	} else {
		fmt.Println("parse configuration file:", *conf)
	}
}

//拼接url
func initurl(cfg config.Config) string {
	ip := cfg.Url.Ip
	port := cfg.Url.Port
	corpid := cfg.Url.Corpid
	pwd := cfg.Url.Pwd
	urlStr = "http://" + ip + ":" + port + "/" + "ws/BatchSend2.aspx?CorpID=" + corpid + "&" + "Pwd=" + pwd
	return urlStr
}

func gosendMessages() {
	cfg := config.Get()
	//如果发送短信的并发太大，怕务器受不了
	semaphore = make(chan int, cfg.Consumer.Worker)
	for {
		messages := redisc.Pop(1, cfg.Consumer.Queue)
		if len(messages) == 0 {
			time.Sleep(time.Duration(300) * time.Millisecond)
			continue
		}
		go sendMessages(messages)
	}
}

func sendMessages(messages []*config.Message) {
	for _, message := range messages {
		semaphore <- 1
		go sendMessage(message)
	}
}
func sendMessage(message *config.Message) {
	defer func() {
		<-semaphore
	}()
	postParam := make(map[string]string)
	mobile := "&Mobile="
	for _, v := range message.Tos {
		mobile += v + ","
	}
	mobile = strings.TrimRight(mobile, ",")
	urlStr += mobile
	postParam["Content"] = genContent(message,sidEtime)
	post, err := util.DoHttpPost(urlStr, postParam)
	if err != nil {
		log.WithFields(logger.Weblog).Errorln("post请求失败：", err)
	}
	fmt.Println(post)
}

var ET = map[string]string{
	"alert":    "告警",
	"recovery": "恢复",
}

func parseEtime(etime int64) string {
	t := time.Unix(etime, 0)
	return t.Format("2006-01-02 15:04:05")
}

func genContent(message *config.Message,sidEtime map[int64]int64) string {
	if message.Event == nil {
		return ""
	}
	sname := message.Event.Sname

	endpointAlias := message.Event.EndpointAlias

	eventType := message.Event.EventType

	sid := message.Event.Sid

	if strings.Contains(eventType, "recov") {
		delete(sidEtime, sid)
		content := endpointAlias + sname + ",已经恢复," + "当前时间：" + time.Now().Format("2006-01-02 15:04:05")
		return content
	}

	if _, ok := sidEtime[sid]; ok != true {
		sidEtime[sid] = message.Event.Etime
	}

	content := endpointAlias + sname + ",已经异常" + strconv.FormatInt(int64(time.Now().Sub(time.Unix(sidEtime[sid], 0))/60000000000), 10) +
		"分钟" + ",异常时间:" + parseEtime(sidEtime[sid])
	return content
}
func ending() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	select {
	case <-c:
		fmt.Printf("stop signal caught, stopping... pid=%d\n", os.Getpid())
	}
	redisc.CloseRedis()
	fmt.Println("sender stopped successfully")
}
