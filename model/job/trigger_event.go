package job

import (
	"github.com/cozy/cozy-stack/pkg/consts"
	"github.com/cozy/cozy-stack/pkg/realtime"
)

// EventTrigger implements Trigger for realtime triggered events
type EventTrigger struct {
	*TriggerInfos
	unscheduled chan struct{}
}

// NewEventTrigger returns a new instance of EventTrigger given the specified
// options.
func NewEventTrigger(infos *TriggerInfos) (*EventTrigger, error) {
	return &EventTrigger{
		TriggerInfos: infos,
		unscheduled:  make(chan struct{}),
	}, nil
}

// Type implements the Type method of the Trigger interface.
func (t *EventTrigger) Type() string {
	return t.TriggerInfos.Type
}

// DocType implements the permission.Matcher interface
func (t *EventTrigger) DocType() string {
	return consts.Triggers
}

// ID implements the permission.Matcher interface
func (t *EventTrigger) ID() string {
	return t.TriggerInfos.TID
}

// Match implements the permission.Matcher interface
func (t *EventTrigger) Match(key, value string) bool {
	switch key {
	case WorkerType:
		return t.TriggerInfos.WorkerType == value
	}
	return false
}

// Schedule implements the Schedule method of the Trigger interface.
func (t *EventTrigger) Schedule() <-chan *JobRequest {
	ch := make(chan *JobRequest)
	go func() {
		sub := realtime.GetHub().Subscriber(t)
		defer func() {
			sub.Close()
			close(ch)
		}()
		for {
			select {
			case e := <-sub.Channel:
				found := false
				if found {
					if evt, err := t.Infos().JobRequestWithEvent(e); err == nil {
						ch <- evt
					}
				}
			case <-t.unscheduled:
				return
			}
		}
	}()
	return ch
}

// Unschedule implements the Unschedule method of the Trigger interface.
func (t *EventTrigger) Unschedule() {
	close(t.unscheduled)
}

// Infos implements the Infos method of the Trigger interface.
func (t *EventTrigger) Infos() *TriggerInfos {
	return t.TriggerInfos
}
