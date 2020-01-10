package controller
//
//import (
//	"context"
//	"fmt"
//	"github.com/pkg/errors"
//	"github.com/solo-io/autopilot/pkg/ezkube"
//	aphandler "github.com/solo-io/autopilot/pkg/handler"
//	"github.com/solo-io/autopilot/pkg/request"
//	"github.com/solo-io/autopilot/pkg/workqueue"
//	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
//	"k8s.io/apimachinery/pkg/runtime"
//	"sigs.k8s.io/controller-runtime/pkg/controller"
//	"sigs.k8s.io/controller-runtime/pkg/handler"
//	"sigs.k8s.io/controller-runtime/pkg/log"
//	"sigs.k8s.io/controller-runtime/pkg/manager"
//	"sigs.k8s.io/controller-runtime/pkg/predicate"
//	"sigs.k8s.io/controller-runtime/pkg/reconcile"
//	"sigs.k8s.io/controller-runtime/pkg/source"
//)
//
//// starts watches on a resource
//type Watcher interface {
//	Watch(res metav1.Object, h handler.EventHandler) (context.CancelFunc, error)
//}
//
//type watcher struct {
//	c controller.Controller
//}
//
//func (w *watcher) Watch(res metav1.Object, h handler.EventHandler) (context.CancelFunc, error) {
//	w.c.Watch()
//}