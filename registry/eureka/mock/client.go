// This file was generated by counterfeiter
package mock

import (
	"sync"
	"time"

	"github.com/st3v/go-eureka"
)

type Client struct {
	RegisterStub        func(*eureka.Instance) error
	registerMutex       sync.RWMutex
	registerArgsForCall []struct {
		arg1 *eureka.Instance
	}
	registerReturns struct {
		result1 error
	}
	DeregisterStub        func(*eureka.Instance) error
	deregisterMutex       sync.RWMutex
	deregisterArgsForCall []struct {
		arg1 *eureka.Instance
	}
	deregisterReturns struct {
		result1 error
	}
	HeartbeatStub        func(*eureka.Instance) error
	heartbeatMutex       sync.RWMutex
	heartbeatArgsForCall []struct {
		arg1 *eureka.Instance
	}
	heartbeatReturns struct {
		result1 error
	}
	AppsStub        func() ([]*eureka.App, error)
	appsMutex       sync.RWMutex
	appsArgsForCall []struct{}
	appsReturns     struct {
		result1 []*eureka.App
		result2 error
	}
	AppStub        func(appName string) (*eureka.App, error)
	appMutex       sync.RWMutex
	appArgsForCall []struct {
		appName string
	}
	appReturns struct {
		result1 *eureka.App
		result2 error
	}
	AppInstanceStub        func(appName, instanceID string) (*eureka.Instance, error)
	appInstanceMutex       sync.RWMutex
	appInstanceArgsForCall []struct {
		appName    string
		instanceID string
	}
	appInstanceReturns struct {
		result1 *eureka.Instance
		result2 error
	}
	WatchStub        func(pollInterval time.Duration) *eureka.Watcher
	watchMutex       sync.RWMutex
	watchArgsForCall []struct {
		pollInterval time.Duration
	}
	watchReturns struct {
		result1 *eureka.Watcher
	}
}

func (fake *Client) Register(arg1 *eureka.Instance) error {
	fake.registerMutex.Lock()
	fake.registerArgsForCall = append(fake.registerArgsForCall, struct {
		arg1 *eureka.Instance
	}{arg1})
	fake.registerMutex.Unlock()
	if fake.RegisterStub != nil {
		return fake.RegisterStub(arg1)
	} else {
		return fake.registerReturns.result1
	}
}

func (fake *Client) RegisterCallCount() int {
	fake.registerMutex.RLock()
	defer fake.registerMutex.RUnlock()
	return len(fake.registerArgsForCall)
}

func (fake *Client) RegisterArgsForCall(i int) *eureka.Instance {
	fake.registerMutex.RLock()
	defer fake.registerMutex.RUnlock()
	return fake.registerArgsForCall[i].arg1
}

func (fake *Client) RegisterReturns(result1 error) {
	fake.RegisterStub = nil
	fake.registerReturns = struct {
		result1 error
	}{result1}
}

func (fake *Client) Deregister(arg1 *eureka.Instance) error {
	fake.deregisterMutex.Lock()
	fake.deregisterArgsForCall = append(fake.deregisterArgsForCall, struct {
		arg1 *eureka.Instance
	}{arg1})
	fake.deregisterMutex.Unlock()
	if fake.DeregisterStub != nil {
		return fake.DeregisterStub(arg1)
	} else {
		return fake.deregisterReturns.result1
	}
}

func (fake *Client) DeregisterCallCount() int {
	fake.deregisterMutex.RLock()
	defer fake.deregisterMutex.RUnlock()
	return len(fake.deregisterArgsForCall)
}

func (fake *Client) DeregisterArgsForCall(i int) *eureka.Instance {
	fake.deregisterMutex.RLock()
	defer fake.deregisterMutex.RUnlock()
	return fake.deregisterArgsForCall[i].arg1
}

func (fake *Client) DeregisterReturns(result1 error) {
	fake.DeregisterStub = nil
	fake.deregisterReturns = struct {
		result1 error
	}{result1}
}

func (fake *Client) Heartbeat(arg1 *eureka.Instance) error {
	fake.heartbeatMutex.Lock()
	fake.heartbeatArgsForCall = append(fake.heartbeatArgsForCall, struct {
		arg1 *eureka.Instance
	}{arg1})
	fake.heartbeatMutex.Unlock()
	if fake.HeartbeatStub != nil {
		return fake.HeartbeatStub(arg1)
	} else {
		return fake.heartbeatReturns.result1
	}
}

func (fake *Client) HeartbeatCallCount() int {
	fake.heartbeatMutex.RLock()
	defer fake.heartbeatMutex.RUnlock()
	return len(fake.heartbeatArgsForCall)
}

func (fake *Client) HeartbeatArgsForCall(i int) *eureka.Instance {
	fake.heartbeatMutex.RLock()
	defer fake.heartbeatMutex.RUnlock()
	return fake.heartbeatArgsForCall[i].arg1
}

func (fake *Client) HeartbeatReturns(result1 error) {
	fake.HeartbeatStub = nil
	fake.heartbeatReturns = struct {
		result1 error
	}{result1}
}

func (fake *Client) Apps() ([]*eureka.App, error) {
	fake.appsMutex.Lock()
	fake.appsArgsForCall = append(fake.appsArgsForCall, struct{}{})
	fake.appsMutex.Unlock()
	if fake.AppsStub != nil {
		return fake.AppsStub()
	} else {
		return fake.appsReturns.result1, fake.appsReturns.result2
	}
}

func (fake *Client) AppsCallCount() int {
	fake.appsMutex.RLock()
	defer fake.appsMutex.RUnlock()
	return len(fake.appsArgsForCall)
}

func (fake *Client) AppsReturns(result1 []*eureka.App, result2 error) {
	fake.AppsStub = nil
	fake.appsReturns = struct {
		result1 []*eureka.App
		result2 error
	}{result1, result2}
}

func (fake *Client) App(appName string) (*eureka.App, error) {
	fake.appMutex.Lock()
	fake.appArgsForCall = append(fake.appArgsForCall, struct {
		appName string
	}{appName})
	fake.appMutex.Unlock()
	if fake.AppStub != nil {
		return fake.AppStub(appName)
	} else {
		return fake.appReturns.result1, fake.appReturns.result2
	}
}

func (fake *Client) AppCallCount() int {
	fake.appMutex.RLock()
	defer fake.appMutex.RUnlock()
	return len(fake.appArgsForCall)
}

func (fake *Client) AppArgsForCall(i int) string {
	fake.appMutex.RLock()
	defer fake.appMutex.RUnlock()
	return fake.appArgsForCall[i].appName
}

func (fake *Client) AppReturns(result1 *eureka.App, result2 error) {
	fake.AppStub = nil
	fake.appReturns = struct {
		result1 *eureka.App
		result2 error
	}{result1, result2}
}

func (fake *Client) AppInstance(appName string, instanceID string) (*eureka.Instance, error) {
	fake.appInstanceMutex.Lock()
	fake.appInstanceArgsForCall = append(fake.appInstanceArgsForCall, struct {
		appName    string
		instanceID string
	}{appName, instanceID})
	fake.appInstanceMutex.Unlock()
	if fake.AppInstanceStub != nil {
		return fake.AppInstanceStub(appName, instanceID)
	} else {
		return fake.appInstanceReturns.result1, fake.appInstanceReturns.result2
	}
}

func (fake *Client) AppInstanceCallCount() int {
	fake.appInstanceMutex.RLock()
	defer fake.appInstanceMutex.RUnlock()
	return len(fake.appInstanceArgsForCall)
}

func (fake *Client) AppInstanceArgsForCall(i int) (string, string) {
	fake.appInstanceMutex.RLock()
	defer fake.appInstanceMutex.RUnlock()
	return fake.appInstanceArgsForCall[i].appName, fake.appInstanceArgsForCall[i].instanceID
}

func (fake *Client) AppInstanceReturns(result1 *eureka.Instance, result2 error) {
	fake.AppInstanceStub = nil
	fake.appInstanceReturns = struct {
		result1 *eureka.Instance
		result2 error
	}{result1, result2}
}

func (fake *Client) Watch(pollInterval time.Duration) *eureka.Watcher {
	fake.watchMutex.Lock()
	fake.watchArgsForCall = append(fake.watchArgsForCall, struct {
		pollInterval time.Duration
	}{pollInterval})
	fake.watchMutex.Unlock()
	if fake.WatchStub != nil {
		return fake.WatchStub(pollInterval)
	} else {
		return fake.watchReturns.result1
	}
}

func (fake *Client) WatchCallCount() int {
	fake.watchMutex.RLock()
	defer fake.watchMutex.RUnlock()
	return len(fake.watchArgsForCall)
}

func (fake *Client) WatchArgsForCall(i int) time.Duration {
	fake.watchMutex.RLock()
	defer fake.watchMutex.RUnlock()
	return fake.watchArgsForCall[i].pollInterval
}

func (fake *Client) WatchReturns(result1 *eureka.Watcher) {
	fake.WatchStub = nil
	fake.watchReturns = struct {
		result1 *eureka.Watcher
	}{result1}
}
