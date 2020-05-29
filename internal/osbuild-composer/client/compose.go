// Package client - compose contains functions for the compose API
// Copyright (C) 2020 by Red Hat, Inc.
package client

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/ondrejbudai/osbuild-image/internal/osbuild-composer/weldr"
)

// PostComposeV0 sends a JSON compose string to the API
// and returns an APIResponse
func PostComposeV0(socket *http.Client, compose string) (*APIResponse, error) {
	body, resp, err := PostJSON(socket, "/api/v0/compose", compose)
	if resp != nil || err != nil {
		return resp, err
	}
	return NewAPIResponse(body)
}

// GetComposeStatusV0 returns a list of composes matching the optional filter parameters
func GetComposeStatusV0(socket *http.Client, uuids, blueprint, status, composeType string) ([]weldr.ComposeEntryV0, *APIResponse, error) {
	// Build the query string
	route := "/api/v0/compose/status/" + uuids

	params := url.Values{}
	if len(blueprint) > 0 {
		params.Add("blueprint", blueprint)
	}
	if len(status) > 0 {
		params.Add("status", status)
	}
	if len(composeType) > 0 {
		params.Add("type", composeType)
	}

	if len(params) > 0 {
		route = route + "?" + params.Encode()
	}

	body, resp, err := GetRaw(socket, "GET", route)
	if resp != nil || err != nil {
		return []weldr.ComposeEntryV0{}, resp, err
	}
	var composes weldr.ComposeStatusResponseV0
	err = json.Unmarshal(body, &composes)
	if err != nil {
		return []weldr.ComposeEntryV0{}, nil, err
	}
	return composes.UUIDs, nil, nil
}

// GetComposeTypesV0 returns a list of the failed composes
func GetComposesTypesV0(socket *http.Client) ([]weldr.ComposeTypeV0, *APIResponse, error) {
	body, resp, err := GetRaw(socket, "GET", "/api/v0/compose/types")
	if resp != nil || err != nil {
		return []weldr.ComposeTypeV0{}, resp, err
	}
	var composeTypes weldr.ComposeTypesResponseV0
	err = json.Unmarshal(body, &composeTypes)
	if err != nil {
		return []weldr.ComposeTypeV0{}, nil, err
	}
	return composeTypes.Types, nil, nil
}

// DeleteComposeV0 deletes one or more composes based on their uuid
func DeleteComposeV0(socket *http.Client, uuids string) (weldr.DeleteComposeResponseV0, *APIResponse, error) {
	body, resp, err := DeleteRaw(socket, "/api/v0/compose/delete/"+uuids)
	if resp != nil || err != nil {
		return weldr.DeleteComposeResponseV0{}, resp, err
	}
	var deleteResponse weldr.DeleteComposeResponseV0
	err = json.Unmarshal(body, &deleteResponse)
	if err != nil {
		return weldr.DeleteComposeResponseV0{}, nil, err
	}
	return deleteResponse, nil, nil
}

// WriteComposeImageV0 requests the image for a compose and writes it to an io.Writer
func WriteComposeImageV0(socket *http.Client, w io.Writer, uuid string) (*APIResponse, error) {
	body, resp, err := GetRawBody(socket, "GET", "/api/v0/compose/image/"+uuid)
	if resp != nil || err != nil {
		return resp, err
	}
	_, err = io.Copy(w, body)
	body.Close()

	return nil, err
}

// WriteComposeLogV0 requests the log for a compose and writes it to an io.Writer
func WriteComposeLogV0(socket *http.Client, w io.Writer, uuid string) (*APIResponse, error) {
	body, resp, err := GetRawBody(socket, "GET", "/api/v0/compose/log/"+uuid)
	if resp != nil || err != nil {
		return resp, err
	}
	_, err = io.Copy(w, body)
	body.Close()

	return nil, err
}

// WriteComposeMetadataV0 requests the metadata for a compose and writes it to an io.Writer
func WriteComposeMetadataV0(socket *http.Client, w io.Writer, uuid string) (*APIResponse, error) {
	body, resp, err := GetRawBody(socket, "GET", "/api/v0/compose/metadata/"+uuid)
	if resp != nil || err != nil {
		return resp, err
	}
	_, err = io.Copy(w, body)
	body.Close()

	return nil, err
}
