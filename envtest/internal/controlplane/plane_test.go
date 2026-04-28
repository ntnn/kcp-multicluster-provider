/*
Copyright 2021 The Kubernetes Authors.

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

package controlplane_test

import (
	"context"

	kauthn "k8s.io/api/authorization/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	. "github.com/kcp-dev/multicluster-provider/envtest/internal/controlplane"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Control Plane", func() {
	It("should start and stop successfully", func() {
		plane := &Kcp{}
		Expect(plane.Start()).To(Succeed())
		Expect(plane.Stop()).To(Succeed())
	})

	// sharded-test-server uses hardcoded ports, so restart on the same
	// machine is unreliable without waiting for TIME_WAIT to expire.
	PIt("should be able to restart", func() {
		plane := &Kcp{}
		Expect(plane.Start()).To(Succeed())
		Expect(plane.Stop()).To(Succeed())
		Expect(plane.Start()).To(Succeed())
		Expect(plane.Stop()).To(Succeed())
	})

	Context("after having started", func() {
		var plane *Kcp
		BeforeEach(func() {
			plane = &Kcp{}
			Expect(plane.Start()).To(Succeed())
		})
		AfterEach(func() {
			Expect(plane.Stop()).To(Succeed())
		})

		It("should provision a working admin user", func() {
			user, err := plane.AddUser(User{Name: "admin", Groups: []string{"system:kcp:admin"}}, nil)
			Expect(err).NotTo(HaveOccurred())

			cfg := user.Config()
			cfg.Host += "/clusters/root"
			cl, err := client.New(cfg, client.Options{})
			Expect(err).NotTo(HaveOccurred(), "should be able to create a client")

			sar := &kauthn.SelfSubjectAccessReview{
				Spec: kauthn.SelfSubjectAccessReviewSpec{
					ResourceAttributes: &kauthn.ResourceAttributes{
						Verb:     "*",
						Group:    "*",
						Version:  "*",
						Resource: "*",
					},
				},
			}
			Expect(cl.Create(context.Background(), sar)).To(Succeed(), "should be able to make a Self-SAR")
			Expect(sar.Status.Allowed).To(BeTrue(), "admin user should be able to do everything")
		})
	})
})

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))
})
