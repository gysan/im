package main

import (
	"errors"
	"fmt"
	"github.com/Terry-Mao/gopush-cluster/hash"
	"github.com/Terry-Mao/gopush-cluster/hlist"
	"github.com/golang/glog"
	"net"
	"sync"
)

var (
	ErrChannelNotExist = errors.New("Channle not exist")
	ErrConnProto       = errors.New("Unknown connection protocol")
	UserChannel        *ChannelList
)

// The subscriber interface.
type Channel interface {
	// AddConn add a connection for the subscriber.
	// Exceed the max number of subscribers per key will return errors.
	AddConn(key string, conn *Connection) (*hlist.Element, error)
	// RemoveConn remove a connection for the  subscriber.
	RemoveConn(key string, e *hlist.Element) error
	// Expire expire the channle and clean data.
	Close() error
}

// Connection
type Connection struct {
	Conn    net.Conn
	Buf     chan []byte
}

// HandleWrite start a goroutine get msg from chan, then send to the conn.
func (c *Connection) HandleWrite(key string) {
	go func() {
		var (
			n   int
			err error
		)
		glog.V(1).Infof("user_key: \"%s\" HandleWrite goroutine start", key)
		for {
			msg := <-c.Buf
			n, err = c.Conn.Write([]byte(fmt.Sprintf("$%d\r\n%s\r\n", len(msg), string(msg))))
			// update stat
			if err != nil {
				glog.Errorf("user_key: \"%s\" conn.Write() error(%v)", key, err)
				glog.V(1).Infof("user_key: \"%s\" HandleWrite goroutine stop", key)
				return
			} else {
				glog.V(1).Infof("user_key: \"%s\" conn.Write() %d bytes", key, n)
			}
		}
	}()
}

// Write different message to client by different protocol
func (c *Connection) Write(key string, msg []byte) {
	select {
	case c.Buf <- msg:
	default:
		c.Conn.Close()
		glog.Warningf("user_key: \"%s\" discard message: \"%s\" and close connection", key, string(msg))
	}
}

// Channel bucket.
type ChannelBucket struct {
	Data  map[string]Channel
	mutex *sync.Mutex
}

// Channel list.
type ChannelList struct {
	Channels []*ChannelBucket
}

// Lock lock the bucket mutex.
func (c *ChannelBucket) Lock() {
	c.mutex.Lock()
}

// Unlock unlock the bucket mutex.
func (c *ChannelBucket) Unlock() {
	c.mutex.Unlock()
}

// NewChannelList create a new channel bucket set.
func NewChannelList() *ChannelList {
	l := &ChannelList{Channels: []*ChannelBucket{}}
	// split hashmap to many bucket
	glog.V(1).Infof("create %d ChannelBucket", Conf.ChannelBucket)
	for i := 0; i < Conf.ChannelBucket; i++ {
		c := &ChannelBucket{
			Data:  map[string]Channel{},
			mutex: &sync.Mutex{},
		}
		glog.V(1).Infof("Data: %v, mutex: %v", c.Data, c.mutex)
		l.Channels = append(l.Channels, c)
	}
	return l
}

// Count get the bucket total channel count.
func (l *ChannelList) Count() int {
	c := 0
	for i := 0; i < Conf.ChannelBucket; i++ {
		c += len(l.Channels[i].Data)
	}
	return c
}

// bucket return a channelBucket use murmurhash3.
func (l *ChannelList) bucket(key string) *ChannelBucket {
	h := hash.NewMurmur3C()
	glog.V(1).Infof("hash: %v", h)
	h.Write([]byte(key))
	idx := uint(h.Sum32()) & uint(Conf.ChannelBucket - 1)
	glog.V(1).Infof("user_key:\"%s\" hit channel bucket index:%d, hash: $v", key, idx, h.Sum32())
	return l.Channels[idx]
}

// New create a user channle.
func (l *ChannelList) New(key string) (Channel, error) {
	// get a channel bucket
	b := l.bucket(key)
	b.Lock()
	if c, ok := b.Data[key]; ok {
		b.Unlock()
		glog.Infof("user_key:\"%s\" refresh channel bucket expire time", key)
		return c, nil
	} else {
		c = NewSeqChannel()
		b.Data[key] = c
		b.Unlock()
		glog.Infof("user_key:\"%s\" create a new channel", key)
		return c, nil
	}
}

// Get a user channel from ChannleList.
func (l *ChannelList) Get(key string, newOne bool) (Channel, error) {
	// get a channel bucket
	b := l.bucket(key)
	glog.V(1).Infof("Bucket: %v", b.Data)
	b.Lock()
	if c, ok := b.Data[key]; !ok {
		glog.V(1).Infof("channel: %v", c)
		if newOne {
			c = NewSeqChannel()
			b.Data[key] = c
			b.Unlock()
			glog.Infof("user_key:\"%s\" create a new channel: %v", key, c)
			return c, nil
		} else {
			b.Unlock()
			glog.Warningf("user_key:\"%s\" channle not exists", key)
			return nil, ErrChannelNotExist
		}
	} else {
		b.Unlock()
		glog.Infof("user_key:\"%s\" refresh channel bucket expire time", key)
		return c, nil
	}
}

// Delete a user channel from ChannleList.
func (l *ChannelList) Delete(key string) (Channel, error) {
	// get a channel bucket
	b := l.bucket(key)
	b.Lock()
	if c, ok := b.Data[key]; !ok {
		b.Unlock()
		glog.Warningf("user_key:\"%s\" delete channle not exists", key)
		return nil, ErrChannelNotExist
	} else {
		delete(b.Data, key)
		b.Unlock()
		glog.Infof("user_key:\"%s\" delete channel", key)
		return c, nil
	}
}

// Close close all channel.
func (l *ChannelList) Close() {
	glog.Info("channel close")
	chs := make([]Channel, 0, l.Count())
	for _, c := range l.Channels {
		c.Lock()
		for _, c := range c.Data {
			chs = append(chs, c)
		}
		c.Unlock()
	}
	// close all channels
	for _, c := range chs {
		if err := c.Close(); err != nil {
			glog.Errorf("c.Close() error(%v)", err)
		}
	}
}
