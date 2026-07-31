/*
Copyright 2026 amghazanfari.

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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	axonhubv1alpha1 "github.com/amghazanfari/axon-operator/api/v1alpha1"
)

var _ = Describe("AxonHub Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default",
		}
		axonhub := &axonhubv1alpha1.AxonHub{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind AxonHub")
			err := k8sClient.Get(ctx, typeNamespacedName, axonhub)
			if err != nil && errors.IsNotFound(err) {
				resource := &axonhubv1alpha1.AxonHub{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: axonhubv1alpha1.AxonHubSpec{
						Image:    "looplj/axonhub:latest",
						Replicas: ptr(int32(1)),
						Port:     8090,
						Postgres: axonhubv1alpha1.PostgresConfig{
							Enabled: ptr(true),
							Embedded: &axonhubv1alpha1.EmbeddedPostgres{
								Image:    "postgres:16",
								Database: "axonhub",
								User:     "axonhub",
								Storage:  "1Gi",
							},
						},
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			resource := &axonhubv1alpha1.AxonHub{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance AxonHub")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})
		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &AxonHubReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

func ptr[T any](v T) *T {
	return &v
}
