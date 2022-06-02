package internal

import (
	"time"

	"github.com/sirupsen/logrus"
)

// heartbeat polls the EDS server every X seconds and updates the Exposed Things
func (pb *OWServerPB) heartBeat() {
	logrus.Infof("OWServerPB.heartbeat started. TDinterval=%d seconds, Value interval is %d seconds",
		pb.Config.TDInterval, pb.Config.ValueInterval)
	var tdCountDown = 0
	var valueCountDown = 0
	for {
		pb.mu.Lock()
		isRunning := pb.running
		pb.mu.Unlock()
		if !isRunning {
			break
		}

		tdCountDown--
		if tdCountDown <= 0 {
			// create ExposedThing's as they are discovered
			_ = pb.UpdateExposedThings()
			tdCountDown = pb.Config.TDInterval
		}
		valueCountDown--
		if valueCountDown <= 0 {
			_ = pb.UpdatePropertyValues()
			valueCountDown = pb.Config.ValueInterval
		}
		time.Sleep(time.Second)
	}
}
