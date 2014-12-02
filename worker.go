// worker
package main

import (
	"fmt"
	"log"

	"github.com/sclasen/swf4go"
)

var client *swf.Client

func init() {
	client = swf.NewClient(swf.MustGetenv("AWS_ACCESS_KEY_ID"), swf.MustGetenv("AWS_SECRET_ACCESS_KEY"), swf.APNorthEast1)
}

func onTask(tln string) func(resp *swf.PollForActivityTaskResponse) {
	return func(resp *swf.PollForActivityTaskResponse) {
		if resp.ActivityID == "" {
			fmt.Println("")
		} else {
			result := doTask(resp.ActivityType)
			fmt.Printf("Worker%v: Activity task result - %v\n", tln, result)
			req := swf.RespondActivityTaskCompletedRequest{
				Result:    `{ "activityResult": ` + result + ` }`,
				TaskToken: resp.TaskToken,
			}
			if err := client.RespondActivityTaskCompleted(req); err != nil {
				log.Println(err)
				return
			}
			fmt.Printf("Worker%v: Activity task completed. ActivityID - %v\n", tln, resp.ActivityID)
		}
	}
}

func doTask(at swf.ActivityType) string {
	switch at.Name {
	case "HelloTask3":
		return "HELLO!!!"
	case "WorldTask3":
		return "WORLD!!!"
	default:
		return "MISSED!!!"
	}
}

func main() {

	sm := swf.RegisterPollerShutdownManager()

	//  main loop for "Hello" task list
	go func(tln string) {
		poller := swf.NewActivityTaskPoller(client, "ota-test", "", tln)
		fmt.Printf("Worker%v: Polling for activity task ...\n", tln)
		poller.PollUntilShutdownBy(sm, "helloActivityTaskPoller", onTask(tln))
	}("Hello")

	//  main loop for "World" task list
	go func(tln string) {
		poller := swf.NewActivityTaskPoller(client, "ota-test", "", tln)
		fmt.Printf("Worker%v: Polling for activity task ...\n", tln)
		poller.PollUntilShutdownBy(sm, "worldActivityTaskPoller", onTask(tln))
	}("World")

	select {}
}
