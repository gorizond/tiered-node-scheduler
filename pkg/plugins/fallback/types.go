package fallback

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	// DefaultFallbackTimeoutSeconds specifies the default timeout before allowing scheduling on fallback nodes.
	DefaultFallbackTimeoutSeconds = 20

	// DefaultPrimaryNodeLabelKey specifies the default label key used to identify primary nodes.
	DefaultPrimaryNodeLabelKey = "pool"

	// DefaultPrimaryNodeLabelValue specifies the default label value used to identify primary nodes.
	DefaultPrimaryNodeLabelValue = "primary"
)

// Args defines the configuration parameters for the TieredNodeFallback plugin.
type Args struct {
	metav1.TypeMeta `json:",inline"`

	// FallbackTimeoutSeconds is the duration in seconds to wait before allowing pod placement on fallback nodes.
	FallbackTimeoutSeconds int `json:"fallbackTimeoutSeconds,omitempty"`

	// PrimaryNodeLabelKey is the label key identifying primary nodes.
	PrimaryNodeLabelKey string `json:"primaryNodeLabelKey,omitempty"`

	// PrimaryNodeLabelValue is the label value identifying primary nodes.
	PrimaryNodeLabelValue string `json:"primaryNodeLabelValue,omitempty"`
}

// DeepCopyObject returns a generically typed copy of an object.
func (a *Args) DeepCopyObject() runtime.Object {
	if a == nil {
		return nil
	}
	out := new(Args)
	a.DeepCopyInto(out)
	return out
}

// DeepCopyInto copies all properties of this object into another object of the same type.
func (a *Args) DeepCopyInto(out *Args) {
	*out = *a
	out.TypeMeta = a.TypeMeta
}

// SetDefaults assigns default values to unset fields.
func (a *Args) SetDefaults() {
	if a.FallbackTimeoutSeconds <= 0 {
		a.FallbackTimeoutSeconds = DefaultFallbackTimeoutSeconds
	}
	if a.PrimaryNodeLabelKey == "" {
		a.PrimaryNodeLabelKey = DefaultPrimaryNodeLabelKey
	}
	if a.PrimaryNodeLabelValue == "" {
		a.PrimaryNodeLabelValue = DefaultPrimaryNodeLabelValue
	}
}

// Validate verifies configuration integrity.
func (a *Args) Validate() error {
	if a.FallbackTimeoutSeconds < 0 {
		return fmt.Errorf("fallbackTimeoutSeconds cannot be negative, got %d", a.FallbackTimeoutSeconds)
	}
	if a.PrimaryNodeLabelKey == "" {
		return fmt.Errorf("primaryNodeLabelKey cannot be empty")
	}
	if a.PrimaryNodeLabelValue == "" {
		return fmt.Errorf("primaryNodeLabelValue cannot be empty")
	}
	return nil
}
