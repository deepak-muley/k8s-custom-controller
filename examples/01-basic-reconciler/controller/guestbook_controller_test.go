/*
Copyright 2025.

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
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	examplesv1alpha1 "github.com/deepak-muley/k8s-custom-controller/examples/01-basic-reconciler/api/v1alpha1"
)

var _ = Describe("Guestbook Controller", func() {
	Context("When reconciling a Guestbook resource", func() {
		const (
			timeout  = time.Second * 10
			interval = time.Millisecond * 250
		)

		ctx := context.Background()

		It("Should create and update Guestbook status", func() {
			By("Creating a new Guestbook")
			guestbook := &examplesv1alpha1.Guestbook{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-guestbook",
					Namespace: "default",
				},
				Spec: examplesv1alpha1.GuestbookSpec{
					Message:  "Hello World",
					Replicas: 3,
				},
			}
			Expect(k8sClient.Create(ctx, guestbook)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: guestbook.Name, Namespace: guestbook.Namespace}
			createdGuestbook := &examplesv1alpha1.Guestbook{}

			By("Checking if the Guestbook was created")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, lookupKey, createdGuestbook)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			By("Checking if the status is updated")
			Eventually(func() string {
				err := k8sClient.Get(ctx, lookupKey, createdGuestbook)
				if err != nil {
					return ""
				}
				return createdGuestbook.Status.State
			}, timeout, interval).Should(Equal("Active"))

			By("Verifying the status fields")
			Expect(createdGuestbook.Status.ObservedGeneration).Should(Equal(createdGuestbook.Generation))
			Expect(createdGuestbook.Status.LastUpdated.IsZero()).Should(BeFalse())

			By("Cleaning up the Guestbook")
			Expect(k8sClient.Delete(ctx, guestbook)).Should(Succeed())
		})

		It("Should update status when spec changes", func() {
			By("Creating a Guestbook")
			guestbook := &examplesv1alpha1.Guestbook{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-guestbook-update",
					Namespace: "default",
				},
				Spec: examplesv1alpha1.GuestbookSpec{
					Message:  "Initial Message",
					Replicas: 1,
				},
			}
			Expect(k8sClient.Create(ctx, guestbook)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: guestbook.Name, Namespace: guestbook.Namespace}
			createdGuestbook := &examplesv1alpha1.Guestbook{}

			By("Waiting for initial status update")
			Eventually(func() string {
				err := k8sClient.Get(ctx, lookupKey, createdGuestbook)
				if err != nil {
					return ""
				}
				return createdGuestbook.Status.State
			}, timeout, interval).Should(Equal("Active"))

			initialGeneration := createdGuestbook.Status.ObservedGeneration
			initialUpdatedTime := createdGuestbook.Status.LastUpdated

			By("Updating the Guestbook spec")
			createdGuestbook.Spec.Message = "Updated Message"
			createdGuestbook.Spec.Replicas = 5
			Expect(k8sClient.Update(ctx, createdGuestbook)).Should(Succeed())

			By("Verifying status is updated after spec change")
			Eventually(func() int64 {
				err := k8sClient.Get(ctx, lookupKey, createdGuestbook)
				if err != nil {
					return 0
				}
				return createdGuestbook.Status.ObservedGeneration
			}, timeout, interval).Should(BeNumerically(">", initialGeneration))

			By("Verifying LastUpdated timestamp changed")
			Expect(createdGuestbook.Status.LastUpdated.After(initialUpdatedTime.Time)).Should(BeTrue())

			By("Cleaning up")
			Expect(k8sClient.Delete(ctx, guestbook)).Should(Succeed())
		})
	})
})
