package main

import (
	"github.com/golang/glog"
	"io"
	"net"
	"strconv"
	"time"
	"encoding/binary"
	"code.google.com/p/goprotobuf/proto"
	"github.com/gysan/im/common"
	"github.com/gysan/im/tunnel"
)

// StartTCP Start tcp listen.
func StartTCP() {
	for _, bind := range Conf.TCPBind {
		glog.Infof("start tcp listen addr:\"%s\"", bind)
		go tcpListen(bind)
	}
}

func tcpListen(bind string) {
	glog.V(1).Info("TCP listen start")
	addr, err := net.ResolveTCPAddr("tcp", bind)
	if err != nil {
		glog.Errorf("net.ResolveTCPAddr(\"tcp\"), %s) error(%v)", bind, err)
		panic(err)
	}
	glog.V(1).Infof("net.ResolveTCPAddr success: %v", addr)
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		glog.Errorf("net.ListenTCP(\"tcp4\", \"%s\") error(%v)", bind, err)
		panic(err)
	}
	glog.V(1).Infof("net.ListenTCP success: %v", l)
	// free the listener resource
	defer func() {
		glog.Infof("tcp addr: \"%s\" close", bind)
		if err := l.Close(); err != nil {
			glog.Errorf("listener.Close() error(%v)", err)
		}
	}()
	// init reader buffer instance
	for {
		glog.V(1).Info("start accept")
		conn, err := l.AcceptTCP()
		if err != nil {
			glog.Errorf("listener.AcceptTCP() error(%v)", err)
			continue
		}
		glog.V(1).Infof("AcceptTCP: %v", conn)
		if err = conn.SetKeepAlive(Conf.TCPKeepalive); err != nil {
			glog.Errorf("conn.SetKeepAlive() error(%v)", err)
			conn.Close()
			continue
		}
		if err = conn.SetReadBuffer(Conf.RcvbufSize); err != nil {
			glog.Errorf("conn.SetReadBuffer(%d) error(%v)", Conf.RcvbufSize, err)
			conn.Close()
			continue
		}
		if err = conn.SetWriteBuffer(Conf.SndbufSize); err != nil {
			glog.Errorf("conn.SetWriteBuffer(%d) error(%v)", Conf.SndbufSize, err)
			conn.Close()
			continue
		}
		// first packet must sent by client in specified seconds
		if err = conn.SetReadDeadline(time.Now().Add(time.Second*5)); err != nil {
			glog.Errorf("conn.SetReadDeadLine() error(%v)", err)
			conn.Close()
			continue
		}
		go handle(conn)
		glog.V(1).Info("accept finished")
	}
	glog.V(1).Info("TCP listen end.")
}

func handle(conn net.Conn) {
	head := make([]byte, 12)
	addr := conn.RemoteAddr().String()
	glog.V(1).Infof("<%s> handleTcpConn routine start", addr)
	conn.Read(head)
	glog.Infof("Head: %v", head)

	HeartbeatInitHandle(conn, head)

	// close the connection
	if err := conn.Close(); err != nil {
		glog.Errorf("<%s> conn.Close() error(%v)", addr, err)
	}
	glog.V(1).Infof("<%s> handleTcpConn routine stop", addr)
	return
}

func HeartbeatInitHandle(conn net.Conn, args []byte) {
	addr := conn.RemoteAddr().String()
	glog.Infof("addr: %v", addr)

	body := make([]byte, int32(binary.BigEndian.Uint32(args[8:12])))
	conn.Read(body)
	glog.V(1).Infof("Data: %v", body)

	heartbeatInit := &common.HeartbeatInit{}
	err := proto.Unmarshal(body, heartbeatInit)
	glog.Infof("Heartbeat init: %s", heartbeatInit.String())
	if err != nil {
		glog.Errorf("Unmarshal error: %s", err)
	}
	key := heartbeatInit.GetDeviceId()

	heartbeatStr := "30"
	i, err := strconv.Atoi(heartbeatStr)
	if err != nil {
		//		conn.Write(ParamReply)
		glog.Errorf("<%s> user_key:\"%s\" heartbeat:\"%s\" argument error (%s)", addr, key, heartbeatStr, err)
		return
	}
	if i < 30 {
		//		conn.Write(ParamReply)
		glog.Warningf("<%s> user_key:\"%s\" heartbeat argument error, less than %d", addr, key, 30)
		return
	}
	heartbeat := i + 5

	glog.Infof("<%s> subscribe to deviceId = %s, heartbeat = %d", addr, key, heartbeat)
	// fetch subscriber from the channel
	c, err := UserChannel.Get(key, true)
	glog.Infof("User channel: %v", c)
	if err != nil {
		glog.Warningf("<%s> user_key:\"%s\" can't get a channel (%s)", addr, key, err)
		//		conn.Write(ChannelReply)
		return
	}
	// add a conn to the channel
	connElem, err := c.AddConn(key, &Connection{Conn: conn})
	if err != nil {
		glog.Errorf("<%s> user_key:\"%s\" add conn error(%v)", addr, key, err)
		return
	}
	glog.V(1).Infof("connection elem: %v", connElem.Value)

	// blocking wait client heartbeat
	head := make([]byte, 12)
	begin := time.Now().UnixNano()
	end := begin + int64(time.Second)
	glog.V(1).Infof("Begin: %v, End: %v", begin, end)
	for {
		// more then 1 sec, reset the timer
		if end-begin >= int64(time.Second) {
			if err = conn.SetReadDeadline(time.Now().Add(time.Second*time.Duration(heartbeat))); err != nil {
				glog.Errorf("<%s> user_key:\"%s\" conn.SetReadDeadLine() error(%v)", addr, key, err)
				break
			}
			begin = end
		}
		glog.V(1).Infof("Begin: %v, End: %v", begin, end)
		if _, err = conn.Read(head); err != nil {
			if err != io.EOF {
				glog.Warningf("<%s> user_key:\"%s\" conn.Read() failed, read heartbeat timedout error(%v)", addr, key, err)
			} else {
				// client connection close
				glog.Warningf("<%s> user_key:\"%s\" client connection close error(%s)", addr, key, err)
			}
			break
		}

		glog.Infof("Head: %v", head)

		body := make([]byte, int32(binary.BigEndian.Uint32(head[8:12])))
		conn.Read(body)
		glog.V(1).Infof("Data: %v", body)

		glog.Infof("Case %d", int32(binary.BigEndian.Uint32(head[4:8])))
		switch int32(binary.BigEndian.Uint32(head[4:8])){
		case int32(common.MessageCommand_HEARTBEAT_INIT):
			hb := new(tunnel.HeartbeatInitTunnelLet)
			td := tunnel.TunnelData{Type: int32(common.MessageCommand_HEARTBEAT_INIT_RESPONSE)}
			hb.MessageReceived(conn, td)
		case int32(common.MessageCommand_HEARTBEAT):
			hb := new(tunnel.HeartbeatTunnelLet)
			td := tunnel.TunnelData{Type: int32(common.MessageCommand_HEARTBEAT_RESPONSE)}
			hb.MessageReceived(conn, td)
		case int32(common.MessageCommand_MESSAGE):
			MessageHandle(conn, body)
		}
		end = time.Now().UnixNano()
	}
	// remove exists conn
	if err := c.RemoveConn(key, connElem); err != nil {
		glog.Errorf("<%s> user_key:\"%s\" remove conn error(%s)", addr, key, err)
	}
	return
}


func MessageHandle(conn net.Conn, body []byte) {
	glog.V(1).Info("MessageHandle")
	addr := conn.RemoteAddr().String()
	glog.Infof("addr: %v", addr)
	// key, heartbeat

	message := &common.Message{}
	glog.V(1).Infof("body: %s", body)
	err := proto.Unmarshal(body, message)
	if err != nil {
		glog.Errorf("Unmarshal error: %s", err)
	}

	messageResponse := &common.MessageResponse {
		MessageId : proto.String("msg" + message.GetMessageId()),
		Status: proto.Bool(true),
	}
	conn.Write(tunnel.GetMessageBytes(0, int32(common.MessageCommand_MESSAGE_RESPONSE), messageResponse))

	receiver := message.GetReceiver()
	glog.V(1).Infof("Receiver: %s", receiver)

	c, err := UserChannel.Get(receiver, true)
	ch := c.(*SeqChannel)

	glog.Infof("User channel: %v", ch.conn.Front())
	if err != nil {
		glog.Warningf("<%s> user_key:\"%s\" can't get a channel (%s)", addr, receiver, err)
		//		conn.Write(ChannelReply)
		return
	}

	for e := ch.conn.Front(); e != nil; e = e.Next() {
		conn, _ := e.Value.(*Connection)
		glog.Infof("conn: %v", conn)
		conn.Conn.Write(tunnel.GetMessageBytes(0, int32(common.MessageCommand_RECEIVE_MESSAGE), message))
	}
	return
}
