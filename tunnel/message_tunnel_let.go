package tunnel

import (
	"github.com/golang/glog"
	"net"
)

type MessageTunnelLet struct {

}

func (messageTunnelLet *MessageTunnelLet) MessageReceived(connection net.Conn, data TunnelData) {
	glog.Info("MessageTunnelLet.MessageReceived")

//	addr := connection.RemoteAddr().String()
//
//	receiver := data.Object.GetReceiver()
//	glog.V(1).Infof("Receiver: %s", receiver)
//
//	c, err := UserChannel.Get(receiver, true)
//	ch := c.(*SeqChannel)
//
//	glog.Infof("User channel: %v", ch.conn.Front())
//	if err != nil {
//		glog.Warningf("<%s> user_key:\"%s\" can't get a channel (%s)", addr, receiver, err)
//		return
//	}
//
//	messageResponse := &common.MessageResponse {
//		MessageId : proto.String("msg" + data.Object.GetMessageId()),
//		Status: proto.Bool(true),
//	}
//	sendByte := getMessageBytes(data.Tag, data.Type, messageResponse)
//
//	for e := ch.conn.Front(); e != nil; e = e.Next() {
//		conn, _ := e.Value.(*Connection)
//		glog.Infof("conn: %v", conn)
//		conn.Conn.Write(sendByte)
//	}
//	return
}
