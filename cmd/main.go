package main // import "github.com/dbsystel/kibana-config-controller"

import (
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/dbsystel/kibana-config-controller/controller"
	"github.com/dbsystel/kibana-config-controller/kibana"
	"github.com/dbsystel/kube-controller-dbsystel-go-common/controller/configmap"
	"github.com/dbsystel/kube-controller-dbsystel-go-common/kubernetes"
	k8sflag "github.com/dbsystel/kube-controller-dbsystel-go-common/kubernetes/flag"
	opslog "github.com/dbsystel/kube-controller-dbsystel-go-common/log"
	logflag "github.com/dbsystel/kube-controller-dbsystel-go-common/log/flag"
	"github.com/go-kit/kit/log/level"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	app = kingpin.New(filepath.Base(os.Args[0]), "Kibana Controller")
	//Here you can define more flags for your application
	kibanaURL = app.Flag("kibana-url", "The url to issue requests to update dashboards to.").Required().String()
	id        = app.Flag("id", "The kibana id to issue requests to update dashboards to.").Default("0").Int()
	namespace = app.Flag("namespace", "The namespace to watching.").Default("").String()
)

func main() {
	//Define config for logging
	var logcfg opslog.Config
	//Definie if controller runs outside of k8s
	var runOutsideCluster bool
	//Add two additional flags to application for logging and decision if inside or outside k8s
	logflag.AddFlags(app, &logcfg)
	k8sflag.AddFlags(app, &runOutsideCluster)
	//Parse all arguments
	_, err := app.Parse(os.Args[1:])
	if err != nil {
		//Received error while parsing arguments from function app.Parse
		fmt.Fprintln(os.Stderr, "Catched the following error while parsing arguments: ", err)
		app.Usage(os.Args[1:])
		os.Exit(2)
	}

	//Initialize new logger from opslog
	logger, err := opslog.New(logcfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		app.Usage(os.Args[1:])
		os.Exit(2)
	}
	//First usage of initialized logger for testing
	// nolint:errcheck
	level.Debug(logger).Log("msg", "Logging initiated...")
	//Initialize new k8s client from common k8s package
	k8sClient, err := kubernetes.NewClientSet(runOutsideCluster)
	if err != nil {
		//nolint:errcheck
		level.Error(logger).Log("msg", err.Error())
		os.Exit(2)
	}

	uRL, err := url.Parse(*kibanaURL)
	if err != nil {
		//nolint:errcheck
		level.Error(logger).Log("msg", "Kibana URL could not be parsed: "+*kibanaURL)
		os.Exit(2)
	}

	k := kibana.New(uRL, *id, logger)

	sigs := make(chan os.Signal, 1) // Create channel to receive OS signals
	stop := make(chan struct{})     // Create channel to receive stop signal

	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM, syscall.SIGINT) // Register the sigs channel to receieve SIGTERM

	wg := &sync.WaitGroup{} // Goroutines can add themselves to this to be waited on so that they finish

	//Initialize new k8s configmap-controller from common k8s package
	configMapController := &configmap.ConfigMapController{}
	configMapController.Controller = controller.New(*k, logger)
	configMapController.Initialize(k8sClient, *namespace)
	//Run initiated configmap-controller as go routine
	go configMapController.Run(stop, wg)

	<-sigs // Wait for signals (this hangs until a signal arrives)

	//nolint:errcheck
	level.Info(logger).Log("msg", "Shutting down...")

	close(stop) // Tell goroutines to stop themselves
	wg.Wait()   // Wait for all to be stopped
}
