package remotestate_test

import (
	"encoding/json"
	"testing"

	"errors"

	"github.com/gruntwork-io/terragrunt/internal/remotestate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseTerraformStateLocal(t *testing.T) {
	t.Parallel()

	stateFile :=
		`
	{
		"version": 1,
		"serial": 0,
		"modules": [
			{
				"path": [
					"root"
				],
				"outputs": {},
				"resources": {}
			}
		]
	}
	`

	expectedTerraformState := &remotestate.TerraformState{
		Version: 1,
		Serial:  0,
		Backend: nil,
		Modules: []remotestate.TerraformStateModule{
			{
				Path:      []string{"root"},
				Outputs:   map[string]any{},
				Resources: map[string]any{},
			},
		},
	}

	actualTerraformState, err := remotestate.ParseTerraformState([]byte(stateFile))

	require.NoError(t, err)
	assert.Equal(t, expectedTerraformState, actualTerraformState)
	assert.False(t, actualTerraformState.IsRemote())
}

func TestParseTerraformStateRemote(t *testing.T) {
	t.Parallel()

	stateFile :=
		`
	{
		"version": 5,
		"serial": 12,
		"backend": {
			"type": "s3",
			"config": {
				"bucket": "bucket",
				"encrypt": true,
				"key": "experiment-1.tfstate",
				"region": "us-east-1"
			}
		},
		"modules": [
			{
				"path": [
					"root"
				],
				"outputs": {},
				"resources": {}
			}
		]
	}
	`

	expectedTerraformState := &remotestate.TerraformState{
		Version: 5,
		Serial:  12,
		Backend: &remotestate.TerraformBackend{
			Type: "s3",
			Config: map[string]any{
				"bucket":  "bucket",
				"encrypt": true,
				"key":     "experiment-1.tfstate",
				"region":  "us-east-1",
			},
		},
		Modules: []remotestate.TerraformStateModule{
			{
				Path:      []string{"root"},
				Outputs:   map[string]any{},
				Resources: map[string]any{},
			},
		},
	}

	actualTerraformState, err := remotestate.ParseTerraformState([]byte(stateFile))

	require.NoError(t, err)
	assert.Equal(t, expectedTerraformState, actualTerraformState)
	assert.True(t, actualTerraformState.IsRemote())
}

func TestParseTerraformStateRemoteFull(t *testing.T) {
	t.Parallel()

	// This is a small snippet (with lots of editing) of Terraform templates that created a VPC
	stateFile :=
		`
	{
	    "version": 1,
	    "serial": 51,
	    "backend": {
		"type": "s3",
		"config": {
		    "bucket": "bucket",
		    "encrypt": true,
		    "key": "terraform.tfstate",
		    "region": "us-east-1"
		}
	    },
	    "modules": [
		{
		    "path": [
			"root"
		    ],
		    "outputs": {
			"key1": "value1",
			"key2": "value2",
			"key3": "value3"
		    },
		    "resources": {}
		},
		{
		    "path": [
			"root",
			"module_with_outputs_no_resources"
		    ],
		    "outputs": {
			"key1": "",
			"key2": ""
		    },
		    "resources": {}
		},
		{
		    "path": [
			"root",
			"module_with_resources_no_outputs"
		    ],
		    "outputs": {},
		    "resources": {
			"aws_eip.nat.0": {
			    "type": "aws_eip",
			    "depends_on": [
				"aws_internet_gateway.main"
			    ],
			    "primary": {
				"id": "eipalloc-b421becd",
				"attributes": {
				    "association_id": "",
				    "domain": "vpc",
				    "id": "eipalloc-b421becd",
				    "instance": "",
				    "network_interface": "",
				    "private_ip": "",
				    "public_ip": "23.20.182.117",
				    "vpc": "true"
				}
			    }
			},
			"aws_eip.nat.1": {
			    "type": "aws_eip",
			    "depends_on": [
				"aws_internet_gateway.main"
			    ],
			    "primary": {
				"id": "eipalloc-95d846ec",
				"attributes": {
				    "association_id": "",
				    "domain": "vpc",
				    "id": "eipalloc-95d846ec",
				    "instance": "",
				    "network_interface": "",
				    "private_ip": "",
				    "public_ip": "52.21.82.253",
				    "vpc": "true"
				}
			    }
			}
		    }
		},
		{
		    "path": [
			"root",
			"module_level_1",
			"module_level_2"
		    ],
		    "outputs": {},
		    "resources": {}
		}
	    ]
	}

	`

	expectedTerraformState := &remotestate.TerraformState{
		Version: 1,
		Serial:  51,
		Backend: &remotestate.TerraformBackend{
			Type: "s3",
			Config: map[string]any{
				"bucket":  "bucket",
				"encrypt": true,
				"key":     "terraform.tfstate",
				"region":  "us-east-1",
			},
		},
		Modules: []remotestate.TerraformStateModule{
			{
				Path: []string{"root"},
				Outputs: map[string]any{
					"key1": "value1",
					"key2": "value2",
					"key3": "value3",
				},
				Resources: map[string]any{},
			},
			{
				Path: []string{"root", "module_with_outputs_no_resources"},
				Outputs: map[string]any{
					"key1": "",
					"key2": "",
				},
				Resources: map[string]any{},
			},
			{
				Path:    []string{"root", "module_with_resources_no_outputs"},
				Outputs: map[string]any{},
				Resources: map[string]any{
					"aws_eip.nat.0": map[string]any{
						"type":       "aws_eip",
						"depends_on": []any{"aws_internet_gateway.main"},
						"primary": map[string]any{
							"id": "eipalloc-b421becd",
							"attributes": map[string]any{
								"association_id":    "",
								"domain":            "vpc",
								"id":                "eipalloc-b421becd",
								"instance":          "",
								"network_interface": "",
								"private_ip":        "",
								"public_ip":         "23.20.182.117",
								"vpc":               "true",
							},
						},
					},
					"aws_eip.nat.1": map[string]any{
						"type":       "aws_eip",
						"depends_on": []any{"aws_internet_gateway.main"},
						"primary": map[string]any{
							"id": "eipalloc-95d846ec",
							"attributes": map[string]any{
								"association_id":    "",
								"domain":            "vpc",
								"id":                "eipalloc-95d846ec",
								"instance":          "",
								"network_interface": "",
								"private_ip":        "",
								"public_ip":         "52.21.82.253",
								"vpc":               "true",
							},
						},
					},
				},
			},
			{
				Path:      []string{"root", "module_level_1", "module_level_2"},
				Outputs:   map[string]any{},
				Resources: map[string]any{},
			},
		},
	}

	actualTerraformState, err := remotestate.ParseTerraformState([]byte(stateFile))

	require.NoError(t, err)
	assert.Equal(t, expectedTerraformState, actualTerraformState)
	assert.True(t, actualTerraformState.IsRemote())
}

func TestParseTerraformStateEmpty(t *testing.T) {
	t.Parallel()

	stateFile := `{}`

	expectedTerraformState := &remotestate.TerraformState{}

	actualTerraformState, err := remotestate.ParseTerraformState([]byte(stateFile))

	require.NoError(t, err)
	assert.Equal(t, expectedTerraformState, actualTerraformState)
	assert.False(t, actualTerraformState.IsRemote())
}

func TestParseTerraformStateInvalid(t *testing.T) {
	t.Parallel()

	stateFile := `not-valid-json`

	actualTerraformState, err := remotestate.ParseTerraformState([]byte(stateFile))

	assert.Nil(t, actualTerraformState)
	require.Error(t, err)

	var jsonSyntaxError *json.SyntaxError
	ok := errors.As(err, &jsonSyntaxError)
	assert.True(t, ok)
}
