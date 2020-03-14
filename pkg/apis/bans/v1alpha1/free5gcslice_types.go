package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// S-NSSAI
type Snssai struct {
	Sst int32  `json:"sst"`
	Sd  string `json:"sd"`
}

// Free5GCSliceSpec defines the desired state of Free5GCSlice
type Free5GCSliceSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// S-NSSAI list
	SnssaiList []Snssai `json:"snssaiList"`

	// gNodeB address
	GNBAddr string `json:"gNBAddr"`
}

// Free5GCSliceStatus defines the observed state of Free5GCSlice
type Free5GCSliceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html

	// State of free5GC Slice
	State string `json:"state"`

	// AMF address
	AmfAddr string `json:"amfAddr"`

	// UPF address
	UpfAddr string `json:"upfAddr"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Free5GCSlice is the Schema for the free5gcslice API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=free5gcslice,scope=Namespaced
type Free5GCSlice struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   Free5GCSliceSpec   `json:"spec,omitempty"`
	Status Free5GCSliceStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Free5GCSliceList contains a list of Free5GCSlice
type Free5GCSliceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Free5GCSlice `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Free5GCSlice{}, &Free5GCSliceList{})
}
