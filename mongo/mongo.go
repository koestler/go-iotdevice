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

func Run(session *mgo.Session, databaseName string, rawValuesIntervall int) {
	collection := session.DB(databaseName).C("RawValues")

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
				} else {
					log.Printf("mongo: write ...")
				}
			}

		}
	}()
}
