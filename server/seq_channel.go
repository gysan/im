package main

import (
	"errors"
	"github.com/Terry-Mao/gopush-cluster/hlist"
	"github.com/golang/glog"
	"sync"
	"github.com/gysan/im/common"
	"github.com/gysan/im/tunnel"
)

var (
	ErrMessageSave   = errors.New("Message set failed")
	ErrMessageGet    = errors.New("Message get failed")
	ErrMessageRPC    = errors.New("Message RPC not init")
	ErrAssectionConn = errors.New("Assection type Connection failed")
)

// Sequence Channel struct.
type SeqChannel struct {
	// Mutex
	mutex *sync.Mutex
	// client conn double linked-list
	conn *hlist.Hlist
}

// New a user seq stored message channel.
func NewSeqChannel() *SeqChannel {
	ch := &SeqChannel{
		mutex:  &sync.Mutex{},
		conn:   hlist.New(),
	}
	// save memory
	glog.V(1).Infof("NewSeqChannel: %v, %v", ch.mutex, ch.conn)
	return ch
}

// AddConn implements the Channel AddConn method.
func (c *SeqChannel) AddConn(key string, conn *Connection) (*hlist.Element, error) {
	c.mutex.Lock()
	if c.conn.Len()+1 > Conf.MaxSubscriberPerChannel {
		c.mutex.Unlock()
		glog.Errorf("user_key:\"%s\" exceed conn", key)
		return nil, errors.New("Exceed the max subscriber connection per key")
	}

	heartbeatInitResponse := &common.HeartbeatInitResponse{}
	sendByte := tunnel.GetMessageBytes(0, int32(common.MessageCommand_HEARTBEAT_INIT_RESPONSE), heartbeatInitResponse)

	// send first heartbeat to tell client service is ready for accept heartbeat
	if _, err := conn.Conn.Write(sendByte); err != nil {
		c.mutex.Unlock()
		glog.Errorf("user_key:\"%s\" write first heartbeat to client error(%v)", key, err)
		return nil, err
	}
	// add conn
	conn.Buf = make(chan []byte, Conf.MsgBufNum)
	conn.HandleWrite(key)
	e := c.conn.PushFront(conn)
	c.mutex.Unlock()
	glog.Infof("user_key:\"%s\" add conn = %d", key, c.conn.Len())
	return e, nil
}

// RemoveConn implements the Channel RemoveConn method.
func (c *SeqChannel) RemoveConn(key string, e *hlist.Element) error {
	c.mutex.Lock()
	c.conn.Remove(e)
	c.mutex.Unlock()
	glog.Infof("user_key:\"%s\" remove conn = %d", key, c.conn.Len())
	return nil
}

// Close implements the Channel Close method.
func (c *SeqChannel) Close() error {
	c.mutex.Lock()
	for e := c.conn.Front(); e != nil; e = e.Next() {
		if conn, ok := e.Value.(*Connection); !ok {
			c.mutex.Unlock()
			return ErrAssectionConn
		} else {
			if err := conn.Conn.Close(); err != nil {
				// ignore close error
				glog.Warningf("conn.Close() error(%v)", err)
			}
		}
	}
	c.mutex.Unlock()
	return nil
}
