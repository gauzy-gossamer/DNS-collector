
# DNS-collector - Development

To compile DNS-collector, we assume you have a working Go setup.
First, make sure your golang version is `1.21` or higher

How to userguides:

- [Add a new worker](#add-a-worker-collector-or-logger)
- [Add a new transform](#add-transformer)

## Build and run from source

Building from source

- The very fast way, go to the top of the project and run go command

```bash
go run .
```

- Uses the `MakeFile` (prefered way)

```bash
make
```

Execute the binary

```bash
make run
```

- From the `DockerFile`

## Run linters

Install linter

```bash
sudo apt install build-essential
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
```

Execute linter before to commit

```bash
make lint
```

## Run test units

Execute testunits before to commit

```bash
sudo make tests
```

Execute a test for one specific testcase in a package

```bash
go test -timeout 10s -cover -v ./workers -run Test_SyslogRun
```

Run bench

```bash
cd dnsutils/
go test -run=^$ -bench=.
```

## Update Golang version and package dependencies

Update go version

```bash
go mod edit -go=1.24
```

Update package dependencies

```bash
make dep
```

## How to userguides

### Add transformer

Add Configuration `dnsutils/config.go` and `config.yml`

```golang
type ConfigTransformers struct {
 MyTransform struct {
  Enable         bool `yaml:"enable"`
    }
}
```

```golang
func (c *ConfigTransformers) SetDefault() {
    c.MyTransform.Enable = false
}
```

Create the following file `transformers/mytransform.go` and `transformers/mytransform_test.go`

```golang
type MyTransform struct {
	GenericTransformer
}

func MyTransform(config *pkgconfig.ConfigTransformers, logger *logger.Logger, name string, instance int, nextWorkers []chan dnsutils.DNSMessage) *MyTransform {
	t := &MyTransform{GenericTransformer: NewTransformer(config, logger, "mytransform", name, instance, nextWorkers)}
	return t
}
```

Declare the transfomer in the following file `tranformers.go`
Finally update the docs `doc/transformers.md` and `README.md`

### Add a worker (collector or logger)

1. Add Configuration in `pkgconfig/logger.go`  or `pkgconfig/collectors.go`

```golang
Loggers struct {
    MyLogger struct {
        Enable   bool   `yaml:"enable"`
    }
}
```

```golang
func (c *Config) SetDefault() {
    c.Loggers.MyLogger.Enable = false
}
```

2. Create the following file `workers/mylogger.go` and `loggers/mylogger_test.go`

```golang
package workers

import (
	"github.com/dmachard/go-dnscollector/pkgconfig"
	"github.com/dmachard/go-dnscollector/pkgutils"
	"github.com/dmachard/go-logger"
)

type MyWorker struct {
	*GenericWorker
}

func NewMyWorker(config *pkgconfig.Config, console *logger.Logger, name string) *MyWorker {
	s := &MyWorker{GenericWorker: NewGenericWorker(config, console, name, "worker", DefaultBufferSize)}
	s.ReadConfig()
	return s
}

func (w *DevNull) StartCollect() {
	w.LogInfo("starting data collection")
	defer w.CollectDone()

	// goroutine to process transformed dns messages
	go w.StartLogging()

	// loop to process incoming messages
	for {
		select {
		case <-w.OnStop():
			w.StopLogger()

		case _, opened := <-w.GetInputChannel():
			if !opened {
				w.LogInfo("run: input channel closed!")
				return
			}
		}
	}
}

func (w *DevNull) StartLogging() {
	w.LogInfo("logging has started")
	defer w.LoggingDone()

	for {
		select {
		case <-w.OnLoggerStopped():
			return

		case _, opened := <-w.GetOutputChannel():
			if !opened {
				w.LogInfo("process: output channel closed!")
				return
			}

		}
	}
}
```

3. Update the main file `pkglinker` in `pipelines.go`

```golang
if subcfg.Loggers.MyLogger.Enable && IsLoggerRouted(config, output.Name) {
    mapLoggers[output.Name] = loggers.NewMyLogger(subcfg, logger, output.Name)
}
```

4. Finally update the docs `doc/loggers.md` or `doc/collectors.md` and `README.md`