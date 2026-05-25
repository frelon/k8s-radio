package controller

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	radiov1beta1 "github.com/frelon/k8s-radio/api/v1beta1"
)

// RtlSdrReceiverReconciler reconciles a RtlSdrReceiver object
type RtlSdrReceiverReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=radio.frelon.se,resources=rtlsdrreceivers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=radio.frelon.se,resources=rtlsdrreceivers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=radio.frelon.se,resources=rtlsdrreceivers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the RtlSdrReceiver object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.23.3/pkg/reconcile
func (r *RtlSdrReceiverReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = logf.FromContext(ctx)

	// TODO(user): your logic here

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RtlSdrReceiverReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&radiov1beta1.RtlSdrReceiver{}).
		Named("rtlsdrreceiver").
		Complete(r)
}
