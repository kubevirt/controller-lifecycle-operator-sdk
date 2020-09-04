package crdgenerator

import (
	"github.com/kubevirt/controller-lifecycle-operator-sdk/pkg/sdk"
	"golang.org/x/tools/go/packages"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-tools/pkg/loader"
	"sigs.k8s.io/controller-tools/pkg/markers"

	crdgen "sigs.k8s.io/controller-tools/pkg/crd"
	crdmarkers "sigs.k8s.io/controller-tools/pkg/crd/markers"
)

// CreateCRD creates CRD with given Group Kind and marks its version with given name based on sources find in the working directory
func CreateCRD(groupKind schema.GroupKind, crVersionName string) *extv1.CustomResourceDefinition {
	pkgs, err := loader.LoadRoots(sdk.GetOperatorToplevel() + "/...")
	if err != nil {
		panic(err)
	}
	reg := &markers.Registry{}
	crdmarkers.Register(reg)

	parser := &crdgen.Parser{
		Collector: &markers.Collector{Registry: reg},
		Checker:   &loader.TypeChecker{},
	}
	crdgen.AddKnownTypes(parser)
	if len(pkgs) == 0 {
		panic("Failed identifying packages")
	}
	for _, p := range pkgs {
		parser.NeedPackage(p)
	}
	parser.NeedCRDFor(groupKind, nil)
	for _, p := range pkgs {
		err = PackageErrors(p, packages.TypeError)
		if err != nil {
			panic(err)
		}
	}
	c := parser.CustomResourceDefinitions[groupKind]
	// enforce validation of CR name to prevent multiple CRs

	for _, v := range c.Spec.Versions {
		v.Schema.OpenAPIV3Schema.Properties["metadata"] = extv1.JSONSchemaProps{
			Type: "object",
			Properties: map[string]extv1.JSONSchemaProps{
				"name": {
					Type:    "string",
					Pattern: crVersionName,
				},
			},
		}
	}

	return &c
}

func PackageErrors(pkg *loader.Package, filterKinds ...packages.ErrorKind) error {
	toSkip := make(map[packages.ErrorKind]struct{})
	for _, errKind := range filterKinds {
		toSkip[errKind] = struct{}{}
	}
	var outErr error
	packages.Visit([]*packages.Package{pkg.Package}, nil, func(pkgRaw *packages.Package) {
		for _, err := range pkgRaw.Errors {
			if _, skip := toSkip[err.Kind]; skip {
				continue
			}
			outErr = err
		}
	})
	return outErr
}
