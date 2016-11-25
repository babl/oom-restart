package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/Shopify/sarama"
	log "github.com/Sirupsen/logrus"
	"github.com/larskluge/babl-server/kafka"
	. "github.com/larskluge/babl-server/utils"
)

const (
	Version    = "0.0.1"
	OomTimeout = 60 * time.Second
	MAX_OOM    = 3
	cacheDir   = "./cache"
)

type RestartRequest struct {
	Brokers    []string
	InstanceId string
	Module     string
}

type server struct {
	kafkaProducer *sarama.SyncProducer
}

func main() {
	log.SetOutput(os.Stderr)
	log.SetFormatter(&log.JSONFormatter{})

	m := RestartRequest{Brokers: strings.Split(os.Getenv("BABL_KAFKA_BROKERS"), ","), InstanceId: os.Getenv("INSTANCE_ID"), Module: os.Getenv("MODULE")}
	if m.InstanceId == "" || m.Module == "" || len(m.Brokers) == 0 {
		log.WithFields(log.Fields{"InstanceId": m.InstanceId, "module": m.Module, "brokers": m.Brokers}).Error("Missing Data")
		os.Exit(101)
	}
	ParseEvents(&m)
}

func ParseEvents(m *RestartRequest) {
	n := IncrementOom(m.InstanceId)
	if n >= MAX_OOM {
		fmt.Println("Module Restart Sent", m.InstanceId, n, time.Now())
		broadcastInstanceRestart(m)
		notify(m)
	} else {
		fmt.Println("Module OOM Counter", m.InstanceId, n, time.Now())
	}
}

func broadcastInstanceRestart(m *RestartRequest) {
	s := server{}
	s.kafkaProducer = kafka.NewProducer(m.Brokers, "oom-restart")
	defer (*s.kafkaProducer).Close()
	s.BroadcastModuleRestartRequest(m)
}

func notify(m *RestartRequest) {
	Cluster := strings.Index(m.Brokers[0], ".")
	str := fmt.Sprintf("[%s] %s(%s) will restart... gracefully!", Cluster, m.Module, m.InstanceId[0:12])
	args := []string{"-c", "sandbox.babl.sh:4445", "babl/events", "-e", "EVENT=babl:error"}
	cmd := exec.Command("/bin/babl", args...)
	cmd.Stdin = strings.NewReader(str)
	err := cmd.Run()
	Check(err)
}
