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

package v1beta1

import (
	"testing"

	"github.com/Azure/go-autorest/autorest/to"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestAzureManagedMachinePoolDefaultingWebhook(t *testing.T) {
	g := NewWithT(t)

	t.Logf("Testing ammp defaulting webhook with mode system")
	ammp := &AzureManagedMachinePool{
		ObjectMeta: metav1.ObjectMeta{
			Name: "fooName",
		},
		Spec: AzureManagedMachinePoolSpec{
			Mode:         "System",
			SKU:          "StandardD2S_V3",
			OSDiskSizeGB: to.Int32Ptr(512),
		},
	}
	var client client.Client
	ammp.Default(client)
	g.Expect(ammp.Labels).ToNot(BeNil())
	val, ok := ammp.Labels[LabelAgentPoolMode]
	g.Expect(ok).To(BeTrue())
	g.Expect(val).To(Equal("System"))
	g.Expect(*ammp.Spec.Name).To(Equal("fooName"))

	t.Logf("Testing ammp defaulting webhook with empty string name specified in Spec")
	emptyName := ""
	ammp.Spec.Name = &emptyName
	ammp.Default(client)
	g.Expect(*ammp.Spec.Name).To(Equal("fooName"))

	t.Logf("Testing ammp defaulting webhook with normal name specified in Spec")
	normalName := "barName"
	ammp.Spec.Name = &normalName
	ammp.Default(client)
	g.Expect(*ammp.Spec.Name).To(Equal("barName"))
}

func TestAzureManagedMachinePoolValidateCreateWebhook(t *testing.T) {
	g := NewWithT(t)

	t.Logf("Testing ValidateCreate webhook")

	tests := []struct {
		name    string
		pool    *AzureManagedMachinePool
		wantErr bool
	}{
		{
			name: "AutoScaling enabled, expect both MinCount and MaxCount to be set",
			pool: &AzureManagedMachinePool{
				Spec: AzureManagedMachinePoolSpec{
					Mode:         "System",
					SKU:          "StandardD2S_V3",
					OSDiskSizeGB: to.Int32Ptr(512),
					AutoScaling: &AutoScaling{
						MinCount: to.Int32Ptr(2),
						MaxCount: to.Int32Ptr(5),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Node taint should follow patten key=value:NoSchedule",
			pool: &AzureManagedMachinePool{
				Spec: AzureManagedMachinePoolSpec{
					Mode:         "User",
					SKU:          "StandardD2S_V3",
					OSDiskSizeGB: to.Int32Ptr(512),
					NodeTaints:   []string{"key1=value1:NoSchedule"},
				},
			},
			wantErr: false,
		},
		{
			name: "Node taint should follow patten key=value:NoSchedule",
			pool: &AzureManagedMachinePool{
				Spec: AzureManagedMachinePoolSpec{
					Mode:         "User",
					SKU:          "StandardD2S_V3",
					OSDiskSizeGB: to.Int32Ptr(512),
					NodeTaints:   []string{"key1=value1NoSchedule"},
				},
			},
			wantErr: true,
		},
		{
			name: "AllowedUnsafeSysctls should be one of \"kernel.shm*\", \"kernel.msg*\", \"kernel.sem\", \"fs.mqueue.*\", \"net.*\"",
			pool: &AzureManagedMachinePool{
				Spec: AzureManagedMachinePoolSpec{
					Mode:          "User",
					SKU:           "StandardD2S_V3",
					OSDiskSizeGB:  to.Int32Ptr(512),
					KubeletConfig: &KubeletConfig{AllowedUnsafeSysctls: &[]string{"net.*"}},
				},
			},
			wantErr: false,
		},
		{
			name: "AllowedUnsafeSysctls should be one of \"kernel.shm*\", \"kernel.msg*\", \"kernel.sem\", \"fs.mqueue.*\", \"net.*\"",
			pool: &AzureManagedMachinePool{
				Spec: AzureManagedMachinePoolSpec{
					Mode:          "User",
					SKU:           "StandardD2S_V3",
					OSDiskSizeGB:  to.Int32Ptr(512),
					KubeletConfig: &KubeletConfig{AllowedUnsafeSysctls: &[]string{"net4.*"}},
				},
			},
			wantErr: true,
		},
	}

	var client client.Client
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.pool.ValidateCreate(client)
			if tc.wantErr {
				g.Expect(err).To(HaveOccurred())
			} else {
				g.Expect(err).NotTo(HaveOccurred())
			}
		})
	}
}

func TestAzureManagedMachinePoolUpdatingWebhook(t *testing.T) {
	g := NewWithT(t)

	t.Logf("Testing ammp updating webhook with mode system")

	tests := []struct {
		name    string
		new     *AzureManagedMachinePool
		old     *AzureManagedMachinePool
		wantErr bool
	}{
		{
			name: "Cannot change SKU of the agentpool",
			new: &AzureManagedMachinePool{
				Spec: AzureManagedMachinePoolSpec{
					Mode:         "System",
					SKU:          "StandardD2S_V3",
					OSDiskSizeGB: to.Int32Ptr(512),
				},
			},
			old: &AzureManagedMachinePool{
				Spec: AzureManagedMachinePoolSpec{
					Mode:         "System",
					SKU:          "StandardD2S_V4",
					OSDiskSizeGB: to.Int32Ptr(512),
				},
			},
			wantErr: true,
		},
		{
			name: "Cannot change OSDiskSizeGB of the agentpool",
			new: &AzureManagedMachinePool{
				Spec: AzureManagedMachinePoolSpec{
					Mode:         "System",
					SKU:          "StandardD2S_V3",
					OSDiskSizeGB: to.Int32Ptr(512),
				},
			},
			old: &AzureManagedMachinePool{
				Spec: AzureManagedMachinePoolSpec{
					Mode:         "System",
					SKU:          "StandardD2S_V3",
					OSDiskSizeGB: to.Int32Ptr(1024),
				},
			},
			wantErr: true,
		},
		{
			name: "Cannot change EnableFIPS of the agentpool",
			new: &AzureManagedMachinePool{
				Spec: AzureManagedMachinePoolSpec{
					Mode:         "System",
					SKU:          "StandardD2S_V3",
					OSDiskSizeGB: to.Int32Ptr(512),
					EnableFIPS:   to.BoolPtr(true),
				},
			},
			old: &AzureManagedMachinePool{
				Spec: AzureManagedMachinePoolSpec{
					Mode:         "System",
					SKU:          "StandardD2S_V3",
					OSDiskSizeGB: to.Int32Ptr(512),
					EnableFIPS:   to.BoolPtr(false),
				},
			},
			wantErr: true,
		},
		{
			name: "Cannot change EnableNodePublicIP of the agentpool",
			new: &AzureManagedMachinePool{
				Spec: AzureManagedMachinePoolSpec{
					Mode:               "System",
					SKU:                "StandardD2S_V3",
					OSDiskSizeGB:       to.Int32Ptr(512),
					EnableNodePublicIP: to.BoolPtr(true),
				},
			},
			old: &AzureManagedMachinePool{
				Spec: AzureManagedMachinePoolSpec{
					Mode:               "System",
					SKU:                "StandardD2S_V3",
					OSDiskSizeGB:       to.Int32Ptr(512),
					EnableNodePublicIP: to.BoolPtr(false),
				},
			},
			wantErr: true,
		},
		{
			name: "Cannot change OsDiskType of the agentpool",
			new: &AzureManagedMachinePool{
				Spec: AzureManagedMachinePoolSpec{
					Mode:         "System",
					SKU:          "StandardD2S_V3",
					OSDiskSizeGB: to.Int32Ptr(512),
					OsDiskType:   to.StringPtr("Managed"),
				},
			},
			old: &AzureManagedMachinePool{
				Spec: AzureManagedMachinePoolSpec{
					Mode:         "System",
					SKU:          "StandardD2S_V3",
					OSDiskSizeGB: to.Int32Ptr(512),
					OsDiskType:   to.StringPtr("Ephemeral"),
				},
			},
			wantErr: true,
		},
		{
			name: "Cannot change ScaleSetPriority of the agentpool",
			new: &AzureManagedMachinePool{
				Spec: AzureManagedMachinePoolSpec{
					Mode:             "System",
					SKU:              "StandardD2S_V3",
					OSDiskSizeGB:     to.Int32Ptr(512),
					ScaleSetPriority: to.StringPtr("Regular"),
				},
			},
			old: &AzureManagedMachinePool{
				Spec: AzureManagedMachinePoolSpec{
					Mode:             "System",
					SKU:              "StandardD2S_V3",
					OSDiskSizeGB:     to.Int32Ptr(512),
					ScaleSetPriority: to.StringPtr("Spot"),
				},
			},
			wantErr: true,
		},
		{
			name: "Cannot change MaxPods of the agentpool",
			new: &AzureManagedMachinePool{
				Spec: AzureManagedMachinePoolSpec{
					Mode:         "System",
					SKU:          "StandardD2S_V3",
					OSDiskSizeGB: to.Int32Ptr(512),
					MaxPods:      to.Int32Ptr(50),
				},
			},
			old: &AzureManagedMachinePool{
				Spec: AzureManagedMachinePoolSpec{
					Mode:         "System",
					SKU:          "StandardD2S_V3",
					OSDiskSizeGB: to.Int32Ptr(512),
					MaxPods:      to.Int32Ptr(40),
				},
			},
			wantErr: true,
		},
		{
			name: "Cannot change NodeTaints of the agentpool",
			new: &AzureManagedMachinePool{
				Spec: AzureManagedMachinePoolSpec{
					Mode:         "System",
					SKU:          "StandardD2S_V3",
					OSDiskSizeGB: to.Int32Ptr(512),
					NodeTaints:   []string{"key1=value1:NoSchedule"},
				},
			},
			old: &AzureManagedMachinePool{
				Spec: AzureManagedMachinePoolSpec{
					Mode:         "System",
					SKU:          "StandardD2S_V3",
					OSDiskSizeGB: to.Int32Ptr(512),
					NodeTaints:   []string{"key2=value2:NoSchedule"},
				},
			},
			wantErr: true,
		},
	}
	var client client.Client
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.new.ValidateUpdate(tc.old, client)
			if tc.wantErr {
				g.Expect(err).To(HaveOccurred())
			} else {
				g.Expect(err).NotTo(HaveOccurred())
			}
		})
	}
}
