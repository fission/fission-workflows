// package workqueue is a amended copy of k8s' workqueue implementation
//
// Changes made:
// - workqueue.go - Added MaxSize field to default workqueue.
// - workqueue.go - Added bool return value whether value was added.
// - delaying_queue.go - added non-blocking TryAddAfter.
//
// upstream source: https://github.com/kubernetes/client-go/tree/master/util/workqueue
package workqueue
