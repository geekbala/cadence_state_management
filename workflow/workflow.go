package workflow

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"time"

	"go.uber.org/cadence"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/workflow"
)

func init() {
	workflow.Register(DemoWorkFlow)
	activity.Register(getNameActivity)
	activity.Register(sayHello)
	activity.Register(bye)
}

func find(slice []int, val int) bool {
	log.Println("find method", val)
	for _, item := range slice {
		log.Println("item", item)
		if item == val {
			return true
		}
	}
	return false
}

func getNameActivity() (string, error) {
	fmt.Println("getnameactivity begins ")
	rand.Seed(time.Now().UnixNano())
	min := 1
	max := 10
	r := rand.Intn(max-min+1) + min
	log.Println(time.Now().UTC(), " name activity random value", r)
	ok := find([]int{1,2,3,4,5,6}, r)
	log.Println("ok result", ok)
	if ok {
		log.Println("throws an error from getname")
		return "", errors.New("keep-name-error")
	}
	log.Println("going to return cadence")
	return "cadence", nil
}

func sayHello(name string) (string, error) {
	fmt.Println("sayhello begins ")
	rand.Seed(time.Now().UnixNano())
	min := 1
	max := 10
	r := rand.Intn(max-min+1) + min
	log.Println(time.Now().UTC(), " hello activity random value", r)
	ok := find([]int{1, 2,3, 4,5,8, 9, 10}, r)
	if ok {
		log.Println("throws an error from sayhello")
		return "", errors.New("keep-hello-error")
	}
	return "Helloooooooooooooooooooooooooooo " + name + "!!!!!!!!!!!!!!!!!!!!!!!", nil
}

func bye(name string) (string, error) {
	fmt.Println("bye begins ")
	rand.Seed(time.Now().UnixNano())
	min := 1
	max := 10
	r := rand.Intn(max-min+1) + min
	log.Println(time.Now().UTC(), " bye activity random value", r)
	ok := find([]int{2, 3, 4}, r)
	if ok {
		log.Println("throws an error from bye")
		return "", errors.New("keep-bye-error")
	}
	return "bye bye" + name, nil
}

//DemoWorkFlow comment
func DemoWorkFlow(ctx workflow.Context, state string) error {
	var err error
	ao := workflow.ActivityOptions{
		TaskList:               "cadence-helloworld-task",
		StartToCloseTimeout:    1 * time.Minute,
		ScheduleToStartTimeout: 1 * time.Minute,
		RetryPolicy: &cadence.RetryPolicy{
			InitialInterval: time.Second,
			MaximumInterval:    10 * time.Second,
			ExpirationInterval: 1 * time.Minute,
			MaximumAttempts:    5,
		},
	}

	ctx = workflow.WithActivityOptions(ctx, ao)

	var name, result string
	switch state {
	case "start":
		state = "name"
		fallthrough
	case "name":
		err = workflow.ExecuteActivity(ctx, getNameActivity).Get(ctx, &name)
		log.Println("got the result as ", name)
		if err != nil {
			log.Println(" getNameActivity failed ", err)
			return workflow.NewContinueAsNewError(ctx, DemoWorkFlow, state)
		}
		log.Println("changing state to hello")
		state = "hello"
		fallthrough
	case "hello":
		err = workflow.ExecuteActivity(ctx, sayHello, name).Get(ctx, &result)
		if err != nil {
			log.Println(" sayHelloActivity failed ", err)
			return workflow.NewContinueAsNewError(ctx, DemoWorkFlow, state)
		}
		state = "bye"
		fallthrough
	case "bye":
		err = workflow.ExecuteActivity(ctx, bye, name).Get(ctx, &result)
		if err != nil {
			log.Println(" byeActivity failed ", err)
			return workflow.NewContinueAsNewError(ctx, DemoWorkFlow, state)
		}
	}
	workflow.GetLogger(ctx).Info("Result " + result)

	return nil
}
