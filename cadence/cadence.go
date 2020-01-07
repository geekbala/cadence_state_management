package cadence

import (
	"cadence_helloworld/workflow"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/uber-go/tally"
	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	"go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/client"
	"go.uber.org/cadence/worker"
	"go.uber.org/yarpc"
	"go.uber.org/yarpc/transport/tchannel"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var hostPort = "192.168.99.101:7933"
var domain = "metering"
var taskListName = "cadence-helloworld-task"
var clientName = "cadence-helloworld-client"
var cadenceService = "cadence-frontend"

func buildLogger() *zap.Logger {
	config := zap.NewDevelopmentConfig()
	config.Level.SetLevel(zapcore.InfoLevel)

	var err error
	logger, err := config.Build()
	if err != nil {
		panic("Failed to setup logger")
	}

	return logger
}

func buildWorkflowServiceClient() workflowserviceclient.Interface {
	ch, err := tchannel.NewChannelTransport(tchannel.ServiceName(clientName))
	if err != nil {
		panic("Failed to setup tchannel")
	}
	dispatcher := yarpc.NewDispatcher(yarpc.Config{
		Name: clientName,
		Outbounds: yarpc.Outbounds{
			cadenceService: {Unary: ch.NewSingleOutbound(hostPort)},
		},
	})
	if err := dispatcher.Start(); err != nil {
		fmt.Println(err)
		panic("Failed to start dispatcher")
	}

	return workflowserviceclient.New(dispatcher.ClientConfig(cadenceService))
}

func getCadenceClient() client.Client {
	clientOptions := &client.Options{}
	return client.NewClient(
		buildWorkflowServiceClient(), domain, clientOptions,
	)
}

// StartCadenceWorker comment
func StartCadenceWorker() {
	var err error

	serviceClient := buildWorkflowServiceClient()
	logger := buildLogger()
	var days int32
	days = 7
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	err = serviceClient.RegisterDomain(ctx, &shared.RegisterDomainRequest{
		Name:                                   &domain,
		WorkflowExecutionRetentionPeriodInDays: &days,
	})
	if _, ok := err.(*shared.DomainAlreadyExistsError); ok {
		logger.Info("Cadence domain already exists")
	} else if err != nil {
		logger.Info("failed to register Cadence domain")
	} else {
		logger.Info("Created Cadence domain")
	}

	workerOptions := worker.Options{
		Logger:       buildLogger(),
		MetricsScope: tally.NewTestScope(taskListName, map[string]string{}),
	}

	worker := worker.New(
		serviceClient,
		domain,
		taskListName,
		workerOptions)
	err = worker.Start()
	select {}
	if err != nil {
		panic("Failed to start worker")
	}
}

// StartWorkflow comment
func StartWorkflow() {
	ctx := context.Background()
	cadenceClient := getCadenceClient()

	workflowOptions := client.StartWorkflowOptions{
		TaskList:                     taskListName,
		ExecutionStartToCloseTimeout: time.Minute * 10,
		//DecisionTaskStartToCloseTimeout: time.Minute,
	}

	run, err := cadenceClient.ExecuteWorkflow(ctx, workflowOptions, workflow.DemoWorkFlow, "start")
	fmt.Println(run)
	if err != nil {
		fmt.Println(err)
		panic("Failed to start wrokflow")
	}
	log.Printf("workflow=%q, run=%q", run.GetID(), run.GetRunID())
	log.Println("done")
}
