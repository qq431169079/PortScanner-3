package server

import (
	"time"

	"github.com/darkMoon1973/PortScanner/common/lib/go-nmap"
	"github.com/darkMoon1973/PortScanner/common/util"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type MongoDriver struct {
	Url     string
	DbName  string
	Session *mgo.Session
}

func (m *MongoDriver) Init() {
	var err error
	m.Session, err = mgo.Dial(m.Url)
	if err != nil {
		panic(err)
	}
	m.Session.SetMode(mgo.Monotonic, true)
	m.Session.SetPoolLimit(1024)
}

func (m *MongoDriver) GetSession() *mgo.Session {
	if m.Session == nil {
		m.Init()
	}
	if m.Session.Ping() != nil {
		m.Session.Refresh()
	}
	return m.Session.Clone()
}

func upsert(table string, scanList ipInfo) (err error) {
	s := logic.MongoDriver.GetSession()
	defer s.Close()
	_, err = s.DB(logic.MongoDriver.DbName).C(table).Upsert(bson.M{"ip": scanList.Ip}, scanList)
	return err
}

func get(table string) (result []ipInfo, err error) {
	s := logic.MongoDriver.GetSession()
	defer s.Close()
	err = s.DB(logic.MongoDriver.DbName).C(table).Find(bson.M{}).All(&result)
	return result, err
}

func resultToMongo(result nmap.Result) error {
	timeStamp := time.Now().Unix()
	result.TimeStamp = timeStamp
	result.ScanDate = util.TimeToDate(timeStamp)
	s := logic.MongoDriver.GetSession()
	_, err := s.DB(logic.MongoDriver.DbName).C("scan_result").Upsert(bson.M{
		"ip":              result.IP,
		"port":            result.Port,
		"protocol":        result.Protocol,
		"service.name":    result.Service.Name,
		"service.product": result.Service.Product,
		"service.version": result.Service.Version,
		"service.info":    result.Service.Info,
	}, result)
	s.Close()
	if err != nil {
		logic.Log.Error("Upsert to mongodb fail, ", err)
		return nil
	}
	return nil
}
