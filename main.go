package main

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	natsServer "github.com/nats-io/nats-server/v2/server"
	natsClient "github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

func NewNatsServer(port int) (*natsServer.Server, *natsClient.Conn, error) {
	sysAcct := natsServer.NewAccount("$SYS")
	appAcct := natsServer.NewAccount("app")

	routes := []*url.URL{
		{
			Scheme: "nats",
			Host:   "localhost:5222",
			User:   url.UserPassword("foo", "bar"),
		},
		{
			Scheme: "nats",
			Host:   "localhost:5223",
			User:   url.UserPassword("foo", "bar"),
		},
		{
			Scheme: "nats",
			Host:   "localhost:5224",
			User:   url.UserPassword("foo", "bar"),
		},
	}

	opts := &natsServer.Options{
		ServerName: fmt.Sprintf("server-%d", port),
		StoreDir:   fmt.Sprintf("./data/jetstream/%d", port),
		JetStream:  true,
		Routes:     routes,
		Port:       port,
		Cluster: natsServer.ClusterOpts{
			Name:      "in-cluster-embeded",
			Port:      port + 1000,
			Host:      "0.0.0.0",
			Advertise: fmt.Sprintf("localhost:%d", port+1000),
		},
		SystemAccount:          "$SYS",
		DisableJetStreamBanner: true,
		Users: []*natsServer.User{
			{
				Username: "thomas",
				Password: "thomas",
				Account:  sysAcct,
			},
			{
				Username: "app",
				Password: "app",
				Account:  appAcct,
			},
		},
		Accounts: []*natsServer.Account{sysAcct, appAcct},
	}

	ns, err := natsServer.NewServer(opts)
	if err != nil {
		return nil, nil, err
	}

	ns.ConfigureLogger()

	ns.Start()

	appAcct, err = ns.LookupAccount("app")
	if err != nil {
		return nil, nil, err
	}

	err = appAcct.EnableJetStream(map[string]natsServer.JetStreamAccountLimits{
		"": {
			MaxMemory:    1024 * 1024 * 1024,
			MaxStore:     -1,
			MaxStreams:   512,
			MaxConsumers: 512,
		},
	})

	if err != nil {
		return nil, nil, err
	}

	if !ns.ReadyForConnections(10 * time.Second) {
		panic("server failed to be ready for connection")
	}

	nc, err := natsClient.Connect(
		fmt.Sprintf("nats://localhost:%d", port),
		natsClient.UserInfo("app", "app"),
	)
	if err != nil {
		return nil, nil, err
	}

	go TrySetupStreams(nc)

	return ns, nc, nil
}

func TrySetupStreams(nc *natsClient.Conn) {
	maxAttempts := 10
	for i := 0; i < maxAttempts; i++ {
		fmt.Println("attempting to create stream `some-stream`")
		js, err := jetstream.New(nc)
		if err != nil {
			panic(err)
		}

		_, err = js.CreateOrUpdateStream(context.Background(), jetstream.StreamConfig{
			Name:      "some-stream",
			Subjects:  []string{"some-stream.>"},
			Retention: jetstream.LimitsPolicy,
			Discard:   jetstream.DiscardOld,
			MaxAge:    time.Hour * 24 * 7,
			MaxBytes:  1024 * 1024 * 1024,
			Replicas:  3,
		})

		if err != nil {
			fmt.Println("failed to create stream: ", err)
		} else {
			fmt.Println("successfully created streams")
			return
		}

		time.Sleep(time.Second * 10)
	}
}

var (
	flagPort int
)

func init() {
	flag.IntVar(&flagPort, "port", 4442, "port to use")
}

func main() {
	flag.Parse()

	_, _, err := NewNatsServer(flagPort)
	if err != nil {
		panic(err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT, syscall.SIGSEGV)
	for {
		s := <-c
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			return
		default:
			return
		}
	}
}
