package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

//FirstCrd is a top-level type
type FirstCrd struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +optional
	Status FirstCrdStatus `json:"status,omitempty"`
	// This is where you can define
	// your own custom spec
	Spec FirstCrdSpec `json:"spec,omitempty"`
}

//FirstCrdSpec is a custom spec
type FirstCrdSpec struct {
	Message string `json:"message,omitempty"`
}

//FirstCrdStatus is custom status
type FirstCrdStatus struct {
	Name string
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

//FirstCrdList struct no client needed for list as it's been created in above
type FirstCrdList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []FirstCrd `json:"items"`
}
