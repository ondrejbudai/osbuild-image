// Package weldr - json contains Exported API request/response structures
// Copyright (C) 2020 by Red Hat, Inc.
package weldr

import (
	"github.com/google/uuid"

	"github.com/ondrejbudai/osbuild-image/internal/osbuild-composer/common"
)

// ResponseError holds the API response error details
type ResponseError struct {
	Code int    `json:"code,omitempty"`
	ID   string `json:"id"`
	Msg  string `json:"msg"`
}

type ComposeRequestV0 struct {
	BlueprintName string `json:"blueprint_name"`
	ComposeType   string `json:"compose_type"`
	Branch        string `json:"branch"`
}
type ComposeResponseV0 struct {
	BuildID uuid.UUID `json:"build_id"`
	Status  bool      `json:"status"`
}

// This is similar to weldr.ComposeEntry but different because internally the image types are capitalized
type ComposeEntryV0 struct {
	ID          uuid.UUID              `json:"id"`
	Blueprint   string                 `json:"blueprint"`
	Version     string                 `json:"version"`
	ComposeType string                 `json:"compose_type"`
	ImageSize   uint64                 `json:"image_size"` // This is user-provided image size, not actual file size
	QueueStatus common.ImageBuildState `json:"queue_status"`
	JobCreated  float64                `json:"job_created"`
	JobStarted  float64                `json:"job_started,omitempty"`
	JobFinished float64                `json:"job_finished,omitempty"`
	Uploads     []uploadResponse       `json:"uploads,omitempty"`
}

type ComposeStatusResponseV0 struct {
	UUIDs []ComposeEntryV0 `json:"uuids"`
}

type ComposeTypeV0 struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

type ComposeTypesResponseV0 struct {
	Types []ComposeTypeV0 `json:"types"`
}

type DeleteComposeStatusV0 struct {
	UUID   uuid.UUID `json:"uuid"`
	Status bool      `json:"status"`
}

type DeleteComposeResponseV0 struct {
	UUIDs  []DeleteComposeStatusV0 `json:"uuids"`
	Errors []ResponseError         `json:"errors"`
}
