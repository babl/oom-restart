package main

import (
	"encoding/gob"
	_ "fmt"
	"os"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/larskluge/babl-server/kafka"
	. "github.com/larskluge/babl-server/utils"
	bn "github.com/larskluge/babl/bablnaming"
	pb "github.com/larskluge/babl/protobuf/messages"
	_ "github.com/muesli/cache2go"
	cache "github.com/patrickmn/go-cache"
)

func getCache() *cache.Cache {
	fp, err := os.Open(cacheDir)
	defer fp.Close()
	if err != nil {
		return cache.New(1*time.Minute, 15*time.Second)
	} else {
		dec := gob.NewDecoder(fp)
		items := map[string]cache.Item{}
		err := dec.Decode(&items)
		Check(err)
		return cache.NewFrom(1*time.Minute, 15*time.Second, items)
	}
}

func saveCache(c *cache.Cache) {
	items := map[string]cache.Item{}
	fp, err := os.Create(cacheDir)
	defer fp.Close()
	enc := gob.NewEncoder(fp)
	items = c.Items()
	err = enc.Encode(&items)
	Check(err)
}

func IncrementOom(id string) int {
	c := getCache()
	var n int
	var err error

	_, found := c.Get(id)
	if found {
		n, err = c.IncrementInt(id, 1)
		Check(err)
	} else {
		c.Set(id, 1, OomTimeout)
		n = 1
	}
	saveCache(c)
	return n
}

func (s *server) BroadcastModuleRestartRequest(m *RestartRequest) error {
	dr := pb.RestartRequest{InstanceId: m.InstanceId}
	req := pb.Meta{
		Restart: &dr,
	}
	msg, err := proto.Marshal(&req)
	if err != nil {
		return err
	}

	topic := bn.ModuleToTopic(m.Module, true)
	kafka.SendMessage(s.kafkaProducer, "", topic, &msg) // TODO return err
	return nil
}
