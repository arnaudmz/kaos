package v1

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KaosRule describes a Kaos Rule.
type KaosRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec KaosRuleSpec `json:"spec"`
}

// KaosRuleSpec is the spec for a KaosRule resource
type KaosRuleSpec struct {
	Cron        string                `json:"cron"`
	PodSelector *metav1.LabelSelector `json:"podselector,omitempty" protobuf:"bytes,2,opt,name=podselector"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KaosRuleList is a list of KaosRule resources
type KaosRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []KaosRule `json:"items"`
}

// String simply prints kr id
func (kr *KaosRule) String() string {
	return fmt.Sprintf("KaosRule (%s/%s)", kr.ObjectMeta.Namespace, kr.ObjectMeta.Name)
}
