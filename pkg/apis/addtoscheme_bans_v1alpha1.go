package apis

import (
	"github.com/stevenchiu30801/free5gc-operator/pkg/apis/bans/v1alpha1"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes, v1alpha1.SchemeBuilder.AddToScheme)
}
