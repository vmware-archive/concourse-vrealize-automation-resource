// Copyright 2020 program was created by VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package resource

// VRASource holds the source configuration
type VRASource struct {
	Host     string `json:"host"`
	Pipeline string `json:"pipeline"`
	APIToken string `json:"apiToken"`
}

// VRAVersion holds the version info
type VRAVersion struct {
	Value string `json:"value"`
}

// VRAResource holds the resource type configuration
type VRAResource struct {
	Src       *VRASource
	Ver       *VRAVersion
	OutParams *OutParams
}

// OutParams holds the out task params
type OutParams struct {
	Wait      bool   `json:"wait"`
	Changeset string `json:"changeset"`
}

type MetadataField struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Source returns pointer to the source definition struct
func (r *VRAResource) Source() interface{} {
	return r.Src
}

// Version returns pointer to the version definition struct
func (r *VRAResource) Version() interface{} {
	return r.Ver
}

// Params returns pointer to the Out params definition struct
func (r *VRAResource) Params() (params interface{}) {
	return r.OutParams
}

// Out Puts the resource and returns the new version and metadata
func (r *VRAResource) Out(dir string) (version interface{}, metadata []interface{}, err error) {
	return out(*r.Src, *r.OutParams)
}
