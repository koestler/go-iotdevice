package mongo

import (
	"github.com/koestler/go-ve-sensor/vedata"
	"gopkg.in/mgo.v2"
	"log"
	"time"
)

func GetSession(mongoHost string) *mgo.Session {
	session, err := mgo.Dial("mongodb://" + mongoHost)

	session.SetMode(mgo.Monotonic, true)

	if err != nil {
		panic(err)
	}
	return session
}

func ensureIndexAndIgnoreError(collection *mgo.Collection, index mgo.Index) {
	err := collection.EnsureIndex(index)

	if err != nil {
		log.Printf("mongo: index creation failed: %v", err)
	}

}

func Run(session *mgo.Session, databaseName string, rawValuesIntervall int) {
	collection := session.DB(databaseName).C("RawValues")

	ensureIndexAndIgnoreError(collection, mgo.Index{
		Key:        []string{"type"},
		Unique:     false,
		DropDups:   false,
		Background: true,
		Sparse:     true,
	})

	ensureIndexAndIgnoreError(collection, mgo.Index{
		Key:        []string{"name", "updated"},
		Unique:     false,
		DropDups:   false,
		Background: true,
		Sparse:     true,
	})

	go func() {
		for _ = range time.Tick(time.Duration(rawValuesIntervall) * time.Millisecond) {
			for _, deviceId := range vedata.ReadDeviceIds() {
				device, err := deviceId.ReadDevice()
				if err != nil {
					log.Print("mongo: cannot read device $v: %v", deviceId, err)
					continue
				}

				if err := collection.Insert(device); err != nil {
					log.Print("mongo: cannot insert: %v", err)
				}
			}
		}
	}()
}
