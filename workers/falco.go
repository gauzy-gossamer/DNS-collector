package workers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/dmachard/go-dnscollector/pkgconfig"
	"github.com/dmachard/go-dnscollector/transformers"
	"github.com/dmachard/go-logger"
)

type FalcoClient struct {
	*GenericWorker
}

func NewFalcoClient(config *pkgconfig.Config, console *logger.Logger, name string) *FalcoClient {
	bufSize := config.Global.Worker.ChannelBufferSize
	if config.Loggers.FalcoClient.ChannelBufferSize > 0 {
		bufSize = config.Loggers.FalcoClient.ChannelBufferSize
	}
	w := &FalcoClient{GenericWorker: NewGenericWorker(config, console, name, "falco", bufSize, pkgconfig.DefaultMonitor)}
	w.ReadConfig()
	return w
}

func (w *FalcoClient) StartCollect() {
	w.LogInfo("starting data collection")
	defer w.CollectDone()

	// prepare next channels
	defaultRoutes, defaultNames := GetRoutes(w.GetDefaultRoutes())
	droppedRoutes, droppedNames := GetRoutes(w.GetDroppedRoutes())

	// prepare transforms
	subprocessors := transformers.NewTransforms(&w.GetConfig().OutgoingTransformers, w.GetLogger(), w.GetName(), w.GetOutputChannelAsList(), 0)

	// goroutine to process transformed dns messages
	go w.StartLogging()

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

func (w *FalcoClient) StartLogging() {
	w.LogInfo("logging has started")
	defer w.LoggingDone()

	buffer := new(bytes.Buffer)

	for {
		select {
		case <-w.OnLoggerStopped():
			return

		// incoming dns message to process
		case dm, opened := <-w.GetOutputChannel():
			if !opened {
				w.LogInfo("output channel closed!")
				return
			}

			// encode
			json.NewEncoder(buffer).Encode(dm)

			req, _ := http.NewRequest("POST", w.GetConfig().Loggers.FalcoClient.URL, buffer)
			req.Header.Set("Content-Type", "application/json")
			client := &http.Client{
				Timeout: 5 * time.Second,
			}
			_, err := client.Do(req)
			if err != nil {
				w.LogError(err.Error())
			}

			// finally reset the buffer for next iter
			buffer.Reset()
		}
	}
}
