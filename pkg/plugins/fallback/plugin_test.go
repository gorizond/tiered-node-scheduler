package fallback

import (
	"context"
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

func TestTieredNodeFallback_Filter(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name       string
		args       *Args
		pod        *v1.Pod
		node       *v1.Node
		wantStatus *framework.Status
	}{
		{
			name: "Primary node placement succeeds immediately before timeout",
			args: &Args{
				FallbackTimeoutSeconds: 20,
				PrimaryNodeLabelKey:    "pool",
				PrimaryNodeLabelValue:  "primary",
			},
			pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "pod-fresh",
					Namespace:         "default",
					CreationTimestamp: metav1.NewTime(now.Add(-5 * time.Second)),
				},
			},
			node: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-primary-1",
					Labels: map[string]string{
						"pool": "primary",
					},
				},
			},
			wantStatus: framework.NewStatus(framework.Success),
		},
		{
			name: "Secondary node placement is rejected when age < fallbackTimeoutSeconds",
			args: &Args{
				FallbackTimeoutSeconds: 20,
				PrimaryNodeLabelKey:    "pool",
				PrimaryNodeLabelValue:  "primary",
			},
			pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "pod-young",
					Namespace:         "default",
					CreationTimestamp: metav1.NewTime(now.Add(-10 * time.Second)),
				},
			},
			node: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-fallback-1",
					Labels: map[string]string{
						"pool": "secondary",
					},
				},
			},
			wantStatus: framework.NewStatus(framework.Unschedulable),
		},
		{
			name: "Secondary node placement succeeds when age >= fallbackTimeoutSeconds",
			args: &Args{
				FallbackTimeoutSeconds: 20,
				PrimaryNodeLabelKey:    "pool",
				PrimaryNodeLabelValue:  "primary",
			},
			pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "pod-aged",
					Namespace:         "default",
					CreationTimestamp: metav1.NewTime(now.Add(-25 * time.Second)),
				},
			},
			node: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-fallback-2",
					Labels: map[string]string{
						"pool": "secondary",
					},
				},
			},
			wantStatus: framework.NewStatus(framework.Success),
		},
		{
			name: "Node with nil/empty labels treated as secondary, rejected before timeout",
			args: &Args{
				FallbackTimeoutSeconds: 20,
				PrimaryNodeLabelKey:    "pool",
				PrimaryNodeLabelValue:  "primary",
			},
			pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "pod-young-unlabeled-node",
					Namespace:         "default",
					CreationTimestamp: metav1.NewTime(now.Add(-2 * time.Second)),
				},
			},
			node: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-unlabeled",
				},
			},
			wantStatus: framework.NewStatus(framework.Unschedulable),
		},
		{
			name: "Node with nil/empty labels treated as secondary, accepted after timeout",
			args: &Args{
				FallbackTimeoutSeconds: 20,
				PrimaryNodeLabelKey:    "pool",
				PrimaryNodeLabelValue:  "primary",
			},
			pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "pod-aged-unlabeled-node",
					Namespace:         "default",
					CreationTimestamp: metav1.NewTime(now.Add(-21 * time.Second)),
				},
			},
			node: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-unlabeled",
				},
			},
			wantStatus: framework.NewStatus(framework.Success),
		},
		{
			name: "Pod with zero CreationTimestamp rejected on secondary node",
			args: &Args{
				FallbackTimeoutSeconds: 20,
				PrimaryNodeLabelKey:    "pool",
				PrimaryNodeLabelValue:  "primary",
			},
			pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod-zero-creation",
					Namespace: "default",
				},
			},
			node: &v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-fallback-3",
					Labels: map[string]string{
						"pool": "fallback",
					},
				},
			},
			wantStatus: framework.NewStatus(framework.Unschedulable),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.SetDefaults()
			pl := &TieredNodeFallback{
				args: tt.args,
			}

			nodeInfo := framework.NewNodeInfo()
			nodeInfo.SetNode(tt.node)

			status := pl.Filter(context.Background(), nil, tt.pod, nodeInfo)
			if status.Code() != tt.wantStatus.Code() {
				t.Errorf("Filter() code = %v, want %v (Message: %s)", status.Code(), tt.wantStatus.Code(), status.Message())
			}
		})
	}
}

func TestArgs_ValidationAndDefaults(t *testing.T) {
	t.Run("Default assignments", func(t *testing.T) {
		args := &Args{}
		args.SetDefaults()

		if args.FallbackTimeoutSeconds != DefaultFallbackTimeoutSeconds {
			t.Errorf("expected fallbackTimeoutSeconds %d, got %d", DefaultFallbackTimeoutSeconds, args.FallbackTimeoutSeconds)
		}
		if args.PrimaryNodeLabelKey != DefaultPrimaryNodeLabelKey {
			t.Errorf("expected primaryNodeLabelKey %s, got %s", DefaultPrimaryNodeLabelKey, args.PrimaryNodeLabelKey)
		}
		if args.PrimaryNodeLabelValue != DefaultPrimaryNodeLabelValue {
			t.Errorf("expected primaryNodeLabelValue %s, got %s", DefaultPrimaryNodeLabelValue, args.PrimaryNodeLabelValue)
		}
	})

	t.Run("Validation failure on negative timeout", func(t *testing.T) {
		args := &Args{
			FallbackTimeoutSeconds: -1,
			PrimaryNodeLabelKey:    "tier",
			PrimaryNodeLabelValue:  "gold",
		}
		if err := args.Validate(); err == nil {
			t.Errorf("expected error for negative timeout, got nil")
		}
	})
}
