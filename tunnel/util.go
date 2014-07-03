package tunnel

import (
	"encoding/binary"
	"code.google.com/p/goprotobuf/proto"
	"github.com/golang/glog"
	"sync"
)

var l sync.Mutex

func GetMessageBytes(tag int32, type1 int32, class proto.Message) []byte {
	l.Lock()
	var tag2 = make([]byte, 4)
	binary.BigEndian.PutUint32(tag2, uint32(tag))

	var type2 = make([]byte, 4)
	binary.BigEndian.PutUint32(type2, uint32(type1))

	transmitData, err := proto.Marshal(class)
	if err != nil {
		glog.Warningf("Marshaling error: ", err)
	}

	var length = make([]byte, 4)
	binary.BigEndian.PutUint32(length, uint32(proto.Size(class)))

	var sendByte = append(tag2, type2...)
	sendByte = append(sendByte, length...)
	sendByte = append(sendByte, []byte(transmitData)...)
	glog.Infof("Send data: %v %s", sendByte, class.String())
	l.Unlock()
	return sendByte
}
