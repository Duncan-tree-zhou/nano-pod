package webhookhandler

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	admissionv1 "k8s.io/api/admission/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/kubernetes/scheme"
	nanopodv1 "nano-pod-operator/api/v1"
	"nano-pod-operator/internal/patcherfactory"
	"nano-pod-operator/internal/patcherfactory/patcher"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"testing"
)

var (
	terminationGracePeriodSeconds = int64(10)
	nanoPod                       = &nanopodv1.NanoPod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default",
			Namespace: "nano-pod-test",
		},
		Spec: nanopodv1.NanoPodSpec{
			PatchStrategy: nanopodv1.OverWritePatch,
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"nanopod/test": "enabled-case",
					},
					Labels: map[string]string{
						"nanopod-test": "enabled-case",
					},
				},
				Spec: v1.PodSpec{
					TerminationGracePeriodSeconds: &terminationGracePeriodSeconds,
					Containers: []v1.Container{
						{
							Image: "treezh-docker.pkg.coding.net/demo03/public/nginx:1.21",
							Name:  "nginx02",
							Env: []v1.EnvVar{
								{
									Name:  "env0201",
									Value: "value0201",
								},
							},
							LivenessProbe: &v1.Probe{
								ProbeHandler: v1.ProbeHandler{
									TCPSocket: &v1.TCPSocketAction{
										Port: intstr.IntOrString{
											Type:   intstr.String,
											StrVal: "80",
										},
									},
								},
								InitialDelaySeconds: 5,
								PeriodSeconds:       10,
							},
							ReadinessProbe: &v1.Probe{
								ProbeHandler: v1.ProbeHandler{
									TCPSocket: &v1.TCPSocketAction{
										Port: intstr.IntOrString{
											Type:   intstr.String,
											StrVal: "80",
										},
									},
								},
								InitialDelaySeconds: 5,
								PeriodSeconds:       10,
							},
							Resources: v1.ResourceRequirements{
								Limits: v1.ResourceList{
									v1.ResourceCPU:    *resource.NewQuantity(500, resource.DecimalSI),
									v1.ResourceMemory: *resource.NewQuantity(128, resource.BinarySI),
								},
								Requests: v1.ResourceList{
									v1.ResourceCPU:    *resource.NewQuantity(500, resource.DecimalSI),
									v1.ResourceMemory: *resource.NewQuantity(128, resource.BinarySI),
								},
							},
						},
						{
							Name: "mysql01",
							Env: []v1.EnvVar{
								{
									Name:  "env0102",
									Value: "value0102",
								},
								{
									Name:  "env0101",
									Value: "value0103",
								},
							},
							LivenessProbe: &v1.Probe{
								ProbeHandler: v1.ProbeHandler{
									Exec: &v1.ExecAction{
										Command: []string{
											"mysqladmin", "ping",
										},
									},
								},
								InitialDelaySeconds: 30,
								PeriodSeconds:       10,
								TimeoutSeconds:      5,
							},
							ReadinessProbe: &v1.Probe{
								ProbeHandler: v1.ProbeHandler{
									Exec: &v1.ExecAction{
										Command: []string{
											"mysql", "-h", "127.0.0.1", "-e", "SELECT 1",
										},
									},
								},
								InitialDelaySeconds: 5,
								PeriodSeconds:       2,
								TimeoutSeconds:      1,
							},
						},
					},
				},
			},
		},
	}
	nginxPod = &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "nginx-nanopod-enabled-64f887f69f-",
			Namespace:    "nano-pod-test",
			Labels: map[string]string{
				"app":       "nginx",
				"nano-pods": "",
			},
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  "mysql01",
					Image: "mysql:8.0.31",
					Env: []v1.EnvVar{
						{
							Name:  "env0101",
							Value: "value0101",
						},
						{
							Name:  "MYSQL_ALLOW_EMPTY_PASSWORD",
							Value: "true",
						},
						{
							Name:  "MYSQL_DATABASE",
							Value: "mydb",
						},
					},
					LivenessProbe: &v1.Probe{
						ProbeHandler: v1.ProbeHandler{
							TCPSocket: &v1.TCPSocketAction{
								Port: intstr.IntOrString{
									Type:   intstr.String,
									StrVal: "3306",
								},
							},
						},
						InitialDelaySeconds: 5,
						PeriodSeconds:       10,
					},
					ReadinessProbe: &v1.Probe{
						ProbeHandler: v1.ProbeHandler{
							TCPSocket: &v1.TCPSocketAction{
								Port: intstr.IntOrString{
									Type:   intstr.String,
									StrVal: "3306",
								},
							},
						},
						InitialDelaySeconds: 5,
						PeriodSeconds:       10,
					},
				},
			},
		},
	}
	namespace = &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "nano-pod-test",
		},
	}
)

func TestOverwritePatch(t *testing.T) {
	t.Run("TestOverwritePatch", func(t *testing.T) {
		err := k8sClient.Create(context.Background(), namespace)
		require.NoError(t, err)
		defer func() {
			_ = k8sClient.Delete(context.Background(), namespace)
		}()

		err = k8sClient.Create(context.Background(), nanoPod)

		nginxPodEncoded, err := json.Marshal(nginxPod)
		require.NoError(t, err)

		// the actual request we see in the webhook
		req := admission.Request{
			AdmissionRequest: admissionv1.AdmissionRequest{
				Namespace: namespace.Name,
				Object: runtime.RawExtension{
					Raw: nginxPodEncoded,
				},
			},
		}

		// the webhook handler
		decoder, err := admission.NewDecoder(scheme.Scheme)
		require.NoError(t, err)

		injector := NewHandler(k8sClient, logger, patcherfactory.NewBuilder().Register(nanopodv1.OverWritePatch, &patcher.OverWritePatcher{}).Build())
		err = injector.InjectDecoder(decoder)
		require.NoError(t, err)

		// test
		res := injector.Handle(context.Background(), req)
		logger.V(1).Info("admission response.", "patches", res.Patches)

		// verify
		assert.True(t, res.Allowed)
		assert.Nil(t, res.AdmissionResponse.Result)

	})

}
