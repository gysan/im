package main

import (
	"github.com/golang/glog"
	"os"
	"os/signal"
	"syscall"
)

// InitSignal register signals handler.
func InitSignal() chan os.Signal {
	glog.Info("InitSignal start.")
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT, syscall.SIGSTOP)
	glog.Info("Signaling . ")
	return c
}

// HandleSignal fetch signal from chan then do exit or reload.
func HandleSignal(c chan os.Signal) {
	glog.Info("HandleSignal start.")
	// Block until a signal is received.
	for {
		s := <-c
		glog.Infof("comet get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGSTOP, syscall.SIGINT:
			glog.Info("Signal quit, term, stop, int.")
			return
		case syscall.SIGHUP:
			glog.Info("Signal hup.")
			// TODO reload
			//return
		default:
			glog.Info("Signal default.")
			return
		}
	}
	glog.Info("HandleSignal end.")
}
