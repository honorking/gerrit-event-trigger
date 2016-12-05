package observer

import (
	"bytes"
	"container/list"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"sync"
	"time"

	"gerrit-observatory/log"
	"gerrit-observatory/redis"
)

type EventChan chan map[string]interface{}

type ObserverContr struct {
	*sync.Mutex
	incomingEvent EventChan
	observers     *list.List
	ObserverMap   map[int]*list.Element
	Timeout       int
}

type Observer struct {
	http.Client
	subscribe       *redis.Subscribe
	subscribeFilter map[string]interface{}
	eventChan       EventChan
	contr           *ObserverContr
}

func NewObserverContr(c chan map[string]interface{}, Timeout int) *ObserverContr {
	return &ObserverContr{
		Mutex:         &sync.Mutex{},
		incomingEvent: c,
		observers:     list.New(),
		ObserverMap:   make(map[int]*list.Element),
		Timeout:       Timeout,
	}
}

func NewObserver(sub *redis.Subscribe, ch EventChan, contr *ObserverContr) (obs *Observer, err error) {
	filter := make(map[string]interface{})
	if err = mustCompileFilter(sub.Detail.Filter, filter); err != nil {
		return nil, err
	}
	obs = &Observer{
		Client:          http.Client{Timeout: time.Duration(contr.Timeout) * time.Second},
		subscribe:       sub,
		subscribeFilter: filter,
		eventChan:       ch,
		contr:           contr,
	}
	return
}

func (contr *ObserverContr) Start() {
	for {
		msg := <-contr.incomingEvent

		for e := contr.observers.Front(); e != nil; e = e.Next() {
			observer := e.Value.(*Observer)
			select {
			case observer.eventChan <- msg:
				log.Logger.Infof("event: %v pushed to observer: %v", msg, observer)
			default:
				log.Logger.Warningf("event: %v NOT pushed to observer: %v", msg, observer)
			}
		}
	}
}

func (contr *ObserverContr) AddObserver(sub *redis.Subscribe) (err error) {
	contr.Lock()
	defer contr.Unlock()

	if _, ok := contr.ObserverMap[sub.ID]; ok {
		return fmt.Errorf("subscribe id %d already added", sub.ID)
	}
	eventChan := make(EventChan, 100)
	observer, err := NewObserver(sub, eventChan, contr)
	if err != nil {
		return err
	}
	go observer.Start()
	element := contr.observers.PushFront(observer)
	contr.ObserverMap[sub.ID] = element
	return
}

func (contr *ObserverContr) RemoveObserver(id int) (err error) {
	contr.Lock()
	defer contr.Unlock()

	element, ok := contr.ObserverMap[id]
	if !ok {
		return fmt.Errorf("subscribe id %d not existed", id)
	}
	observer := element.Value.(*Observer)
	if !observer.subscribe.Valid {
		return fmt.Errorf("subscribe id %d already removed", id)
	}
	if err = observer.subscribe.Invalid(); err != nil {
		return
	}
	contr.observers.Remove(element)
	close(observer.eventChan)
	return
}

func (obs *Observer) Start() {
	for {
		msg, ok := <-obs.eventChan
		if !ok {
			// TODO: log here
			obs.contr.Lock()
			defer obs.contr.Unlock()
			delete(obs.contr.ObserverMap, obs.subscribe.ID)
			break
		}
		matched, err := msgCompare(obs.subscribeFilter, msg)
		if err != nil {
			// TODO: log here
		}
		if !matched {
			log.Logger.Warningf("event:%v not match filter:%v", msg, obs.subscribeFilter)
			continue
		}
		bodyBytes, err := json.Marshal(msg)
		if err != nil {
			// TODO: log here
		}
		req, err := http.NewRequest("POST", obs.subscribe.Detail.HookURL, bytes.NewReader(bodyBytes))
		req.Header.Set("User-Agent", "Gerrit_Observatory")
		resp, err := obs.Do(req)
		if err != nil {
			log.Logger.Warningf("callbak %v err, err: %v", obs.subscribe.Detail.HookURL, err.Error())
		}
		if !(resp.StatusCode >= 200 || resp.StatusCode < 400) {
			log.Logger.Warningf("callbak %v err, status_code: %v", obs.subscribe.Detail.HookURL, resp.StatusCode)
		}
		err = obs.subscribe.Activate()
		if err != nil {
			// TODO: log here
		}
	}
}

func mustCompileFilter(raw map[string]interface{}, target map[string]interface{}) error {
	for filterKey, filterValue := range raw {
		switch filterValue.(type) {
		case map[string]interface{}:
			newTarget := make(map[string]interface{})
			if err := mustCompileFilter(filterValue.(map[string]interface{}), newTarget); err != nil {
				return err
			}
			target[filterKey] = newTarget
		case string:
			re, err := regexp.Compile(filterValue.(string))
			if err != nil {
				return err
			}
			target[filterKey] = re
		default:
			target[filterKey] = filterValue
		}
	}
	return nil
}

func msgCompare(filter map[string]interface{}, msg map[string]interface{}) (matched bool, err error) {
	for filterKey, filterValue := range filter {
		msgValue, ok := msg[filterKey]
		if !ok {
			return false, nil
		}
		switch filterValue.(type) {
		case map[string]interface{}:
			msgValueMap, ok := msgValue.(map[string]interface{})
			if !ok {
				return false, nil
			}
			ret, _ := msgCompare(filterValue.(map[string]interface{}), msgValueMap)
			if !ret {
				return false, nil
			}
		case *regexp.Regexp:
			msgValueBytes, ok := msgValue.(string)
			if !ok {
				return false, nil
			}
			if !filterValue.(*regexp.Regexp).MatchString(msgValueBytes) {
				return false, nil
			}
		default:
			if filterValue != msgValue {
				return false, nil
			}
		}
	}
	return true, nil
}
