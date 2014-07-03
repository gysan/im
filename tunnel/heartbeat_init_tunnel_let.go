package tunnel

import (
	"github.com/golang/glog"
	"net"
	"github.com/gysan/im/common"
)

type HeartbeatInitTunnelLet struct {

}

func (heartbeatInit *HeartbeatInitTunnelLet) MessageReceived(connection net.Conn, data TunnelData) {
	glog.V(1).Info("HeartbeatInitTunnelLet.MessageReceived")
	//	addr := connection.RemoteAddr().String()

	heartbeatInitResponse := &common.HeartbeatInitResponse{}
	sendByte := GetMessageBytes(data.Tag, data.Type, heartbeatInitResponse)
	//	glog.V(1).Infof("Reply: %v", reply)
	if _, err := connection.Write(sendByte); err != nil {
		//		glog.Errorf("<%s> user_key:\"%s\" conn.Write() failed, write heartbeat to client error(%v)", addr, key, err)
		return
	}
	//	glog.V(1).Infof("<%s> user_key:\"%s\" receive heartbeat (%s)", addr, key, reply)
	return
}
