// HO.

package main

import (
	"os"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/pkg/errors"
	"github.com/tufin/oasdiff/diff"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiyaml "k8s.io/apimachinery/pkg/util/yaml"
	k8syaml "sigs.k8s.io/yaml"
)

func getCRD(m string) (*v1.CustomResourceDefinition, error) {
	crd := &v1.CustomResourceDefinition{}
	buff, err := os.ReadFile(m)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load the CRD manifest from file: %s", m)
	}
	if err := apiyaml.Unmarshal(buff, crd); err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal CRD manifest from file: %s", m)
	}
	return crd, nil
}

func getOpenAPIv3Document(crd *v1.CustomResourceDefinition) (*openapi3.T, error) {
	if len(crd.Spec.Versions) != 1 {
		return nil, errors.New("invalid CRD manifest: Only CRDs with exactly one version are supported")
	}
	if crd.Spec.Versions[0].Schema == nil || crd.Spec.Versions[0].Schema.OpenAPIV3Schema == nil {
		return nil, errors.New("invalid CRD manifest: CRD's .Spec.Versions[0].Schema.OpenAPIV3Schema cannot be nil")
	}

	t := &openapi3.T{
		Info:  &openapi3.Info{},
		Paths: make(openapi3.Paths),
	}
	c := make(openapi3.Content)
	t.Paths["/crd"] = &openapi3.PathItem{
		Put: &openapi3.Operation{
			RequestBody: &openapi3.RequestBodyRef{
				Value: &openapi3.RequestBody{
					Content: c,
				},
			},
		},
	}
	s := &openapi3.Schema{}
	c["application/json"] = &openapi3.MediaType{
		Schema: &openapi3.SchemaRef{
			Value: s,
		},
	}

	// convert from CRD validation schema to openAPI v3 schema
	buff, err := k8syaml.Marshal(crd.Spec.Versions[0].Schema.OpenAPIV3Schema)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal CRD validation schema")
	}
	if err := k8syaml.Unmarshal(buff, s); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal CRD validation schema into openAPI v3 schema")
	}
	return t, nil
}

func getBreakingChanges(baseDoc, revisionDoc *openapi3.T) (*diff.Diff, error) {
	config := diff.NewConfig()
	// currently we only need to detect breaking API changes
	config.BreakingOnly = true
	d, err := diff.Get(config, baseDoc, revisionDoc)
	return d, errors.Wrap(err, "failed to compute breaking changes in base and revision CRD schemas")
}
