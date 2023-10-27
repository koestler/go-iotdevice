package dataflow_test

import (
	"context"
	"fmt"
	"github.com/koestler/go-iotdevice/dataflow"
	mock_dataflow "github.com/koestler/go-iotdevice/dataflow/mock"
	"go.uber.org/mock/gomock"
	"reflect"
	"sort"
	"sync"
	"testing"
)

func TestRegisterDb(t *testing.T) {
	const numbPopulate = 21
	const numbGet = 18
	const numbSubscriptions = 32

	rdb := dataflow.NewRegisterDb()

	getCtx, getCancel := context.WithCancel(context.Background())

	// compute expected output
	expect := []string{"A0", "A1", "A2"}
	for i := 0; i < numbPopulate; i++ {
		expect = append(expect, []string{
			fmt.Sprintf("B%d", i),
			fmt.Sprintf("C%d", i),
			fmt.Sprintf("D%d", i),
		}...)
	}
	sort.Strings(expect)

	// populate storage
	rdb.Add(mockRegister(t, "A0"), mockRegister(t, "A1"), mockRegister(t, "A2"))
	wgPopulate := sync.WaitGroup{}
	wgPopulate.Add(numbPopulate)
	for i := 0; i < numbPopulate; i++ {
		i := i
		go func() {
			defer wgPopulate.Done()
			rdb.Add(
				mockRegister(t, fmt.Sprintf("B%d", i)),
				mockRegister(t, fmt.Sprintf("C%d", i)),
				mockRegister(t, fmt.Sprintf("D%d", i)),
			)
		}()
	}

	// call getAll at the same time as populating
	for i := 0; i < numbGet; i++ {
		// call GetAll during populate
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
	wgSubscribe.Add(numbSubscriptions)
	for i := 0; i < numbSubscriptions; i++ {
		i := i
		s := rdb.Subscribe(subscribeCtx, dataflow.AllRegisterFilter)
		go func() {
			defer wgSubscribe.Done()
			got := make([]string, 0)
			for o := range s {
				got = append(got, o.Name())
			}
			sort.Strings(got)
			if !reflect.DeepEqual(expect, got) {
				t.Errorf("subscription %d: expect \n%v but got \n%v", i, expect, got)
			}
		}()

	} // wait for storage to be fully populated
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
	sort.Strings(ret)
	return ret
}

func mockRegister(t *testing.T, name string) dataflow.Register {
	ctrl := gomock.NewController(t)
	m := mock_dataflow.NewMockRegister(ctrl)
	m.EXPECT().Category().Return("")
	m.EXPECT().Name().Return(name)
	m.EXPECT().Description().Return("")
	m.EXPECT().RegisterType().Return(dataflow.NumberRegister)
	m.EXPECT().Enum().Return(nil)
	m.EXPECT().Unit().Return("")
	m.EXPECT().Sort().Return(0)
	m.EXPECT().Controllable().Return(false)

	return m
}
