package tunnel

import (
	"github.com/golang/glog"
	"github.com/gysan/im/common"
	"net"
)

type HeartbeatTunnelLet struct {

}

func (heartbeat *HeartbeatTunnelLet) MessageReceived(connection net.Conn, data TunnelData) {
	glog.V(1).Info("HeartbeatTunnelLet.MessageReceived")

	heartbeatResponse := &common.HeartbeatResponse{}
	sendByte := GetMessageBytes(data.Tag, data.Type, heartbeatResponse)
	//	glog.V(1).Infof("Reply: %v", reply)
	if _, err := connection.Write(sendByte); err != nil {
		//		glog.Errorf("<%s> user_key:\"%s\" conn.Write() failed, write heartbeat to client error(%v)", addr, key, err)
		return
	}
	//	glog.V(1).Infof("<%s> user_key:\"%s\" receive heartbeat (%s)", addr, key, reply)
	return
}
