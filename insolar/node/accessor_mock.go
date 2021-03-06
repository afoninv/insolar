package node

// Code generated by http://github.com/gojuno/minimock (dev). DO NOT EDIT.

import (
	"sync"
	mm_atomic "sync/atomic"
	mm_time "time"

	"github.com/gojuno/minimock/v3"
	"github.com/insolar/insolar/insolar"
)

// AccessorMock implements Accessor
type AccessorMock struct {
	t minimock.Tester

	funcAll          func(pulse insolar.PulseNumber) (na1 []insolar.Node, err error)
	inspectFuncAll   func(pulse insolar.PulseNumber)
	afterAllCounter  uint64
	beforeAllCounter uint64
	AllMock          mAccessorMockAll

	funcInRole          func(pulse insolar.PulseNumber, role insolar.StaticRole) (na1 []insolar.Node, err error)
	inspectFuncInRole   func(pulse insolar.PulseNumber, role insolar.StaticRole)
	afterInRoleCounter  uint64
	beforeInRoleCounter uint64
	InRoleMock          mAccessorMockInRole
}

// NewAccessorMock returns a mock for Accessor
func NewAccessorMock(t minimock.Tester) *AccessorMock {
	m := &AccessorMock{t: t}
	if controller, ok := t.(minimock.MockController); ok {
		controller.RegisterMocker(m)
	}

	m.AllMock = mAccessorMockAll{mock: m}
	m.AllMock.callArgs = []*AccessorMockAllParams{}

	m.InRoleMock = mAccessorMockInRole{mock: m}
	m.InRoleMock.callArgs = []*AccessorMockInRoleParams{}

	return m
}

type mAccessorMockAll struct {
	mock               *AccessorMock
	defaultExpectation *AccessorMockAllExpectation
	expectations       []*AccessorMockAllExpectation

	callArgs []*AccessorMockAllParams
	mutex    sync.RWMutex
}

// AccessorMockAllExpectation specifies expectation struct of the Accessor.All
type AccessorMockAllExpectation struct {
	mock    *AccessorMock
	params  *AccessorMockAllParams
	results *AccessorMockAllResults
	Counter uint64
}

// AccessorMockAllParams contains parameters of the Accessor.All
type AccessorMockAllParams struct {
	pulse insolar.PulseNumber
}

// AccessorMockAllResults contains results of the Accessor.All
type AccessorMockAllResults struct {
	na1 []insolar.Node
	err error
}

// Expect sets up expected params for Accessor.All
func (mmAll *mAccessorMockAll) Expect(pulse insolar.PulseNumber) *mAccessorMockAll {
	if mmAll.mock.funcAll != nil {
		mmAll.mock.t.Fatalf("AccessorMock.All mock is already set by Set")
	}

	if mmAll.defaultExpectation == nil {
		mmAll.defaultExpectation = &AccessorMockAllExpectation{}
	}

	mmAll.defaultExpectation.params = &AccessorMockAllParams{pulse}
	for _, e := range mmAll.expectations {
		if minimock.Equal(e.params, mmAll.defaultExpectation.params) {
			mmAll.mock.t.Fatalf("Expectation set by When has same params: %#v", *mmAll.defaultExpectation.params)
		}
	}

	return mmAll
}

// Inspect accepts an inspector function that has same arguments as the Accessor.All
func (mmAll *mAccessorMockAll) Inspect(f func(pulse insolar.PulseNumber)) *mAccessorMockAll {
	if mmAll.mock.inspectFuncAll != nil {
		mmAll.mock.t.Fatalf("Inspect function is already set for AccessorMock.All")
	}

	mmAll.mock.inspectFuncAll = f

	return mmAll
}

// Return sets up results that will be returned by Accessor.All
func (mmAll *mAccessorMockAll) Return(na1 []insolar.Node, err error) *AccessorMock {
	if mmAll.mock.funcAll != nil {
		mmAll.mock.t.Fatalf("AccessorMock.All mock is already set by Set")
	}

	if mmAll.defaultExpectation == nil {
		mmAll.defaultExpectation = &AccessorMockAllExpectation{mock: mmAll.mock}
	}
	mmAll.defaultExpectation.results = &AccessorMockAllResults{na1, err}
	return mmAll.mock
}

//Set uses given function f to mock the Accessor.All method
func (mmAll *mAccessorMockAll) Set(f func(pulse insolar.PulseNumber) (na1 []insolar.Node, err error)) *AccessorMock {
	if mmAll.defaultExpectation != nil {
		mmAll.mock.t.Fatalf("Default expectation is already set for the Accessor.All method")
	}

	if len(mmAll.expectations) > 0 {
		mmAll.mock.t.Fatalf("Some expectations are already set for the Accessor.All method")
	}

	mmAll.mock.funcAll = f
	return mmAll.mock
}

// When sets expectation for the Accessor.All which will trigger the result defined by the following
// Then helper
func (mmAll *mAccessorMockAll) When(pulse insolar.PulseNumber) *AccessorMockAllExpectation {
	if mmAll.mock.funcAll != nil {
		mmAll.mock.t.Fatalf("AccessorMock.All mock is already set by Set")
	}

	expectation := &AccessorMockAllExpectation{
		mock:   mmAll.mock,
		params: &AccessorMockAllParams{pulse},
	}
	mmAll.expectations = append(mmAll.expectations, expectation)
	return expectation
}

// Then sets up Accessor.All return parameters for the expectation previously defined by the When method
func (e *AccessorMockAllExpectation) Then(na1 []insolar.Node, err error) *AccessorMock {
	e.results = &AccessorMockAllResults{na1, err}
	return e.mock
}

// All implements Accessor
func (mmAll *AccessorMock) All(pulse insolar.PulseNumber) (na1 []insolar.Node, err error) {
	mm_atomic.AddUint64(&mmAll.beforeAllCounter, 1)
	defer mm_atomic.AddUint64(&mmAll.afterAllCounter, 1)

	if mmAll.inspectFuncAll != nil {
		mmAll.inspectFuncAll(pulse)
	}

	mm_params := &AccessorMockAllParams{pulse}

	// Record call args
	mmAll.AllMock.mutex.Lock()
	mmAll.AllMock.callArgs = append(mmAll.AllMock.callArgs, mm_params)
	mmAll.AllMock.mutex.Unlock()

	for _, e := range mmAll.AllMock.expectations {
		if minimock.Equal(e.params, mm_params) {
			mm_atomic.AddUint64(&e.Counter, 1)
			return e.results.na1, e.results.err
		}
	}

	if mmAll.AllMock.defaultExpectation != nil {
		mm_atomic.AddUint64(&mmAll.AllMock.defaultExpectation.Counter, 1)
		mm_want := mmAll.AllMock.defaultExpectation.params
		mm_got := AccessorMockAllParams{pulse}
		if mm_want != nil && !minimock.Equal(*mm_want, mm_got) {
			mmAll.t.Errorf("AccessorMock.All got unexpected parameters, want: %#v, got: %#v%s\n", *mm_want, mm_got, minimock.Diff(*mm_want, mm_got))
		}

		mm_results := mmAll.AllMock.defaultExpectation.results
		if mm_results == nil {
			mmAll.t.Fatal("No results are set for the AccessorMock.All")
		}
		return (*mm_results).na1, (*mm_results).err
	}
	if mmAll.funcAll != nil {
		return mmAll.funcAll(pulse)
	}
	mmAll.t.Fatalf("Unexpected call to AccessorMock.All. %v", pulse)
	return
}

// AllAfterCounter returns a count of finished AccessorMock.All invocations
func (mmAll *AccessorMock) AllAfterCounter() uint64 {
	return mm_atomic.LoadUint64(&mmAll.afterAllCounter)
}

// AllBeforeCounter returns a count of AccessorMock.All invocations
func (mmAll *AccessorMock) AllBeforeCounter() uint64 {
	return mm_atomic.LoadUint64(&mmAll.beforeAllCounter)
}

// Calls returns a list of arguments used in each call to AccessorMock.All.
// The list is in the same order as the calls were made (i.e. recent calls have a higher index)
func (mmAll *mAccessorMockAll) Calls() []*AccessorMockAllParams {
	mmAll.mutex.RLock()

	argCopy := make([]*AccessorMockAllParams, len(mmAll.callArgs))
	copy(argCopy, mmAll.callArgs)

	mmAll.mutex.RUnlock()

	return argCopy
}

// MinimockAllDone returns true if the count of the All invocations corresponds
// the number of defined expectations
func (m *AccessorMock) MinimockAllDone() bool {
	for _, e := range m.AllMock.expectations {
		if mm_atomic.LoadUint64(&e.Counter) < 1 {
			return false
		}
	}

	// if default expectation was set then invocations count should be greater than zero
	if m.AllMock.defaultExpectation != nil && mm_atomic.LoadUint64(&m.afterAllCounter) < 1 {
		return false
	}
	// if func was set then invocations count should be greater than zero
	if m.funcAll != nil && mm_atomic.LoadUint64(&m.afterAllCounter) < 1 {
		return false
	}
	return true
}

// MinimockAllInspect logs each unmet expectation
func (m *AccessorMock) MinimockAllInspect() {
	for _, e := range m.AllMock.expectations {
		if mm_atomic.LoadUint64(&e.Counter) < 1 {
			m.t.Errorf("Expected call to AccessorMock.All with params: %#v", *e.params)
		}
	}

	// if default expectation was set then invocations count should be greater than zero
	if m.AllMock.defaultExpectation != nil && mm_atomic.LoadUint64(&m.afterAllCounter) < 1 {
		if m.AllMock.defaultExpectation.params == nil {
			m.t.Error("Expected call to AccessorMock.All")
		} else {
			m.t.Errorf("Expected call to AccessorMock.All with params: %#v", *m.AllMock.defaultExpectation.params)
		}
	}
	// if func was set then invocations count should be greater than zero
	if m.funcAll != nil && mm_atomic.LoadUint64(&m.afterAllCounter) < 1 {
		m.t.Error("Expected call to AccessorMock.All")
	}
}

type mAccessorMockInRole struct {
	mock               *AccessorMock
	defaultExpectation *AccessorMockInRoleExpectation
	expectations       []*AccessorMockInRoleExpectation

	callArgs []*AccessorMockInRoleParams
	mutex    sync.RWMutex
}

// AccessorMockInRoleExpectation specifies expectation struct of the Accessor.InRole
type AccessorMockInRoleExpectation struct {
	mock    *AccessorMock
	params  *AccessorMockInRoleParams
	results *AccessorMockInRoleResults
	Counter uint64
}

// AccessorMockInRoleParams contains parameters of the Accessor.InRole
type AccessorMockInRoleParams struct {
	pulse insolar.PulseNumber
	role  insolar.StaticRole
}

// AccessorMockInRoleResults contains results of the Accessor.InRole
type AccessorMockInRoleResults struct {
	na1 []insolar.Node
	err error
}

// Expect sets up expected params for Accessor.InRole
func (mmInRole *mAccessorMockInRole) Expect(pulse insolar.PulseNumber, role insolar.StaticRole) *mAccessorMockInRole {
	if mmInRole.mock.funcInRole != nil {
		mmInRole.mock.t.Fatalf("AccessorMock.InRole mock is already set by Set")
	}

	if mmInRole.defaultExpectation == nil {
		mmInRole.defaultExpectation = &AccessorMockInRoleExpectation{}
	}

	mmInRole.defaultExpectation.params = &AccessorMockInRoleParams{pulse, role}
	for _, e := range mmInRole.expectations {
		if minimock.Equal(e.params, mmInRole.defaultExpectation.params) {
			mmInRole.mock.t.Fatalf("Expectation set by When has same params: %#v", *mmInRole.defaultExpectation.params)
		}
	}

	return mmInRole
}

// Inspect accepts an inspector function that has same arguments as the Accessor.InRole
func (mmInRole *mAccessorMockInRole) Inspect(f func(pulse insolar.PulseNumber, role insolar.StaticRole)) *mAccessorMockInRole {
	if mmInRole.mock.inspectFuncInRole != nil {
		mmInRole.mock.t.Fatalf("Inspect function is already set for AccessorMock.InRole")
	}

	mmInRole.mock.inspectFuncInRole = f

	return mmInRole
}

// Return sets up results that will be returned by Accessor.InRole
func (mmInRole *mAccessorMockInRole) Return(na1 []insolar.Node, err error) *AccessorMock {
	if mmInRole.mock.funcInRole != nil {
		mmInRole.mock.t.Fatalf("AccessorMock.InRole mock is already set by Set")
	}

	if mmInRole.defaultExpectation == nil {
		mmInRole.defaultExpectation = &AccessorMockInRoleExpectation{mock: mmInRole.mock}
	}
	mmInRole.defaultExpectation.results = &AccessorMockInRoleResults{na1, err}
	return mmInRole.mock
}

//Set uses given function f to mock the Accessor.InRole method
func (mmInRole *mAccessorMockInRole) Set(f func(pulse insolar.PulseNumber, role insolar.StaticRole) (na1 []insolar.Node, err error)) *AccessorMock {
	if mmInRole.defaultExpectation != nil {
		mmInRole.mock.t.Fatalf("Default expectation is already set for the Accessor.InRole method")
	}

	if len(mmInRole.expectations) > 0 {
		mmInRole.mock.t.Fatalf("Some expectations are already set for the Accessor.InRole method")
	}

	mmInRole.mock.funcInRole = f
	return mmInRole.mock
}

// When sets expectation for the Accessor.InRole which will trigger the result defined by the following
// Then helper
func (mmInRole *mAccessorMockInRole) When(pulse insolar.PulseNumber, role insolar.StaticRole) *AccessorMockInRoleExpectation {
	if mmInRole.mock.funcInRole != nil {
		mmInRole.mock.t.Fatalf("AccessorMock.InRole mock is already set by Set")
	}

	expectation := &AccessorMockInRoleExpectation{
		mock:   mmInRole.mock,
		params: &AccessorMockInRoleParams{pulse, role},
	}
	mmInRole.expectations = append(mmInRole.expectations, expectation)
	return expectation
}

// Then sets up Accessor.InRole return parameters for the expectation previously defined by the When method
func (e *AccessorMockInRoleExpectation) Then(na1 []insolar.Node, err error) *AccessorMock {
	e.results = &AccessorMockInRoleResults{na1, err}
	return e.mock
}

// InRole implements Accessor
func (mmInRole *AccessorMock) InRole(pulse insolar.PulseNumber, role insolar.StaticRole) (na1 []insolar.Node, err error) {
	mm_atomic.AddUint64(&mmInRole.beforeInRoleCounter, 1)
	defer mm_atomic.AddUint64(&mmInRole.afterInRoleCounter, 1)

	if mmInRole.inspectFuncInRole != nil {
		mmInRole.inspectFuncInRole(pulse, role)
	}

	mm_params := &AccessorMockInRoleParams{pulse, role}

	// Record call args
	mmInRole.InRoleMock.mutex.Lock()
	mmInRole.InRoleMock.callArgs = append(mmInRole.InRoleMock.callArgs, mm_params)
	mmInRole.InRoleMock.mutex.Unlock()

	for _, e := range mmInRole.InRoleMock.expectations {
		if minimock.Equal(e.params, mm_params) {
			mm_atomic.AddUint64(&e.Counter, 1)
			return e.results.na1, e.results.err
		}
	}

	if mmInRole.InRoleMock.defaultExpectation != nil {
		mm_atomic.AddUint64(&mmInRole.InRoleMock.defaultExpectation.Counter, 1)
		mm_want := mmInRole.InRoleMock.defaultExpectation.params
		mm_got := AccessorMockInRoleParams{pulse, role}
		if mm_want != nil && !minimock.Equal(*mm_want, mm_got) {
			mmInRole.t.Errorf("AccessorMock.InRole got unexpected parameters, want: %#v, got: %#v%s\n", *mm_want, mm_got, minimock.Diff(*mm_want, mm_got))
		}

		mm_results := mmInRole.InRoleMock.defaultExpectation.results
		if mm_results == nil {
			mmInRole.t.Fatal("No results are set for the AccessorMock.InRole")
		}
		return (*mm_results).na1, (*mm_results).err
	}
	if mmInRole.funcInRole != nil {
		return mmInRole.funcInRole(pulse, role)
	}
	mmInRole.t.Fatalf("Unexpected call to AccessorMock.InRole. %v %v", pulse, role)
	return
}

// InRoleAfterCounter returns a count of finished AccessorMock.InRole invocations
func (mmInRole *AccessorMock) InRoleAfterCounter() uint64 {
	return mm_atomic.LoadUint64(&mmInRole.afterInRoleCounter)
}

// InRoleBeforeCounter returns a count of AccessorMock.InRole invocations
func (mmInRole *AccessorMock) InRoleBeforeCounter() uint64 {
	return mm_atomic.LoadUint64(&mmInRole.beforeInRoleCounter)
}

// Calls returns a list of arguments used in each call to AccessorMock.InRole.
// The list is in the same order as the calls were made (i.e. recent calls have a higher index)
func (mmInRole *mAccessorMockInRole) Calls() []*AccessorMockInRoleParams {
	mmInRole.mutex.RLock()

	argCopy := make([]*AccessorMockInRoleParams, len(mmInRole.callArgs))
	copy(argCopy, mmInRole.callArgs)

	mmInRole.mutex.RUnlock()

	return argCopy
}

// MinimockInRoleDone returns true if the count of the InRole invocations corresponds
// the number of defined expectations
func (m *AccessorMock) MinimockInRoleDone() bool {
	for _, e := range m.InRoleMock.expectations {
		if mm_atomic.LoadUint64(&e.Counter) < 1 {
			return false
		}
	}

	// if default expectation was set then invocations count should be greater than zero
	if m.InRoleMock.defaultExpectation != nil && mm_atomic.LoadUint64(&m.afterInRoleCounter) < 1 {
		return false
	}
	// if func was set then invocations count should be greater than zero
	if m.funcInRole != nil && mm_atomic.LoadUint64(&m.afterInRoleCounter) < 1 {
		return false
	}
	return true
}

// MinimockInRoleInspect logs each unmet expectation
func (m *AccessorMock) MinimockInRoleInspect() {
	for _, e := range m.InRoleMock.expectations {
		if mm_atomic.LoadUint64(&e.Counter) < 1 {
			m.t.Errorf("Expected call to AccessorMock.InRole with params: %#v", *e.params)
		}
	}

	// if default expectation was set then invocations count should be greater than zero
	if m.InRoleMock.defaultExpectation != nil && mm_atomic.LoadUint64(&m.afterInRoleCounter) < 1 {
		if m.InRoleMock.defaultExpectation.params == nil {
			m.t.Error("Expected call to AccessorMock.InRole")
		} else {
			m.t.Errorf("Expected call to AccessorMock.InRole with params: %#v", *m.InRoleMock.defaultExpectation.params)
		}
	}
	// if func was set then invocations count should be greater than zero
	if m.funcInRole != nil && mm_atomic.LoadUint64(&m.afterInRoleCounter) < 1 {
		m.t.Error("Expected call to AccessorMock.InRole")
	}
}

// MinimockFinish checks that all mocked methods have been called the expected number of times
func (m *AccessorMock) MinimockFinish() {
	if !m.minimockDone() {
		m.MinimockAllInspect()

		m.MinimockInRoleInspect()
		m.t.FailNow()
	}
}

// MinimockWait waits for all mocked methods to be called the expected number of times
func (m *AccessorMock) MinimockWait(timeout mm_time.Duration) {
	timeoutCh := mm_time.After(timeout)
	for {
		if m.minimockDone() {
			return
		}
		select {
		case <-timeoutCh:
			m.MinimockFinish()
			return
		case <-mm_time.After(10 * mm_time.Millisecond):
		}
	}
}

func (m *AccessorMock) minimockDone() bool {
	done := true
	return done &&
		m.MinimockAllDone() &&
		m.MinimockInRoleDone()
}
