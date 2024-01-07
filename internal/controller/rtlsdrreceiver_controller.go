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

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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
}

// +kubebuilder:rbac:groups=radio.frelon.se,resources=rtlsdrreceivers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=radio.frelon.se,resources=rtlsdrreceivers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=radio.frelon.se,resources=rtlsdrreceivers/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete
func (r *RtlSdrReceiverReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	receiver := &radiov1beta1.RtlSdrReceiver{}
	if err := r.Get(ctx, req.NamespacedName, receiver); err != nil {
		if apierrors.IsNotFound(err) {
			logger.V(5).Info("Object was not found, not an error")
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

		logger.V(5).Info("Pod not found, creating it...")

		pod.Name = receiver.Name
		pod.Namespace = receiver.Namespace
		pod.Spec = corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "receiver",
					Image:   RtlSdrDefaultImage,
					Command: []string{"/bin/rtl_tcp"},
					Args:    []string{"-a", "0.0.0.0", "-f", receiver.Spec.Frequency.String()},
					Ports: []corev1.ContainerPort{
						{
							HostPort:      1234,
							ContainerPort: 1234,
							Protocol:      corev1.ProtocolTCP,
						},
					},
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
			return reconcile.Result{}, err
		}

		if err := r.Create(ctx, pod); err != nil {
			logger.Error(err, "Error getting pod (Not not found), returning.")
			return reconcile.Result{}, err
		}
	}

	logger.V(5).Info("Reconcile successful.")

	return reconcile.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RtlSdrReceiverReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&radiov1beta1.RtlSdrReceiver{}).
		Owns(&corev1.Pod{}).
		Complete(r)
}
