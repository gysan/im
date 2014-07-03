package main

import (
	"flag"
	"github.com/golang/glog"
	"runtime"
	"time"
)

func main() {
	flag.Parse()

	glog.Info("IM server start.")
	defer glog.Flush()

	if err := InitConfig(); err != nil {
		glog.Errorf("InitConfig() error(%v)", err)
	}

	runtime.GOMAXPROCS(runtime.NumCPU())

	glog.Info("Channel list start.")
	UserChannel = NewChannelList()
	defer UserChannel.Close()
	glog.Info("Channel list end.")

	glog.Info("Comet start.")
	StartTCP()
	glog.Info("Comet end.")

	time.Sleep(time.Second)

	StartHttp()

	glog.Info("Signal start.")
	signalCH := InitSignal()
	HandleSignal(signalCH)
	glog.Info("Signal end.")

	glog.Info("IM server end.")
}
