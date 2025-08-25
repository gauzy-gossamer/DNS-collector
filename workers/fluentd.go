package workers

import (
	"crypto/tls"
	"strconv"
	"time"

	"github.com/IBM/fluent-forward-go/fluent/client"
	"github.com/IBM/fluent-forward-go/fluent/protocol"
	"github.com/dmachard/go-dnscollector/dnsutils"
	"github.com/dmachard/go-dnscollector/pkgconfig"
	"github.com/dmachard/go-dnscollector/transformers"
	"github.com/dmachard/go-logger"
	"github.com/dmachard/go-netutils"
)

type FluentdClient struct {
	*GenericWorker
	transport                          string
	fluentConn                         *client.Client
	transportReady, transportReconnect chan bool
	writerReady                        bool
}

func NewFluentdClient(config *pkgconfig.Config, logger *logger.Logger, name string) *FluentdClient {
	bufSize := config.Global.Worker.ChannelBufferSize
	if config.Loggers.Fluentd.ChannelBufferSize > 0 {
		bufSize = config.Loggers.Fluentd.ChannelBufferSize
	}
	w := &FluentdClient{GenericWorker: NewGenericWorker(config, logger, name, "fluentd", bufSize, pkgconfig.DefaultMonitor)}
	w.transportReady = make(chan bool)
	w.transportReconnect = make(chan bool)
	w.ReadConfig()
	return w
}

func (w *FluentdClient) ReadConfig() {
	w.transport = w.GetConfig().Loggers.Fluentd.Transport

	// begin backward compatibility
	if w.GetConfig().Loggers.Fluentd.TLSSupport {
		w.transport = netutils.SocketTLS
	}
	if len(w.GetConfig().Loggers.Fluentd.SockPath) > 0 {
		w.transport = netutils.SocketUnix
	}
}

func (w *FluentdClient) Disconnect() {
	if w.fluentConn != nil {
		w.LogInfo("closing fluentd connection")
		w.fluentConn.Disconnect()
	}
}

func (w *FluentdClient) ConnectToRemote() {
	for {
		if w.fluentConn != nil {
			w.fluentConn.Disconnect()
			w.fluentConn = nil
		}

		address := w.GetConfig().Loggers.Fluentd.RemoteAddress + ":" + strconv.Itoa(w.GetConfig().Loggers.Fluentd.RemotePort)
		connTimeout := time.Duration(w.GetConfig().Loggers.Fluentd.ConnectTimeout) * time.Second

		// make the connection
		var c *client.Client
		var err error

		switch w.transport {
		case netutils.SocketUnix:
			address = w.GetConfig().Loggers.Fluentd.RemoteAddress
			if len(w.GetConfig().Loggers.Fluentd.SockPath) > 0 {
				address = w.GetConfig().Loggers.Fluentd.SockPath
			}
			w.LogInfo("connecting to %s://%s", w.transport, address)
			c = client.New(client.ConnectionOptions{
				Factory: &client.ConnFactory{
					Network: "unix",
					Address: address,
				},
				ConnectionTimeout: connTimeout,
			})

		case netutils.SocketTCP:
			w.LogInfo("connecting to %s://%s", w.transport, address)
			c = client.New(client.ConnectionOptions{
				Factory: &client.ConnFactory{
					Network: "tcp",
					Address: address,
				},
				ConnectionTimeout: connTimeout,
			})

		case netutils.SocketTLS:
			w.LogInfo("connecting to %s://%s", w.transport, address)

			var tlsConfig *tls.Config

			tlsOptions := netutils.TLSOptions{
				InsecureSkipVerify: w.GetConfig().Loggers.Fluentd.TLSInsecure,
				MinVersion:         w.GetConfig().Loggers.Fluentd.TLSMinVersion,
				CAFile:             w.GetConfig().Loggers.Fluentd.CAFile,
				CertFile:           w.GetConfig().Loggers.Fluentd.CertFile,
				KeyFile:            w.GetConfig().Loggers.Fluentd.KeyFile,
			}
			tlsConfig, _ = netutils.TLSClientConfig(tlsOptions)

			c = client.New(client.ConnectionOptions{
				Factory: &client.ConnFactory{
					Network:   "tcp+tls",
					Address:   address,
					TLSConfig: tlsConfig,
				},
				ConnectionTimeout: connTimeout,
			})

		default:
			w.LogFatal("logger=fluent - invalid transport:", w.transport)
		}

		// something is wrong during connection ?
		err = c.Connect()
		if err != nil {
			w.LogError("connect error: %s", err)
			w.LogInfo("retry to connect in %d seconds", w.GetConfig().Loggers.Fluentd.RetryInterval)
			time.Sleep(time.Duration(w.GetConfig().Loggers.Fluentd.RetryInterval) * time.Second)
			continue
		}

		// save current connection
		w.fluentConn = c

		// block until framestream is ready
		w.transportReady <- true

		// block until an error occurred, need to reconnect
		w.transportReconnect <- true
	}
}

func (w *FluentdClient) FlushBuffer(buf *[]dnsutils.DNSMessage) {

	entries := []protocol.EntryExt{}

	for _, dm := range *buf {
		// Convert DNSMessage to map[]
		flatDm, _ := dm.Flatten()

		// get timestamp from DNSMessage
		timestamp, _ := time.Parse(time.RFC3339, dm.DNSTap.TimestampRFC3339)

		// append DNSMessage to the list
		entries = append(entries, protocol.EntryExt{
			Timestamp: protocol.EventTime{Time: timestamp},
			Record:    flatDm,
		})
	}

	// send all entries with tag, check error on write ?
	err := w.fluentConn.SendForward(w.GetConfig().Loggers.Fluentd.Tag, entries)
	if err != nil {
		w.LogError("forward fluent error", err.Error())
		w.writerReady = false
		<-w.transportReconnect
	}

	// reset buffer
	*buf = nil
}

func (w *FluentdClient) StartCollect() {
	w.LogInfo("starting data collection")
	defer w.CollectDone()

	// prepare next channels
	defaultRoutes, defaultNames := GetRoutes(w.GetDefaultRoutes())
	droppedRoutes, droppedNames := GetRoutes(w.GetDroppedRoutes())

	// prepare transforms
	subprocessors := transformers.NewTransforms(&w.GetConfig().OutgoingTransformers, w.GetLogger(), w.GetName(), w.GetOutputChannelAsList(), 0)

	// goroutine to process transformed dns messages
	go w.StartLogging()

	// init remote conn
	go w.ConnectToRemote()

	// loop to process incoming messages
	for {
		select {
		case <-w.OnStop():
			w.StopLogger()
			subprocessors.Reset()
			return

			// new config provided?
		case cfg := <-w.NewConfig():
			w.SetConfig(cfg)
			w.ReadConfig()
			subprocessors.ReloadConfig(&cfg.OutgoingTransformers)

		case dm, opened := <-w.GetInputChannel():
			if !opened {
				w.LogInfo("input channel closed!")
				return
			}
			// count global messages
			w.CountIngressTraffic()

			// apply transforms, init dns message with additional parts if necessary
			transformResult, err := subprocessors.ProcessMessage(&dm)
			if err != nil {
				w.LogError(err.Error())
			}
			if transformResult == transformers.ReturnDrop {
				w.SendDroppedTo(droppedRoutes, droppedNames, dm)
				continue
			}

			// send to output channel
			w.CountEgressTraffic()
			w.GetOutputChannel() <- dm

			// send to next ?
			w.SendForwardedTo(defaultRoutes, defaultNames, dm)
		}
	}
}

func (w *FluentdClient) StartLogging() {
	w.LogInfo("logging has started")
	defer w.LoggingDone()

	// init buffer
	bufferDm := []dnsutils.DNSMessage{}

	// init flush timer for buffer
	flushInterval := time.Duration(w.GetConfig().Loggers.Fluentd.FlushInterval) * time.Second
	flushTimer := time.NewTimer(flushInterval)

	for {
		select {
		case <-w.OnLoggerStopped():
			return

		case <-w.transportReady:
			w.LogInfo("connected with remote side")
			w.writerReady = true

			// incoming dns message to process
		case dm, opened := <-w.GetOutputChannel():
			if !opened {
				w.LogInfo("output channel closed!")
				return
			}

			// drop dns message if the connection is not ready to avoid memory leak or
			// to block the channel
			if !w.writerReady {
				continue
			}

			// append dns message to buffer
			bufferDm = append(bufferDm, dm)

			// buffer is full ?
			if len(bufferDm) >= w.GetConfig().Loggers.Fluentd.BufferSize {
				w.FlushBuffer(&bufferDm)
			}

		// flush the buffer
		case <-flushTimer.C:
			if !w.writerReady {
				bufferDm = nil
			}

			if len(bufferDm) > 0 {
				w.FlushBuffer(&bufferDm)
			}

			// restart timer
			flushTimer.Reset(flushInterval)
		}
	}
}
