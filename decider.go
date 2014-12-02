// decider
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/sclasen/swf4go"
)

var client *swf.Client

var poller *swf.DecisionTaskPoller

var sm *swf.PollerShutdownManager

func init() {
	client = swf.NewClient(swf.MustGetenv("AWS_ACCESS_KEY_ID"), swf.MustGetenv("AWS_SECRET_ACCESS_KEY"), swf.APNorthEast1)
}

// onTask
func onTask(resp *swf.PollForDecisionTaskResponse) {
	fmt.Println("Decider: Got a new decision task!")
	if resp.TaskToken == "" {
		fmt.Println("Decider: Decider is empty.")
		return
	}

	cReq := nextRequest(resp)

	if err := client.RespondDecisionTaskCompleted(cReq); err != nil {
		log.Println(err)
	}

	fmt.Println("Decider: The decision task end!")
}

// nextRequest decides next decision task by obtained task from task list in SWF
func nextRequest(resp *swf.PollForDecisionTaskResponse) swf.RespondDecisionTaskCompletedRequest {

	comp, total := 0, 0
	for _, he := range resp.Events {
		fmt.Printf("Decider: EventType - %v, EventId - %v\n", he.EventType, he.EventID)
		if he.EventType == swf.EventTypeActivityTaskCompleted {
			comp++
		}
		if isActivityType(he.EventType) {
			total++
		}
	}
	fmt.Printf("... completedCount = %v\n", comp)

	decisions := []swf.Decision{}
	if total == 0 { // beggining of the workflow
		decisions = scheduleActivity("HelloTask3", decisions)
		decisions = scheduleActivity("WorldTask3", decisions)
	} else if comp == 2 {
		// complete workflow
		decisions = append(decisions, swf.Decision{
			DecisionType: swf.DecisionTypeCompleteWorkflowExecution,
			CompleteWorkflowExecutionDecisionAttributes: &swf.CompleteWorkflowExecutionDecisionAttributes{
				Result: `{ "Result": "WF Complete!" }`,
			},
		})

		fmt.Println("Decider: WORKFLOW COMPLETE!!!!!!!!!!!!!!!!!!")
	}

	return swf.RespondDecisionTaskCompletedRequest{
		Decisions: decisions,
		TaskToken: resp.TaskToken,
	}
}

// isActivityType checks that the specified event is a type of activity
func isActivityType(e string) bool {
	return e == swf.EventTypeActivityTaskCancelRequested ||
		e == swf.EventTypeActivityTaskCanceled ||
		e == swf.EventTypeActivityTaskCompleted ||
		e == swf.EventTypeActivityTaskFailed ||
		e == swf.EventTypeActivityTaskFailed ||
		e == swf.EventTypeActivityTaskScheduled ||
		e == swf.EventTypeActivityTaskStarted ||
		e == swf.EventTypeActivityTaskTimedOut
}

// scheduleActivity schedules a activity to task list in SWF
func scheduleActivity(n string, d []swf.Decision) []swf.Decision {
	decision := swf.Decision{
		DecisionType: swf.DecisionTypeScheduleActivityTask,
		ScheduleActivityTaskDecisionAttributes: &swf.ScheduleActivityTaskDecisionAttributes{
			ActivityType: swf.ActivityType{
				Name:    n,
				Version: "1.0",
			},
			ActivityID: n + time.Now().String(),
		},
	}
	fmt.Printf("Decider: ActivityID = %v\n", decision.ScheduleActivityTaskDecisionAttributes.ActivityID)
	return append(d, decision)
}

func main() {
	poller = swf.NewDecisionTaskPoller(client, "ota-test", "", "HWTaskList")
	sm = swf.RegisterPollerShutdownManager()

	// main loop
	fmt.Println("Polling for decision task ...")
	poller.PollUntilShutdownBy(sm, "decisionTaskPoller", onTask)
}
