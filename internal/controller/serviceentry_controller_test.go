package controller

import (
	"context"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/schmiddim/blackbox-operator/pkg/config"
	"istio.io/api/networking/v1alpha3"
	istioNetworking "istio.io/client-go/pkg/apis/networking/v1alpha3"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("ServiceEntry Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default", // TODO(user):Modify as needed
		}
		serviceentry := &istioNetworking.ServiceEntry{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name:      resourceName,
				Namespace: "default",
			},
			Spec: v1alpha3.ServiceEntry{
				Hosts: []string{
					"mein-service.example.com", // Hostname des Services
				},
				Ports: []*v1alpha3.ServicePort{
					{
						Number:   8080,
						Protocol: "http",
						Name:     "http",
					},
				},
			},
		}

		BeforeEach(func() {
			By("creating the custom resource for the Kind ServiceEntry")
			err := k8sClient.Get(ctx, typeNamespacedName, serviceentry)
			if err != nil && errors.IsNotFound(err) {
				//resource := &istioNetworking.ServiceEntry{
				//	ObjectMeta: metav1.ObjectMeta{
				//		Name:      resourceName,
				//		Namespace: "default",
				//	},
				//	// TODO(user): Specify other spec details if needed.
				//}
				Expect(k8sClient.Create(ctx, serviceentry)).To(Succeed())
			}
		})

		AfterEach(func() {
			// TODO(user): Cleanup logic after each test, like removing the resource instance.
			//resource := &v1alpha3.ServiceEntry{}
			//err := k8sClient.Get(ctx, typeNamespacedName, resource)
			//Expect(err).NotTo(HaveOccurred())

			//By("Cleanup the specific resource instance ServiceEntry")
			//Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})
		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &ServiceEntryReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
				Config: &config.Config{
					LogLevel:                    "info",
					DefaultModule:               "http_2xx",
					ServiceMonitorNamingPattern: "sm-%s",
					Interval:                    "10s",
					ScrapeTimeout:               "10s",
					LabelSelector:               metav1.LabelSelector{},
					ExcludeSelector:             metav1.LabelSelector{},
					ProtocolModuleMappings:      nil,
				},
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			resource := &monitoringv1.ServiceMonitor{}
			//resource := &v1alpha3.ServiceEntry{}

			typeNamespacedName := types.NamespacedName{
				Name:      "sm-" + resourceName,
				Namespace: "default",
			}
			err = k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())
			// TODO(user): Add more specific assertions depending on your controller's reconciliation logic.
			// Example: If you expect a certain status condition after reconciliation, verify it here.
		})
	})
})
