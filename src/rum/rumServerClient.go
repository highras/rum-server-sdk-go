package rum

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/highras/fpnn-sdk-go/src/fpnn"
)

const (
	SDKVersion = "0.1.0"
)

type idGenerator struct {
	mutex  sync.Mutex
	idBase int16
}

func (gen *idGenerator) genId() int64 {
	gen.mutex.Lock()
	defer gen.mutex.Unlock()

	now := time.Now()
	mid := now.UnixNano() / 1000000

	gen.idBase += 1
	if gen.idBase > 999 {
		gen.idBase = 0
	}

	return mid*1000 + int64(gen.idBase)
}

type RUMServerClient struct {
	client    *fpnn.TCPClient
	logger    *log.Logger
	idGen     *idGenerator
	pid       int32
	secretKey string
	rid       string
	sid       int64
}

func NewRUMServerClient(pid int32, secretKey string, endpoint string) *RUMServerClient {

	client := &RUMServerClient{}

	client.client = fpnn.NewTCPClient(endpoint)
	client.idGen = &idGenerator{}
	client.secretKey = secretKey
	client.pid = pid
	client.rid = ""
	client.sid = 0

	client.SetLogger(log.New(os.Stdout, "[RUM Go SDK] ", log.LstdFlags))

	return client
}

func (client *RUMServerClient) SetConnectTimeOut(timeout time.Duration) {
	client.client.SetConnectTimeOut(timeout)
}

func (client *RUMServerClient) SetQuestTimeOut(timeout time.Duration) {
	client.client.SetQuestTimeOut(timeout)
}

func (client *RUMServerClient) SetOnConnectedCallback(onConnected func(connId uint64)) {
	client.client.SetOnConnectedCallback(onConnected)
}

func (client *RUMServerClient) SetOnClosedCallback(onClosed func(connId uint64)) {
	client.client.SetOnClosedCallback(onClosed)
}

func (client *RUMServerClient) SetLogger(logger *log.Logger) {
	client.logger = logger
	client.client.SetLogger(logger)
}

func (client *RUMServerClient) SetRumId(rid string) {
	client.rid = rid
}

func (client *RUMServerClient) SetSessionId(sid int64) {
	client.sid = sid
}

func (client *RUMServerClient) Endpoint() string {
	return client.client.Endpoint()
}

/*
	Params:
		rest: can be include following params:
			pemPath		string
			rawPemData	[]byte
			reinforce	bool
*/
func (client *RUMServerClient) EnableEncryptor(rest ...interface{}) (err error) {
	return client.client.EnableEncryptor(rest...)
}

func convertToString(value interface{}) string {
	switch value.(type) {
	case string:
		return value.(string)
	case []byte:
		return string(value.([]byte))
	case []rune:
		return string(value.([]rune))
	default:
		panic("Type convert failed.")
	}
}

func (client *RUMServerClient) GenRumEventPayload(eventName string, attrs interface{}) map[string]interface{} {
	if client.sid == 0 {
		client.sid = client.idGen.genId()
	}
	if client.rid == "" {
		client.rid = strconv.FormatInt(client.idGen.genId(), 10)
	}

	ev := make(map[string]interface{})
	ev["ev"] = eventName
	ev["sid"] = client.sid
	ev["rid"] = client.rid
	ev["ts"] = int32(time.Now().Unix())
	ev["eid"] = client.idGen.genId()
	ev["source"] = "golang"
	ev["attrs"] = attrs
	return ev
}

func (client *RUMServerClient) GenRumQuestByAttrs(eventName string, attrs map[string]interface{}) *fpnn.Quest {
	events := make([]map[string]interface{}, 0, 1)
	events = append(events, client.GenRumEventPayload(eventName, attrs))
	return client.GenRumQuestByList(events)
}

func (client *RUMServerClient) GenRumQuestByList(eventList []map[string]interface{}) *fpnn.Quest {
	events := make([]map[string]interface{}, 0, len(eventList))

	for _, ev := range eventList {
		eventName, okName := ev["ev"]
		if okName {
			attrs, okAttrs := ev["attrs"]
			if okAttrs {
				events = append(events, client.GenRumEventPayload(convertToString(eventName), attrs))
			}
		}
	}

	if len(events) == 0 {
		panic("Invaild params when call RUMServerClient.GenRumQuestByList() function.")
	}

	salt := int32(time.Now().Unix())

	pidStr := strconv.FormatInt(int64(client.pid), 10)
	saltStr := strconv.FormatInt(int64(salt), 10)

	ctx := md5.New()
	io.WriteString(ctx, pidStr)
	io.WriteString(ctx, ":")
	io.WriteString(ctx, client.secretKey)
	io.WriteString(ctx, ":")
	io.WriteString(ctx, saltStr)
	sign := fmt.Sprintf("%X", ctx.Sum(nil))

	quest := fpnn.NewQuest("adds")
	quest.Param("pid", client.pid)
	quest.Param("salt", salt)
	quest.Param("sign", sign)
	quest.Param("events", events)
	return quest
}

func (client *RUMServerClient) sendQuest(quest *fpnn.Quest, timeout time.Duration, callback func(answer *fpnn.Answer, errorCode int)) (*fpnn.Answer, error) {
	if callback == nil {
		if timeout == 0 {
			return client.client.SendQuest(quest)
		} else {
			return client.client.SendQuest(quest, timeout)
		}
	} else {
		if timeout == 0 {
			return nil, client.client.SendQuestWithLambda(quest, callback)
		} else {
			return nil, client.client.SendQuestWithLambda(quest, callback, timeout)
		}
	}
}

func (client *RUMServerClient) sendRumQuest(quest *fpnn.Quest, timeout time.Duration,
	callback func(errorCode int, errInfo string)) error {

	if callback != nil {
		callbackFunc := func(answer *fpnn.Answer, errorCode int) {
			if errorCode == fpnn.FPNN_EC_OK {
				callback(fpnn.FPNN_EC_OK, "")
			} else if answer == nil {
				callback(errorCode, "")
			} else {
				callback(answer.WantInt("code"), answer.WantString("ex"))
			}
		}

		_, err := client.sendQuest(quest, timeout, callbackFunc)
		return err
	}

	answer, err := client.sendQuest(quest, timeout, nil)
	if err != nil {
		return err
	} else if !answer.IsException() {
		return nil
	} else {
		return fmt.Errorf("[Exception] code: %d, ex: %s", answer.WantInt("code"), answer.WantString("ex"))
	}
}

func (client *RUMServerClient) SendCustomEvent(eventName string, attrs map[string]interface{}, rest ...interface{}) error {

	var timeout time.Duration
	var callback func(int, string)

	for _, value := range rest {
		switch value := value.(type) {
		case time.Duration:
			timeout = value
		case func(int, string):
			callback = value
		default:
			panic("Invaild params when call RUMServerClient.SendCustomEvent() function.")
		}
	}

	quest := client.GenRumQuestByAttrs(eventName, attrs)
	return client.sendRumQuest(quest, timeout, callback)
}

func (client *RUMServerClient) SendCustomEvents(eventList []map[string]interface{}, rest ...interface{}) error {
	var timeout time.Duration
	var callback func(int, string)

	for _, value := range rest {
		switch value := value.(type) {
		case time.Duration:
			timeout = value
		case func(int, string):
			callback = value
		default:
			panic("Invaild params when call RUMServerClient.SendCustomEvents() function.")
		}
	}

	quest := client.GenRumQuestByList(eventList)
	return client.sendRumQuest(quest, timeout, callback)
}
