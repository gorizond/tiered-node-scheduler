package fallback

import (
	"context"
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/scheduler/framework"
	frameworkruntime "k8s.io/kubernetes/pkg/scheduler/framework/runtime"
)

// Name is the name of the plugin used in the plugin registry and configuration profiles.
const Name = "TieredNodeFallback"

// TieredNodeFallback is a FilterPlugin prioritizing the primary node pool and falling back to secondary nodes after a timeout.
type TieredNodeFallback struct {
	handle framework.Handle
	args   *Args
}

var _ framework.FilterPlugin = &TieredNodeFallback{}

// Name returns the name of the plugin.
func (pl *TieredNodeFallback) Name() string {
	return Name
}

// New initializes a new instance of the TieredNodeFallback plugin.
func New(ctx context.Context, obj runtime.Object, handle framework.Handle) (framework.Plugin, error) {
	args := &Args{}
	if obj != nil {
		if err := frameworkruntime.DecodeInto(obj, args); err != nil {
			return nil, fmt.Errorf("failed to decode configuration for plugin %s: %w", Name, err)
		}
	}

	args.SetDefaults()
	if err := args.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration for plugin %s: %w", Name, err)
	}

	klog.V(2).InfoS("Initialized TieredNodeFallback plugin",
		"fallbackTimeoutSeconds", args.FallbackTimeoutSeconds,
		"primaryNodeLabelKey", args.PrimaryNodeLabelKey,
		"primaryNodeLabelValue", args.PrimaryNodeLabelValue,
	)

	return &TieredNodeFallback{
		handle: handle,
		args:   args,
	}, nil
}

// Filter evaluates whether a pod is allowed to run on the specified candidate node.
func (pl *TieredNodeFallback) Filter(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {
	if nodeInfo == nil || nodeInfo.Node() == nil {
		return framework.NewStatus(framework.Error, "node not found in NodeInfo")
	}

	node := nodeInfo.Node()

	// 1. Primary node evaluation: immediately allow placement
	if val, ok := node.Labels[pl.args.PrimaryNodeLabelKey]; ok && val == pl.args.PrimaryNodeLabelValue {
		klog.V(5).InfoS("Pod scheduled onto primary pool node",
			"pod", klog.KObj(pod),
			"node", node.Name,
		)
		return framework.NewStatus(framework.Success)
	}

	// 2. Fallback node evaluation: check creation timestamp
	if pod.CreationTimestamp.IsZero() {
		klog.V(4).InfoS("Pod has zero CreationTimestamp, blocking placement on fallback node",
			"pod", klog.KObj(pod),
			"node", node.Name,
		)
		return framework.NewStatus(framework.Unschedulable, "pod has zero creation timestamp, grace period active")
	}

	podAge := time.Since(pod.CreationTimestamp.Time)
	gracePeriod := time.Duration(pl.args.FallbackTimeoutSeconds) * time.Second

	if podAge < gracePeriod {
		remaining := (gracePeriod - podAge).Truncate(time.Second)
		klog.V(4).InfoS("Pod is within primary-only grace period; rejecting fallback node",
			"pod", klog.KObj(pod),
			"node", node.Name,
			"age", podAge.Truncate(time.Millisecond),
			"remainingGracePeriod", remaining,
		)
		return framework.NewStatus(
			framework.Unschedulable,
			fmt.Sprintf("pod is in primary-only scheduling grace period (age: %s, required: %s)", podAge.Truncate(time.Second), gracePeriod),
		)
	}

	klog.V(4).InfoS("Grace period elapsed; permitting pod placement onto fallback node",
		"pod", klog.KObj(pod),
		"node", node.Name,
		"age", podAge.Truncate(time.Second),
	)

	return framework.NewStatus(framework.Success)
}
