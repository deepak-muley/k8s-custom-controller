package config

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Config Suite")
}

var _ = Describe("Config", func() {
	Describe("NewConfig", func() {
		It("should create a new Config with default values", func() {
			cfg := NewConfig()

			Expect(cfg).NotTo(BeNil())
			Expect(cfg.Namespace).To(Equal("default"))
			Expect(cfg.GVKs).To(BeEmpty())
		})

		It("should create an empty GVKs slice", func() {
			cfg := NewConfig()

			Expect(cfg.GVKs).To(HaveLen(0))
			Expect(cfg.GVKs).NotTo(BeNil())
		})
	})

	Describe("Validate", func() {
		Context("with valid configuration", func() {
			It("should validate config with at least one GVK", func() {
				cfg := &Config{
					Namespace: "default",
					GVKs: []schema.GroupVersionKind{
						{
							Group:   "apps",
							Version: "v1",
							Kind:    "Deployment",
						},
					},
				}

				err := cfg.Validate()
				Expect(err).NotTo(HaveOccurred())
			})

			It("should validate config with multiple GVKs", func() {
				cfg := &Config{
					Namespace: "production",
					GVKs: []schema.GroupVersionKind{
						{
							Group:   "apps",
							Version: "v1",
							Kind:    "Deployment",
						},
						{
							Group:   "",
							Version: "v1",
							Kind:    "Service",
						},
						{
							Group:   "apps",
							Version: "v1",
							Kind:    "ReplicaSet",
						},
					},
				}

				err := cfg.Validate()
				Expect(err).NotTo(HaveOccurred())
			})

			It("should validate config with empty namespace (all namespaces)", func() {
				cfg := &Config{
					Namespace: "",
					GVKs: []schema.GroupVersionKind{
						{
							Group:   "apps",
							Version: "v1",
							Kind:    "Deployment",
						},
					},
				}

				err := cfg.Validate()
				Expect(err).NotTo(HaveOccurred())
			})

			It("should validate config with custom namespace", func() {
				cfg := &Config{
					Namespace: "my-namespace",
					GVKs: []schema.GroupVersionKind{
						{
							Group:   "apps",
							Version: "v1",
							Kind:    "Deployment",
						},
					},
				}

				err := cfg.Validate()
				Expect(err).NotTo(HaveOccurred())
			})

			It("should validate config with core API group GVK", func() {
				cfg := &Config{
					Namespace: "default",
					GVKs: []schema.GroupVersionKind{
						{
							Group:   "",
							Version: "v1",
							Kind:    "Pod",
						},
					},
				}

				err := cfg.Validate()
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("with invalid configuration", func() {
			It("should fail validation with empty GVKs slice", func() {
				cfg := &Config{
					Namespace: "default",
					GVKs:      []schema.GroupVersionKind{},
				}

				err := cfg.Validate()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("at least one GVK must be specified"))
			})

			It("should fail validation with nil GVKs slice", func() {
				cfg := &Config{
					Namespace: "default",
					GVKs:      nil,
				}

				err := cfg.Validate()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("at least one GVK must be specified"))
			})
		})

		Context("with edge cases", func() {
			It("should handle config created by NewConfig without modification", func() {
				cfg := NewConfig()

				err := cfg.Validate()
				Expect(err).To(HaveOccurred()) // Should fail because GVKs is empty
				Expect(err.Error()).To(ContainSubstring("at least one GVK must be specified"))
			})

			It("should validate after adding GVKs to NewConfig result", func() {
				cfg := NewConfig()
				cfg.GVKs = []schema.GroupVersionKind{
					{
						Group:   "apps",
						Version: "v1",
						Kind:    "Deployment",
					},
				}

				err := cfg.Validate()
				Expect(err).NotTo(HaveOccurred())
			})

			It("should validate with very long namespace name", func() {
				longNamespace := "a" + string(make([]byte, 200)) // 200 character namespace
				cfg := &Config{
					Namespace: longNamespace,
					GVKs: []schema.GroupVersionKind{
						{
							Group:   "apps",
							Version: "v1",
							Kind:    "Deployment",
						},
					},
				}

				err := cfg.Validate()
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("Config fields", func() {
		It("should preserve Namespace field", func() {
			cfg := &Config{
				Namespace: "test-namespace",
				GVKs: []schema.GroupVersionKind{
					{
						Group:   "apps",
						Version: "v1",
						Kind:    "Deployment",
					},
				},
			}

			Expect(cfg.Namespace).To(Equal("test-namespace"))
		})

		It("should preserve GVKs field", func() {
			gvks := []schema.GroupVersionKind{
				{
					Group:   "apps",
					Version: "v1",
					Kind:    "Deployment",
				},
				{
					Group:   "apps",
					Version: "v1",
					Kind:    "ReplicaSet",
				},
			}

			cfg := &Config{
				Namespace: "default",
				GVKs:      gvks,
			}

			Expect(cfg.GVKs).To(Equal(gvks))
			Expect(cfg.GVKs).To(HaveLen(2))
		})

		It("should allow modification of Namespace after creation", func() {
			cfg := NewConfig()
			cfg.Namespace = "modified-namespace"

			Expect(cfg.Namespace).To(Equal("modified-namespace"))
		})

		It("should allow modification of GVKs after creation", func() {
			cfg := NewConfig()
			cfg.GVKs = []schema.GroupVersionKind{
				{
					Group:   "apps",
					Version: "v1",
					Kind:    "Deployment",
				},
			}

			Expect(cfg.GVKs).To(HaveLen(1))
			Expect(cfg.GVKs[0].Kind).To(Equal("Deployment"))
		})
	})
})
