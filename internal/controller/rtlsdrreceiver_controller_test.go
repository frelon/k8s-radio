package controller

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	radiov1 "github.com/frelon/k8s-radio/api/v1beta1"
)

var _ = Describe("RtlSdrReceiver controller", func() {
	const (
		ReceiverName      = "test-receiver"
		ReceiverNamespace = "default"

		timeout  = time.Second * 10
		duration = time.Second * 10
		interval = time.Millisecond * 250
	)

	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(radiov1.AddToScheme(scheme))

	Context("When updating RtlSdrReceiver Status", func() {
		It("Should successfully create a new RtlSdrReceiver", func() {
			By("By creating a new RtlSdrReceiver")

			ctx := context.Background()
			freq, err := resource.ParseQuantity("101.9M")
			Expect(err).To(Succeed())

			port := int32(1212)

			recv := &radiov1.RtlSdrReceiver{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "radio.frelon.se/v1",
					Kind:       "RtlSdrReceiver",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      ReceiverName,
					Namespace: ReceiverNamespace,
				},
				Spec: radiov1.RtlSdrReceiverSpec{
					Version: radiov1.V3,
					ContainerPort: &corev1.ContainerPort{
						ContainerPort: port,
					},
					Frequency: &freq,
				},
			}

			Expect(k8sClient.Create(ctx, recv)).Should(Succeed())

			receiverLookupKey := types.NamespacedName{Name: ReceiverName, Namespace: ReceiverNamespace}
			createdReceiver := &radiov1.RtlSdrReceiver{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, receiverLookupKey, createdReceiver)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(createdReceiver.Spec.Version).Should(Equal(radiov1.V3))
			Expect(createdReceiver.Spec.Frequency.String()).Should(Equal("101900k"))
			Expect(createdReceiver.Spec.ContainerPort.ContainerPort).Should(Equal(port))
			Expect(createdReceiver.Status.State).Should(BeEmpty())

			By("By checking the RtlSdrReceiver has no active Pod")
			Consistently(func() (bool, error) {
				err := k8sClient.Get(ctx, receiverLookupKey, createdReceiver)
				if err != nil {
					return false, err
				}
				return createdReceiver.Status.Pod == nil, nil
			}, duration, interval).Should(Equal(true))

			By("By running reconciler")
			reconciler := RtlSdrReceiverReconciler{
				Client: k8sClient,
				Scheme: scheme,
				Image:  "test-image",
			}
			_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: receiverLookupKey})
			Expect(err).To(Succeed())

			// By("By checking that the RtlSdrReceiver has the correct pod reference")
			// Eventually(func() (string, error) {
			// 	err := k8sClient.Get(ctx, receiverLookupKey, createdReceiver)
			// 	if err != nil {
			// 		return "", err
			// 	}

			// 	if createdReceiver.Status.Pod == nil {
			// 		return "", nil
			// 	}

			// 	return createdReceiver.Status.Pod.Name, nil
			// }, timeout, interval).Should(Equal(ReceiverName))
		})
	})
})
