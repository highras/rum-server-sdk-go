# RUM Server-End Go SDK

## Depends & Install & Update

### Depends

	go get github.com/highras/fpnn-sdk-go/src/fpnn

### Install

	go get github.com/highras/rum-server-sdk-go/src/rum

### Update

	go get -u github.com/highras/rum-server-sdk-go/src/rum

### Use

	import "github.com/highras/rum-server-sdk-go/src/rum"


## Usage

### Create

	client := rum.NewRUMServerClient(pid int32, secretKey string, endpoint string)

Please get your project params from RUM Console.

### Configure (Optional)

* Basic configs

		client.SetConnectTimeOut(timeout time.Duration)
		client.SetQuestTimeOut(timeout time.Duration)
		client.SetLogger(logger *log.Logger)

* Set connection events' callbacks

		client.SetOnConnectedCallback(onConnected func(connId uint64))
		client.SetOnClosedCallback(onClosed func(connId uint64))

* Config encrypted connection
	
		client.EnableEncryptor(pemKeyPath string)
		client.EnableEncryptor(pemKeyData []byte)

	RUM Server-End Go SDK using **ECC**/**ECDH** to exchange the secret key, and using **AES-128** or **AES-256** in **CFB** mode to encrypt the whole session in **stream** way.


### Send Custom Event

	// sync
	err := SendCustomEvent(eventName string, attrs map[string]interface{})
	err := SendCustomEvent(eventName string, attrs map[string]interface{}, timeout time.Duration)
	
	// async
	err := SendCustomEvent(eventName string, attrs map[string]interface{}, callback func (errorCode int, errInfo string))
	err := SendCustomEvent(eventName string, attrs map[string]interface{}, callback func (errorCode int, errInfo string), timeout time.Duration)

### Send Custom Events

	// sync
	err := SendCustomEvents(eventList []map[string]interface{})
	err := SendCustomEvents(eventList []map[string]interface{}, timeout time.Duration)
	
	// async
	err := SendCustomEvents(eventList []map[string]interface{}, callback func (errorCode int, errInfo string))
	err := SendCustomEvents(eventList []map[string]interface{}, callback func (errorCode int, errInfo string), timeout time.Duration)
	
### Set Rum ID And Session ID (Optional, If not specified, a random one will be generated)

	SetRumId(rid string)
	SetSessionId(sid int64)

### Demo

	client := rum.NewRUMServerClient(41000015, "xxxxxx-xxx-xxx-xxx-xxxxxx", "52.83.220.166:13609")
	
	attrs := make(map[string]interface{})
	attrs["aaa"] = "bbb"
	attrs["bbb"] = 123

	err1 := client.SendCustomEvent("error", attrs)
	
	eventMap := make(map[string]interface{})
	eventMap["ev"] = "error"
	eventMap["attrs"] = attrs

	eventList := make([]map[string]interface{}, 0, 2)
	eventList = append(eventList, eventMap)
	eventList = append(eventList, eventMap)

	err2 := client.SendCustomEvents(eventList)


### SDK Version

	fmt.Println("RUM Server-End Go SDK Version:", rum.SDKVersion)

## Directory structure

* **<rum-server-sdk-go>/src**

	Codes of SDK.

* **<rum-server-sdk-go>/example**

	Examples codes for using this SDK.


