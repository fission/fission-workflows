// package workqueue is a amended copy of k8s' workqueue implementation
//
// Changes made:
// - workqueue.go 		- Added MaxSize field to default workqueue.
// - workqueue.go 		- Added bool return value whether value was added.
// - workqueue.go 		- Added Identifier interface to allow items in workqueue to deviate from the associated ID.
// - workqueue.go 		- Added Replace field to allow subsequent Adds of the same ID to simply replace the value.
// - delaying_queue.go 	- added non-blocking TryAddAfter.
// - all 				- Replaced t and set types with interface{} and map[interface{}]interface{}
//
// upstream source: https://github.com/kubernetes/client-go/tree/master/util/workqueue
package workqueue
