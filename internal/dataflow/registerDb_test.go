package dataflow_test

import (
	"context"
	"fmt"
	"github.com/koestler/go-iotdevice/v3/internal/dataflow"
	mock_dataflow "github.com/koestler/go-iotdevice/v3/internal/dataflow/mock"
	"reflect"
	"slices"
	"sync"
	"testing"
)

func TestRegisterDb(t *testing.T) {
	const numPopulate = 21
	const numGet = 18
	const numSubscriptions = 32

	rdb := dataflow.NewRegisterDb()

	getCtx, getCancel := context.WithCancel(context.Background())

	// compute expected output
	expect := []string{"A0", "A1", "A2"}
	for i := 0; i < numPopulate; i++ {
		expect = append(expect, []string{
			fmt.Sprintf("B%d", i),
			fmt.Sprintf("C%d", i),
			fmt.Sprintf("D%d", i),
		}...)
	}
	slices.Sort(expect)

	// populate storage
	rdb.Add(mockRegister("A0"), mockRegister("A1"), mockRegister("A2"))
	wgPopulate := sync.WaitGroup{}
	wgPopulate.Add(numPopulate)
	for i := 0; i < numPopulate; i++ {
		go func() {
			defer wgPopulate.Done()
			rdb.Add(
				mockRegister(fmt.Sprintf("B%d", i)),
				mockRegister(fmt.Sprintf("C%d", i)),
				mockRegister(fmt.Sprintf("D%d", i)),
			)
		}()
	}

	for i := 0; i < numGet; i++ {
		// call GetAll at the same time as populating
		go func() {
			for {
				select {
				case <-getCtx.Done():
					return
				default:
					rdb.GetAll()
				}
			}
		}()
	}

	// create subscriptions and check output
	subscribeCtx, subscribeCancel := context.WithCancel(context.Background())
	wgSubscribe := sync.WaitGroup{}
	wgSubscribe.Add(numSubscriptions)
	for i := 0; i < numSubscriptions; i++ {
		s := rdb.Subscribe(subscribeCtx, func(dataflow.Filterable) bool { return true })
		go func() {
			defer wgSubscribe.Done()
			got := make([]string, 0)
			for o := range s {
				got = append(got, o.Name())
			}
			slices.Sort(got)
			if !reflect.DeepEqual(expect, got) {
				t.Errorf("subscription %d: expect \n%v but got \n%v", i, expect, got)
			}
		}()

	}

	// wait for storage to be fully populated
	wgPopulate.Wait()

	// check output of getAll
	t.Run("GetAll", func(t *testing.T) {
		if got := nameSlice(rdb.GetAll()); !reflect.DeepEqual(expect, got) {
			t.Errorf("expect %v but got %v", expect, got)
		}
	})

	t.Run("GetByName", func(t *testing.T) {
		if _, ok := rdb.GetByName("non-existent"); ok {
			t.Errorf("expect GetByName to return no result")
		}

		if got, ok := rdb.GetByName("A0"); !ok {
			t.Errorf("expect GetByName to return ok")
		} else if expect, got := "A0", got.Name(); expect != got {
			t.Errorf("expect Register Name to be %s but got %s", expect, got)
		}
	})

	getCancel()
	subscribeCancel()
	wgSubscribe.Wait()
}

func nameSlice(list []dataflow.RegisterStruct) []string {
	ret := make([]string, len(list))
	for i, r := range list {
		ret[i] = r.Name()
	}
	slices.Sort(ret)
	return ret
}

func mockRegister(name string) dataflow.Register {
	return &mock_dataflow.MockRegister{NameV: name, RegisterTypeV: dataflow.NumberRegister}
}
