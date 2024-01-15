/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ref "k8s.io/client-go/tools/reference"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	radiov1beta1 "github.com/frelon/k8s-radio/api/v1beta1"
)

const (
	RtlSdrResourceName = "frelon.se/rtl-sdr"
	RtlSdrDefaultImage = "rtl-sdr:dev"
)

// RtlSdrReceiverReconciler reconciles a RtlSdrReceiver object
type RtlSdrReceiverReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	Image string
}

// +kubebuilder:rbac:groups=radio.frelon.se,resources=rtlsdrreceivers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=radio.frelon.se,resources=rtlsdrreceivers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=radio.frelon.se,resources=rtlsdrreceivers/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete
func (r *RtlSdrReceiverReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithValues("name", req.NamespacedName.String())
	logger.Info("Reconciling RtlSdrReceiver")

	receiver := &radiov1beta1.RtlSdrReceiver{}
	if err := r.Get(ctx, req.NamespacedName, receiver); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("Object was not found, not an error")
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, fmt.Errorf("failed to get seedimage object: %w", err)
	}

	pod := &corev1.Pod{}
	if err := r.Get(ctx, req.NamespacedName, pod); err != nil {
		if !apierrors.IsNotFound(err) {
			logger.Error(err, "Error getting pod (Not not found), returning.")
			return reconcile.Result{}, err
		}

		receiver.Status.State = radiov1beta1.StateWaiting

		logger.Info("Pod not found, creating it...")

		err = r.createPod(ctx, receiver)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else {
		// Already running, update state based on pod Phase
		switch pod.Status.Phase {
		case corev1.PodRunning:
			receiver.Status.State = radiov1beta1.StateRunning
		case corev1.PodFailed:
			receiver.Status.State = radiov1beta1.StateFailed
		}
	}

	podRef, err := ref.GetReference(r.Scheme, pod)
	if err != nil {
		return reconcile.Result{}, err
	}

	receiver.Status.Pod = podRef

	logger.Info("Updating status")
	if err := r.Status().Update(ctx, receiver); err != nil {
		logger.Error(err, "Error updating RtlSdrReceiver status")
		return ctrl.Result{}, err
	}

	logger.Info("Reconcile successful.")
	return reconcile.Result{}, nil
}

func (r *RtlSdrReceiverReconciler) createPod(ctx context.Context, receiver *radiov1beta1.RtlSdrReceiver) error {
	pod := &corev1.Pod{}

	args := []string{"-a", "0.0.0.0"}
	if receiver.Spec.Frequency != nil {
		args = append(args, "-f", receiver.Spec.Frequency.String())
	}

	listenPort := 1234
	ports := []corev1.ContainerPort{}
	if receiver.Spec.ContainerPort != nil {
		ports = append(ports, *receiver.Spec.ContainerPort)
		listenPort = int(receiver.Spec.ContainerPort.ContainerPort)
	}

	args = append(args, "-p", strconv.Itoa(listenPort))

	pod.Name = receiver.Name
	pod.Namespace = receiver.Namespace
	pod.Spec = corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:    "receiver",
				Image:   r.Image,
				Command: []string{"/bin/rtl_tcp"},
				Args:    args,
				Ports:   ports,
				Resources: corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						RtlSdrResourceName: *resource.NewQuantity(1, resource.DecimalSI),
					},
				},
			},
		},
	}

	if err := controllerutil.SetControllerReference(receiver, pod, r.Scheme); err != nil {
		meta.SetStatusCondition(&receiver.Status.Conditions, metav1.Condition{
			Type:    radiov1beta1.ReadyCondition,
			Status:  metav1.ConditionFalse,
			Reason:  radiov1beta1.PodFailedReason,
			Message: err.Error(),
		})
		return err
	}

	return r.Create(ctx, pod)
}

var (
	jobOwnerKey = ".metadata.controller"
	apiGVStr    = radiov1beta1.GroupVersion.String()
)

// SetupWithManager sets up the controller with the Manager.
func (r *RtlSdrReceiverReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &corev1.Pod{}, jobOwnerKey, func(rawObj client.Object) []string {
		// grab the pod object, extract the owner...
		pod := rawObj.(*corev1.Pod)
		owner := metav1.GetControllerOf(pod)
		if owner == nil {
			return nil
		}

		// ...make sure it's a RtlSdrReceiver...
		if owner.APIVersion != apiGVStr || owner.Kind != "RtlSdrReceiver" {
			return nil
		}

		// ...and if so, return it
		return []string{owner.Name}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&radiov1beta1.RtlSdrReceiver{}).
		Owns(&corev1.Pod{}).
		Complete(r)
}
