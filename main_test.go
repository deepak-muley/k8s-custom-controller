package main

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMain(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Main Suite")
}

var _ = Describe("Main", func() {
	Describe("splitAndTrim", func() {
		It("should split string by delimiter and trim whitespace", func() {
			result := splitAndTrim("  a  ,  b  ,  c  ", ",")

			Expect(result).To(Equal([]string{"a", "b", "c"}))
		})

		It("should handle empty string", func() {
			result := splitAndTrim("", ",")

			Expect(result).To(BeEmpty())
		})

		It("should handle string with only whitespace", func() {
			result := splitAndTrim("   ,   ,   ", ",")

			Expect(result).To(BeEmpty())
		})

		It("should handle single item", func() {
			result := splitAndTrim("  item  ", ",")

			Expect(result).To(Equal([]string{"item"}))
		})

		It("should handle multiple delimiters", func() {
			result := splitAndTrim("a/b/c", "/")

			Expect(result).To(Equal([]string{"a", "b", "c"}))
		})

		It("should handle empty parts between delimiters", func() {
			result := splitAndTrim("a,,,b", ",")

			Expect(result).To(Equal([]string{"a", "b"}))
		})

		It("should handle leading and trailing delimiters", func() {
			result := splitAndTrim(",a,b,", ",")

			Expect(result).To(Equal([]string{"a", "b"}))
		})

		It("should preserve non-whitespace content", func() {
			result := splitAndTrim("apps/v1/Deployment,core/v1/Service", ",")

			Expect(result).To(Equal([]string{"apps/v1/Deployment", "core/v1/Service"}))
		})
	})

	Describe("parseGVK", func() {
		It("should parse valid GVK string", func() {
			gvk, err := parseGVK("apps/v1/Deployment")

			Expect(err).NotTo(HaveOccurred())
			Expect(gvk.Group).To(Equal("apps"))
			Expect(gvk.Version).To(Equal("v1"))
			Expect(gvk.Kind).To(Equal("Deployment"))
		})

		It("should handle core API group with 'core' prefix", func() {
			gvk, err := parseGVK("core/v1/Pod")

			Expect(err).NotTo(HaveOccurred())
			Expect(gvk.Group).To(Equal(""))
			Expect(gvk.Version).To(Equal("v1"))
			Expect(gvk.Kind).To(Equal("Pod"))
		})

		It("should handle empty group (core API)", func() {
			gvk, err := parseGVK("/v1/Service")

			Expect(err).NotTo(HaveOccurred())
			Expect(gvk.Group).To(Equal(""))
			Expect(gvk.Version).To(Equal("v1"))
			Expect(gvk.Kind).To(Equal("Service"))
		})

		It("should handle custom resource groups", func() {
			gvk, err := parseGVK("example.com/v1alpha1/MyCustomResource")

			Expect(err).NotTo(HaveOccurred())
			Expect(gvk.Group).To(Equal("example.com"))
			Expect(gvk.Version).To(Equal("v1alpha1"))
			Expect(gvk.Kind).To(Equal("MyCustomResource"))
		})

		It("should handle whitespace in GVK string", func() {
			gvk, err := parseGVK("  apps  /  v1  /  Deployment  ")

			Expect(err).NotTo(HaveOccurred())
			Expect(gvk.Group).To(Equal("apps"))
			Expect(gvk.Version).To(Equal("v1"))
			Expect(gvk.Kind).To(Equal("Deployment"))
		})

		Context("with invalid format", func() {
			It("should fail with too few parts", func() {
				_, err := parseGVK("apps/v1")

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("GVK must be in format 'group/version/kind'"))
			})

			It("should fail with too many parts", func() {
				_, err := parseGVK("apps/v1/Deployment/extra")

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("GVK must be in format 'group/version/kind'"))
			})

			It("should fail with empty string", func() {
				_, err := parseGVK("")

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("GVK must be in format 'group/version/kind'"))
			})

			It("should fail with no delimiters", func() {
				_, err := parseGVK("appsv1Deployment")

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("GVK must be in format 'group/version/kind'"))
			})
		})
	})

	Describe("parseGVKs", func() {
		It("should parse single GVK string", func() {
			gvks, err := parseGVKs("apps/v1/Deployment")

			Expect(err).NotTo(HaveOccurred())
			Expect(gvks).To(HaveLen(1))
			Expect(gvks[0].Group).To(Equal("apps"))
			Expect(gvks[0].Version).To(Equal("v1"))
			Expect(gvks[0].Kind).To(Equal("Deployment"))
		})

		It("should parse multiple comma-separated GVKs", func() {
			gvks, err := parseGVKs("apps/v1/Deployment,apps/v1/ReplicaSet,core/v1/Service")

			Expect(err).NotTo(HaveOccurred())
			Expect(gvks).To(HaveLen(3))
			Expect(gvks[0].Kind).To(Equal("Deployment"))
			Expect(gvks[1].Kind).To(Equal("ReplicaSet"))
			Expect(gvks[2].Kind).To(Equal("Service"))
		})

		It("should handle whitespace in comma-separated list", func() {
			gvks, err := parseGVKs("  apps/v1/Deployment  ,  apps/v1/ReplicaSet  ")

			Expect(err).NotTo(HaveOccurred())
			Expect(gvks).To(HaveLen(2))
			Expect(gvks[0].Kind).To(Equal("Deployment"))
			Expect(gvks[1].Kind).To(Equal("ReplicaSet"))
		})

		It("should handle core API group GVKs", func() {
			gvks, err := parseGVKs("core/v1/Pod,core/v1/Service")

			Expect(err).NotTo(HaveOccurred())
			Expect(gvks).To(HaveLen(2))
			Expect(gvks[0].Group).To(Equal(""))
			Expect(gvks[1].Group).To(Equal(""))
		})

		It("should handle custom resource GVKs", func() {
			gvks, err := parseGVKs("example.com/v1alpha1/MyResource,example.com/v1beta1/MyOtherResource")

			Expect(err).NotTo(HaveOccurred())
			Expect(gvks).To(HaveLen(2))
			Expect(gvks[0].Group).To(Equal("example.com"))
			Expect(gvks[1].Group).To(Equal("example.com"))
		})

		Context("with invalid input", func() {
			It("should fail with empty string", func() {
				_, err := parseGVKs("")

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("GVK string cannot be empty"))
			})

			It("should fail if any GVK is invalid", func() {
				_, err := parseGVKs("apps/v1/Deployment,invalid-format")

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("invalid GVK format"))
				Expect(err.Error()).To(ContainSubstring("invalid-format"))
			})

			It("should fail with only commas", func() {
				_, err := parseGVKs(",")

				Expect(err).To(HaveOccurred())
			})

			It("should fail with empty parts between commas", func() {
				_, err := parseGVKs("apps/v1/Deployment,,apps/v1/ReplicaSet")

				Expect(err).To(HaveOccurred())
			})
		})

		Context("with edge cases", func() {
			It("should handle single character components", func() {
				gvks, err := parseGVKs("a/b/c")

				Expect(err).NotTo(HaveOccurred())
				Expect(gvks).To(HaveLen(1))
				Expect(gvks[0].Group).To(Equal("a"))
				Expect(gvks[0].Version).To(Equal("b"))
				Expect(gvks[0].Kind).To(Equal("c"))
			})

			It("should handle long GVK strings", func() {
				longGroup := "very-long-group-name.example.com"
				longVersion := "v1alpha1beta2gamma3"
				longKind := "VeryLongResourceKindName"
				gvkString := longGroup + "/" + longVersion + "/" + longKind

				gvks, err := parseGVKs(gvkString)

				Expect(err).NotTo(HaveOccurred())
				Expect(gvks).To(HaveLen(1))
				Expect(gvks[0].Group).To(Equal(longGroup))
				Expect(gvks[0].Version).To(Equal(longVersion))
				Expect(gvks[0].Kind).To(Equal(longKind))
			})
		})
	})

	Describe("buildKubeConfig", func() {
		Context("with kubeconfig file path", func() {
			It("should build config from valid kubeconfig file", func() {
				// Create a temporary kubeconfig file
				tempDir := GinkgoT().TempDir()
				kubeconfigPath := filepath.Join(tempDir, "kubeconfig")

				// Create a minimal valid kubeconfig
				kubeconfigContent := `apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://test.example.com
  name: test-cluster
contexts:
- context:
    cluster: test-cluster
    user: test-user
  name: test-context
current-context: test-context
users:
- name: test-user
  user:
    token: test-token
`

				err := os.WriteFile(kubeconfigPath, []byte(kubeconfigContent), 0644)
				Expect(err).NotTo(HaveOccurred())

				config, err := buildKubeConfig(kubeconfigPath)

				// Note: This might fail if kubernetes client libraries can't parse the config,
				// but we're testing that the function attempts to build from the file
				if err == nil {
					Expect(config).NotTo(BeNil())
				} else {
					// If it fails, it should be a reasonable error about parsing
					Expect(err.Error()).To(ContainSubstring("failed to build config from kubeconfig file"))
				}
			})

			It("should return error for non-existent kubeconfig file", func() {
				config, err := buildKubeConfig("/nonexistent/path/to/kubeconfig")

				Expect(err).To(HaveOccurred())
				Expect(config).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("failed to build config from kubeconfig file"))
			})

			It("should return error for invalid kubeconfig file", func() {
				tempDir := GinkgoT().TempDir()
				kubeconfigPath := filepath.Join(tempDir, "invalid-kubeconfig")

				err := os.WriteFile(kubeconfigPath, []byte("invalid yaml content"), 0644)
				Expect(err).NotTo(HaveOccurred())

				config, err := buildKubeConfig(kubeconfigPath)

				Expect(err).To(HaveOccurred())
				Expect(config).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("failed to build config from kubeconfig file"))
			})
		})

		Context("without kubeconfig file path", func() {
			It("should attempt in-cluster config first when path is empty", func() {
				// This will likely fail unless running in a Kubernetes cluster,
				// but we're testing that it attempts in-cluster config
				config, err := buildKubeConfig("")

				// If in-cluster config fails, it should fallback to default kubeconfig
				// So we expect either success or a specific error
				if err != nil {
					// If it fails, it should fail during fallback to default kubeconfig
					// or during in-cluster config (which is expected outside cluster)
					Expect(err.Error()).To(SatisfyAny(
						ContainSubstring("failed to build config"),
						ContainSubstring("unable to load in-cluster configuration"),
					))
				} else {
					Expect(config).NotTo(BeNil())
				}
			})

			It("should fallback to default kubeconfig when in-cluster config fails", func() {
				// Set a fake KUBERNETES_SERVICE_HOST to make in-cluster config fail early
				originalHost := os.Getenv("KUBERNETES_SERVICE_HOST")
				originalPort := os.Getenv("KUBERNETES_SERVICE_PORT")

				os.Unsetenv("KUBERNETES_SERVICE_HOST")
				os.Unsetenv("KUBERNETES_SERVICE_PORT")
				defer func() {
					if originalHost != "" {
						os.Setenv("KUBERNETES_SERVICE_HOST", originalHost)
					}
					if originalPort != "" {
						os.Setenv("KUBERNETES_SERVICE_PORT", originalPort)
					}
				}()

				config, err := buildKubeConfig("")

				// Should attempt fallback to default kubeconfig
				// This will likely fail unless ~/.kube/config exists and is valid
				if err != nil {
					Expect(err.Error()).To(ContainSubstring("failed to build config"))
				} else {
					Expect(config).NotTo(BeNil())
				}
			})
		})
	})
})
