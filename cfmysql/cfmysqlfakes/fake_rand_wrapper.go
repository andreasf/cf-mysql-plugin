// Code generated by counterfeiter. DO NOT EDIT.
package cfmysqlfakes

import (
	"sync"

	"github.com/andreasf/cf-mysql-plugin/cfmysql"
)

type FakeRandWrapper struct {
	IntnStub        func(n int) int
	intnMutex       sync.RWMutex
	intnArgsForCall []struct {
		n int
	}
	intnReturns struct {
		result1 int
	}
	intnReturnsOnCall map[int]struct {
		result1 int
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeRandWrapper) Intn(n int) int {
	fake.intnMutex.Lock()
	ret, specificReturn := fake.intnReturnsOnCall[len(fake.intnArgsForCall)]
	fake.intnArgsForCall = append(fake.intnArgsForCall, struct {
		n int
	}{n})
	fake.recordInvocation("Intn", []interface{}{n})
	fake.intnMutex.Unlock()
	if fake.IntnStub != nil {
		return fake.IntnStub(n)
	}
	if specificReturn {
		return ret.result1
	}
	return fake.intnReturns.result1
}

func (fake *FakeRandWrapper) IntnCallCount() int {
	fake.intnMutex.RLock()
	defer fake.intnMutex.RUnlock()
	return len(fake.intnArgsForCall)
}

func (fake *FakeRandWrapper) IntnArgsForCall(i int) int {
	fake.intnMutex.RLock()
	defer fake.intnMutex.RUnlock()
	return fake.intnArgsForCall[i].n
}

func (fake *FakeRandWrapper) IntnReturns(result1 int) {
	fake.IntnStub = nil
	fake.intnReturns = struct {
		result1 int
	}{result1}
}

func (fake *FakeRandWrapper) IntnReturnsOnCall(i int, result1 int) {
	fake.IntnStub = nil
	if fake.intnReturnsOnCall == nil {
		fake.intnReturnsOnCall = make(map[int]struct {
			result1 int
		})
	}
	fake.intnReturnsOnCall[i] = struct {
		result1 int
	}{result1}
}

func (fake *FakeRandWrapper) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.intnMutex.RLock()
	defer fake.intnMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeRandWrapper) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ cfmysql.RandWrapper = new(FakeRandWrapper)
