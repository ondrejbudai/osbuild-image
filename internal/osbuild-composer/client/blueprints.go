// Package client - blueprints contains functions for the blueprint API
// Copyright (C) 2020 by Red Hat, Inc.
package client

import (
	"net/http"
)

// PostTOMLBlueprintV0 sends a TOML blueprint string to the API
// and returns an APIResponse
func PostTOMLBlueprintV0(socket *http.Client, blueprint string) (*APIResponse, error) {
	body, resp, err := PostTOML(socket, "/api/v0/blueprints/new", blueprint)
	if resp != nil || err != nil {
		return resp, err
	}
	return NewAPIResponse(body)
}

// PostJSONBlueprintV0 sends a JSON blueprint string to the API
// and returns an APIResponse
func PostJSONBlueprintV0(socket *http.Client, blueprint string) (*APIResponse, error) {
	body, resp, err := PostJSON(socket, "/api/v0/blueprints/new", blueprint)
	if resp != nil || err != nil {
		return resp, err
	}
	return NewAPIResponse(body)
}

// DeleteBlueprintV0 deletes the named blueprint and returns an APIResponse
func DeleteBlueprintV0(socket *http.Client, bpName string) (*APIResponse, error) {
	body, resp, err := DeleteRaw(socket, "/api/v0/blueprints/delete/"+bpName)
	if resp != nil || err != nil {
		return resp, err
	}
	return NewAPIResponse(body)
}
