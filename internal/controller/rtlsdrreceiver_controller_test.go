package controller

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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

	Context("When updating RtlSdrReceiver Status", func() {
		It("Should successfully create a new RtlSdrReceiver", func() {
			By("By creating a new RtlSdrReceiver")

			ctx := context.Background()
			freq, err := resource.ParseQuantity("101.9M")
			Expect(err).To(Succeed())

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
						ContainerPort: 1212,
					},
					Frequency: &freq,
				},
			}

			Expect(k8sClient.Create(ctx, recv)).Should(Succeed())
		})
	})
})
