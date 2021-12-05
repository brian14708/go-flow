package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/NYTimes/gziphandler"
	"github.com/sirupsen/logrus"

	"github.com/brian14708/go-flow/flowdebug"
	"github.com/brian14708/go-flow/flowdebug/types"
)

var (
	flagFrontendAddr = flag.String("frontend", ":6063", "Frontend listen address")
	flagBackendAddr  = flag.String("backend", ":6062", "Backend listen address")
)

type subscriber struct {
	id        string
	timestamp time.Time
	handler   func(*flowdebug.StatsMessage)
}

type Server struct {
	addr   string
	ctx    context.Context
	cancel context.CancelFunc
	prof   *flowdebug.Profiler
}

func NewServer(addr string, p *flowdebug.Profiler) *Server {
	ctx, cancel := context.WithCancel(context.Background())

	return &Server{
		addr:   addr,
		ctx:    ctx,
		cancel: cancel,
		prof:   p,
	}
}

func (s *Server) Stop() {
	s.cancel()
}

func (s *Server) Start() {
	ch := make(chan *subscriber)
	go s.startGraphServer(ch)
	go s.startStatsServer(ch)
}

func (s *Server) startGraphServer(ch chan<- *subscriber) {
	l, err := net.Listen("tcp", s.addr)
	if err != nil {
		panic(err)
	}
	defer l.Close()

	subscribeHandler := func(m map[string]flowdebug.Metric) func(*flowdebug.StatsMessage) {
		return func(msg *flowdebug.StatsMessage) {
			for i := range msg.TagMsgs {
				tagMsg := msg.TagMsgs[i]
				key := fmt.Sprintf("%s/%s", msg.StatId, tagMsg.Tag)
				_, ok := m[key]
				if ok {
					m[key].Store(int32(tagMsg.Value))
				}
			}
		}
	}

	for {
		select {
		case <-s.ctx.Done():
			return
		default:
		}
		conn, err := l.Accept()
		if err != nil {
			panic(err)
		}
		go func() {
			defer conn.Close()
			reader := bufio.NewReader(conn)
			buf, err := reader.ReadBytes('$')
			if err != nil {
				logrus.Errorln(err)
				return
			}

			var t types.Topology
			err = json.Unmarshal(buf[:len(buf)-1], &t)
			if err != nil {
				logrus.Errorln(err)
				return
			}

			log.Printf("graph register %s", t.ID)

			var prefix []string
			m := make(map[string]flowdebug.Metric)
			for _, c := range t.Connections {
				if len(c.Metrics) > 0 {
					prefix = append(prefix, c.ID)
				}
				for _, metric := range c.Metrics {
					key := fmt.Sprintf("%s/%s", c.ID, metric)
					m[key] = s.prof.AddMetric(key, false)
				}
			}

			for _, node := range t.Nodes {
				if len(node.Metrics) > 0 {
					prefix = append(prefix, node.Name)
				}
				for _, metric := range node.Metrics {
					key := fmt.Sprintf("%s/%s", node.Name, metric)
					m[key] = s.prof.AddMetric(key, false)
				}
			}

			msg := &subscriber{
				id:      t.ID,
				handler: subscribeHandler(m),
			}

			err = s.prof.AddGraph(t.ID, &t, prefix)
			if err != nil {
				logrus.Errorln(err)
				return
			}

			ch <- msg
		}()
	}
}

func (s *Server) startStatsServer(ch <-chan *subscriber) {
	l, err := net.ListenPacket("udp", s.addr)
	if err != nil {
		panic(err)
	}
	defer l.Close()

	buffer := make([]byte, 1024)

	subs := make(map[string]*subscriber)
	timeout := 60.0

	prev := time.Now()
	for {
		select {
		case <-s.ctx.Done():
			return
		case sub := <-ch:
			sub.timestamp = time.Now()
			subs[sub.id] = sub
		default:
			err := l.SetReadDeadline(time.Now().Add(1 * time.Second))
			if err != nil {
				continue
			}
			n, _, err := l.ReadFrom(buffer)
			if err == nil {
				// id.stat;[tag=value|metric@ratio;]
				err, msg := flowdebug.RegexStatsMessage(string(buffer[:n]))
				if err != nil {
					continue
				}
				sub, ok := subs[msg.GraphId]
				if ok {
					sub.handler(msg)
					sub.timestamp = time.Now()
				}
			}
		}

		if time.Since(prev).Seconds() < timeout {
			continue
		}

		prev = time.Now()
		for id, sub := range subs {
			if time.Since(sub.timestamp).Seconds() > timeout {
				delete(subs, id)
				log.Printf("graph unregister %s", id)
			}
		}
	}
}

func main() {
	flag.Parse()

	prof := flowdebug.NewProfiler(&flowdebug.ProfilerOptions{
		SampleRate: time.Second,
	})
	s := NewServer(*flagBackendAddr, prof)
	s.Start()

	http.Handle("/", gziphandler.GzipHandler(s.prof))
	log.Println(http.ListenAndServe(*flagFrontendAddr, nil))
}
