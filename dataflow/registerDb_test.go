package dataflow

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"sync"
	"testing"
)

func TestRegisterDb(t *testing.T) {
	const numbPopulate = 21
	const numbGet = 18
	const numbSubscriptions = 32

	rdb := NewRegisterDb()

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
	rdb.Add(mockRegister{name: "A0"}, mockRegister{name: "A1"}, mockRegister{name: "A2"})
	wgPopulate := sync.WaitGroup{}
	wgPopulate.Add(numbPopulate)
	for i := 0; i < numbPopulate; i++ {
		i := i
		go func() {
			defer wgPopulate.Done()
			rdb.Add(
				mockRegister{name: fmt.Sprintf("B%d", i)},
				mockRegister{name: fmt.Sprintf("C%d", i)},
				mockRegister{name: fmt.Sprintf("D%d", i)},
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
		s := rdb.Subscribe(subscribeCtx)
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
		if reg := rdb.GetByName("non-existent"); reg != nil {
			t.Errorf("expect GetByName to return no result")
		}

		if got := rdb.GetByName("A0"); got == nil {
			t.Errorf("expect GetByName to return ok")
		} else if expect, got := "A0", got.Name(); expect != got {
			t.Errorf("expect Register Name to be %s but got %s", expect, got)
		}
	})

	getCancel()
	subscribeCancel()
	wgSubscribe.Wait()
}

func nameSlice(list []Register) []string {
	ret := make([]string, len(list))
	for i, r := range list {
		ret[i] = r.Name()
	}
	sort.Strings(ret)
	return ret
}

type mockRegister struct {
	name string
}

func (r mockRegister) Category() string {
	return ""
}

func (r mockRegister) Name() string {
	return r.name
}

func (r mockRegister) Description() string {
	return ""
}

func (r mockRegister) RegisterType() RegisterType {
	return NumberRegister
}

func (r mockRegister) Enum() map[int]string {
	return nil
}

func (r mockRegister) Unit() string {
	return ""
}

func (r mockRegister) Sort() int {
	return 0
}

func (r mockRegister) Controllable() bool {
	return false
}
