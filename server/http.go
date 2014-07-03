package main

import (
	"github.com/golang/glog"
	"net/http"
	"net"
	"encoding/json"
	"time"
	"github.com/gysan/im/tunnel"
	"github.com/gysan/im/common"
	"code.google.com/p/goprotobuf/proto"
	"strconv"
	"io/ioutil"
)

func StartHttp() {
	glog.Info("Http start:")

	pushServeMux := http.NewServeMux()
	pushServeMux.HandleFunc("/push", Push)
	for _, bind := range Conf.WebsocketBind {
		glog.Infof("start http listen addr:\"%s\"", bind)
		go httpListen(pushServeMux, bind)
	}


	glog.Info("Http end.")
}

func Push(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	glog.Infof("%v", params)
	receiver := params.Get("receiver")
	messageid := params.Get("messageId")
	date := params.Get("date")
	res := map[string]interface{}{"result": 0}
	defer returnWrite(w, r, res)
	if receiver == "" {
		res["result"] = -1
		return
	}
	body := ""
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		glog.Errorf("ioutil.ReadAll() failed (%s)", err.Error())
		return
	}
	body = string(bodyBytes)
	glog.Infof("%v", body)

	c, err := UserChannel.Get(receiver, true)
	if err != nil {
		glog.Warningf("user_key:\"%s\" can't get a channel (%s)", receiver, err)
		return
	}

	ch := c.(*SeqChannel)

	glog.Infof("User channel: %v", ch.conn.Front())

	for e := ch.conn.Front(); e != nil; e = e.Next() {
		conn, _ := e.Value.(*Connection)
		glog.Infof("conn: %v", conn)
		i, err := strconv.ParseInt(date, 10, 64)
		if err != nil {
			panic(err)
		}
		normalMessage := &common.NormalMessage{
			MessageId: proto.String(messageid),
			Receiver: proto.String(receiver),
			Content: []byte(body),
			Date: proto.Int64(i),
		}
		conn.Conn.Write(tunnel.GetMessageBytes(0, int32(common.MessageCommand_NORMARL_MESSAGE), normalMessage))
	}
}

func httpListen(mux *http.ServeMux, bind string) {
	glog.Info("httpListen start:")
	server := &http.Server{Handler: mux, ReadTimeout: 30 * time.Second}
	l, err := net.Listen("tcp", bind)
	if err != nil {
		glog.Errorf("net.Listen('tcp', '%s') error(%v)", bind, err)
		panic(err)
	}
	if err := server.Serve(l); err != nil {
		glog.Errorf("server.Serve() error(%v)", err)
		panic(err)
	}
	glog.Info("httpListen end.")
}

func returnWrite(responseWrite http.ResponseWriter, request *http.Request, result map[string]interface{}) {
	data, err := json.Marshal(result)
	if err != nil {
		glog.Errorf("json.Marshal(\"%v\") error(%v)", result, err)
		return
	}

	if n, err := responseWrite.Write([]byte(string(data))); err != nil {
		glog.Errorf("w.Write(\"%s\") error(%v)", string(data), err)
	} else {
		glog.V(1).Infof("w.Write(\"%s\") write %d bytes", string(data), n)
	}
	glog.Infof("url: \"%s\", result:\"%s\", ip:\"%s\"", request.URL.String(), string(data), request.RemoteAddr)
}
