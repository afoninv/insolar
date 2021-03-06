package object

// Code generated by http://github.com/gojuno/minimock (dev). DO NOT EDIT.

import (
	"context"
	"sync"
	mm_atomic "sync/atomic"
	mm_time "time"

	"github.com/gojuno/minimock/v3"
	"github.com/insolar/insolar/insolar"
	"github.com/insolar/insolar/insolar/record"
)

// RecordAccessorMock implements RecordAccessor
type RecordAccessorMock struct {
	t minimock.Tester

	funcForID          func(ctx context.Context, id insolar.ID) (m1 record.Material, err error)
	inspectFuncForID   func(ctx context.Context, id insolar.ID)
	afterForIDCounter  uint64
	beforeForIDCounter uint64
	ForIDMock          mRecordAccessorMockForID
}

// NewRecordAccessorMock returns a mock for RecordAccessor
func NewRecordAccessorMock(t minimock.Tester) *RecordAccessorMock {
	m := &RecordAccessorMock{t: t}
	if controller, ok := t.(minimock.MockController); ok {
		controller.RegisterMocker(m)
	}

	m.ForIDMock = mRecordAccessorMockForID{mock: m}
	m.ForIDMock.callArgs = []*RecordAccessorMockForIDParams{}

	return m
}

type mRecordAccessorMockForID struct {
	mock               *RecordAccessorMock
	defaultExpectation *RecordAccessorMockForIDExpectation
	expectations       []*RecordAccessorMockForIDExpectation

	callArgs []*RecordAccessorMockForIDParams
	mutex    sync.RWMutex
}

// RecordAccessorMockForIDExpectation specifies expectation struct of the RecordAccessor.ForID
type RecordAccessorMockForIDExpectation struct {
	mock    *RecordAccessorMock
	params  *RecordAccessorMockForIDParams
	results *RecordAccessorMockForIDResults
	Counter uint64
}

// RecordAccessorMockForIDParams contains parameters of the RecordAccessor.ForID
type RecordAccessorMockForIDParams struct {
	ctx context.Context
	id  insolar.ID
}

// RecordAccessorMockForIDResults contains results of the RecordAccessor.ForID
type RecordAccessorMockForIDResults struct {
	m1  record.Material
	err error
}

// Expect sets up expected params for RecordAccessor.ForID
func (mmForID *mRecordAccessorMockForID) Expect(ctx context.Context, id insolar.ID) *mRecordAccessorMockForID {
	if mmForID.mock.funcForID != nil {
		mmForID.mock.t.Fatalf("RecordAccessorMock.ForID mock is already set by Set")
	}

	if mmForID.defaultExpectation == nil {
		mmForID.defaultExpectation = &RecordAccessorMockForIDExpectation{}
	}

	mmForID.defaultExpectation.params = &RecordAccessorMockForIDParams{ctx, id}
	for _, e := range mmForID.expectations {
		if minimock.Equal(e.params, mmForID.defaultExpectation.params) {
			mmForID.mock.t.Fatalf("Expectation set by When has same params: %#v", *mmForID.defaultExpectation.params)
		}
	}

	return mmForID
}

// Inspect accepts an inspector function that has same arguments as the RecordAccessor.ForID
func (mmForID *mRecordAccessorMockForID) Inspect(f func(ctx context.Context, id insolar.ID)) *mRecordAccessorMockForID {
	if mmForID.mock.inspectFuncForID != nil {
		mmForID.mock.t.Fatalf("Inspect function is already set for RecordAccessorMock.ForID")
	}

	mmForID.mock.inspectFuncForID = f

	return mmForID
}

// Return sets up results that will be returned by RecordAccessor.ForID
func (mmForID *mRecordAccessorMockForID) Return(m1 record.Material, err error) *RecordAccessorMock {
	if mmForID.mock.funcForID != nil {
		mmForID.mock.t.Fatalf("RecordAccessorMock.ForID mock is already set by Set")
	}

	if mmForID.defaultExpectation == nil {
		mmForID.defaultExpectation = &RecordAccessorMockForIDExpectation{mock: mmForID.mock}
	}
	mmForID.defaultExpectation.results = &RecordAccessorMockForIDResults{m1, err}
	return mmForID.mock
}

//Set uses given function f to mock the RecordAccessor.ForID method
func (mmForID *mRecordAccessorMockForID) Set(f func(ctx context.Context, id insolar.ID) (m1 record.Material, err error)) *RecordAccessorMock {
	if mmForID.defaultExpectation != nil {
		mmForID.mock.t.Fatalf("Default expectation is already set for the RecordAccessor.ForID method")
	}

	if len(mmForID.expectations) > 0 {
		mmForID.mock.t.Fatalf("Some expectations are already set for the RecordAccessor.ForID method")
	}

	mmForID.mock.funcForID = f
	return mmForID.mock
}

// When sets expectation for the RecordAccessor.ForID which will trigger the result defined by the following
// Then helper
func (mmForID *mRecordAccessorMockForID) When(ctx context.Context, id insolar.ID) *RecordAccessorMockForIDExpectation {
	if mmForID.mock.funcForID != nil {
		mmForID.mock.t.Fatalf("RecordAccessorMock.ForID mock is already set by Set")
	}

	expectation := &RecordAccessorMockForIDExpectation{
		mock:   mmForID.mock,
		params: &RecordAccessorMockForIDParams{ctx, id},
	}
	mmForID.expectations = append(mmForID.expectations, expectation)
	return expectation
}

// Then sets up RecordAccessor.ForID return parameters for the expectation previously defined by the When method
func (e *RecordAccessorMockForIDExpectation) Then(m1 record.Material, err error) *RecordAccessorMock {
	e.results = &RecordAccessorMockForIDResults{m1, err}
	return e.mock
}

// ForID implements RecordAccessor
func (mmForID *RecordAccessorMock) ForID(ctx context.Context, id insolar.ID) (m1 record.Material, err error) {
	mm_atomic.AddUint64(&mmForID.beforeForIDCounter, 1)
	defer mm_atomic.AddUint64(&mmForID.afterForIDCounter, 1)

	if mmForID.inspectFuncForID != nil {
		mmForID.inspectFuncForID(ctx, id)
	}

	mm_params := &RecordAccessorMockForIDParams{ctx, id}

	// Record call args
	mmForID.ForIDMock.mutex.Lock()
	mmForID.ForIDMock.callArgs = append(mmForID.ForIDMock.callArgs, mm_params)
	mmForID.ForIDMock.mutex.Unlock()

	for _, e := range mmForID.ForIDMock.expectations {
		if minimock.Equal(e.params, mm_params) {
			mm_atomic.AddUint64(&e.Counter, 1)
			return e.results.m1, e.results.err
		}
	}

	if mmForID.ForIDMock.defaultExpectation != nil {
		mm_atomic.AddUint64(&mmForID.ForIDMock.defaultExpectation.Counter, 1)
		mm_want := mmForID.ForIDMock.defaultExpectation.params
		mm_got := RecordAccessorMockForIDParams{ctx, id}
		if mm_want != nil && !minimock.Equal(*mm_want, mm_got) {
			mmForID.t.Errorf("RecordAccessorMock.ForID got unexpected parameters, want: %#v, got: %#v%s\n", *mm_want, mm_got, minimock.Diff(*mm_want, mm_got))
		}

		mm_results := mmForID.ForIDMock.defaultExpectation.results
		if mm_results == nil {
			mmForID.t.Fatal("No results are set for the RecordAccessorMock.ForID")
		}
		return (*mm_results).m1, (*mm_results).err
	}
	if mmForID.funcForID != nil {
		return mmForID.funcForID(ctx, id)
	}
	mmForID.t.Fatalf("Unexpected call to RecordAccessorMock.ForID. %v %v", ctx, id)
	return
}

// ForIDAfterCounter returns a count of finished RecordAccessorMock.ForID invocations
func (mmForID *RecordAccessorMock) ForIDAfterCounter() uint64 {
	return mm_atomic.LoadUint64(&mmForID.afterForIDCounter)
}

// ForIDBeforeCounter returns a count of RecordAccessorMock.ForID invocations
func (mmForID *RecordAccessorMock) ForIDBeforeCounter() uint64 {
	return mm_atomic.LoadUint64(&mmForID.beforeForIDCounter)
}

// Calls returns a list of arguments used in each call to RecordAccessorMock.ForID.
// The list is in the same order as the calls were made (i.e. recent calls have a higher index)
func (mmForID *mRecordAccessorMockForID) Calls() []*RecordAccessorMockForIDParams {
	mmForID.mutex.RLock()

	argCopy := make([]*RecordAccessorMockForIDParams, len(mmForID.callArgs))
	copy(argCopy, mmForID.callArgs)

	mmForID.mutex.RUnlock()

	return argCopy
}

// MinimockForIDDone returns true if the count of the ForID invocations corresponds
// the number of defined expectations
func (m *RecordAccessorMock) MinimockForIDDone() bool {
	for _, e := range m.ForIDMock.expectations {
		if mm_atomic.LoadUint64(&e.Counter) < 1 {
			return false
		}
	}

	// if default expectation was set then invocations count should be greater than zero
	if m.ForIDMock.defaultExpectation != nil && mm_atomic.LoadUint64(&m.afterForIDCounter) < 1 {
		return false
	}
	// if func was set then invocations count should be greater than zero
	if m.funcForID != nil && mm_atomic.LoadUint64(&m.afterForIDCounter) < 1 {
		return false
	}
	return true
}

// MinimockForIDInspect logs each unmet expectation
func (m *RecordAccessorMock) MinimockForIDInspect() {
	for _, e := range m.ForIDMock.expectations {
		if mm_atomic.LoadUint64(&e.Counter) < 1 {
			m.t.Errorf("Expected call to RecordAccessorMock.ForID with params: %#v", *e.params)
		}
	}

	// if default expectation was set then invocations count should be greater than zero
	if m.ForIDMock.defaultExpectation != nil && mm_atomic.LoadUint64(&m.afterForIDCounter) < 1 {
		if m.ForIDMock.defaultExpectation.params == nil {
			m.t.Error("Expected call to RecordAccessorMock.ForID")
		} else {
			m.t.Errorf("Expected call to RecordAccessorMock.ForID with params: %#v", *m.ForIDMock.defaultExpectation.params)
		}
	}
	// if func was set then invocations count should be greater than zero
	if m.funcForID != nil && mm_atomic.LoadUint64(&m.afterForIDCounter) < 1 {
		m.t.Error("Expected call to RecordAccessorMock.ForID")
	}
}

// MinimockFinish checks that all mocked methods have been called the expected number of times
func (m *RecordAccessorMock) MinimockFinish() {
	if !m.minimockDone() {
		m.MinimockForIDInspect()
		m.t.FailNow()
	}
}

// MinimockWait waits for all mocked methods to be called the expected number of times
func (m *RecordAccessorMock) MinimockWait(timeout mm_time.Duration) {
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

func (m *RecordAccessorMock) minimockDone() bool {
	done := true
	return done &&
		m.MinimockForIDDone()
}
