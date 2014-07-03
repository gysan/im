package main

import (
	"flag"
	"github.com/Terry-Mao/goconf"
	"github.com/golang/glog"
	"runtime"
)

var (
	Conf     *Config
	confFile string
)

func init() {
	flag.StringVar(&confFile, "c", "./comet.conf", " set gopush-cluster comet config file path")
}

type Config struct {
	// base
	TCPBind       []string `goconf:"base:tcp.bind:,"`
	WebsocketBind []string `goconf:"base:websocket.bind:,"`
	// channel
	SndbufSize              int           `goconf:"channel:sndbuf.size:memory"`
	RcvbufSize              int           `goconf:"channel:rcvbuf.size:memory"`
	BufioInstance           int           `goconf:"channel:bufio.instance"`
	BufioNum                int           `goconf:"channel:bufio.num"`
	TCPKeepalive            bool          `goconf:"channel:tcp.keepalive"`
	MaxSubscriberPerChannel int           `goconf:"channel:maxsubscriber"`
	ChannelBucket           int           `goconf:"channel:bucket"`
	MsgBufNum               int           `goconf:"channel:msgbuf.num"`
}

// InitConfig get a new Config struct.
func InitConfig() error {
	Conf = &Config{
		// base
		WebsocketBind: []string{"localhost:6968"},
		TCPBind:       []string{"localhost:6969"},
		// channel
		SndbufSize:              2048,
		RcvbufSize:              256,
		BufioInstance:           runtime.NumCPU(),
		BufioNum:                128,
		TCPKeepalive:            false,
		MaxSubscriberPerChannel: 64,
		ChannelBucket:           runtime.NumCPU(),
		MsgBufNum:               30,
	}
	c := goconf.New()
	if err := c.Parse(confFile); err != nil {
		glog.Errorf("goconf.Parse(\"%s\") error(%v)", confFile, err)
		return err
	}
	if err := c.Unmarshal(Conf); err != nil {
		glog.Errorf("goconf.Unmarshall() error(%v)", err)
		return err
	}
	return nil
}
