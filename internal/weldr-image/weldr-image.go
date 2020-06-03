package weldr_image

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/google/uuid"

	"github.com/ondrejbudai/osbuild-image/internal/osbuild-composer/client"
	"github.com/ondrejbudai/osbuild-image/internal/osbuild-composer/common"
	"github.com/ondrejbudai/osbuild-image/internal/osbuild-composer/weldr"
)

type Request struct {
	ImageType    string
	ImageWriter  io.Writer
	ManifestPath string
	Blueprint      []byte
}

type APIError struct {
	Message string
	Cause   error
}

func (e *APIError) Error() string {
	return fmt.Sprintf("%s\n caused by: %v", e.Message, e.Cause)
}

type UnknownImageTypeError struct {
	unknownImageType string
	validImageTypes  []string
}

func (e *UnknownImageTypeError) Error() string {
	validImageTypes := strings.Join(e.validImageTypes, ", ")
	return fmt.Sprintf("unknown image type: %s\nvalid image types: %s", e.unknownImageType, validImageTypes)
}

type requestHandler struct {
	request *Request

	client *http.Client

	blueprintName string
	composeId     uuid.UUID
}

func (r *Request) Validate() error {
	c := newClient()

	types, response, err := client.GetComposesTypesV0(c)
	if err := translateError(response, err); err != nil {
		return &APIError{
			Message: "cannot retrieve compose types",
			Cause:   err,
		}
	}

	if !isImageTypeValid(r.ImageType, types) {
		var validImageTypes []string

		for _, imageType := range types {
			if imageType.Enabled {
				validImageTypes = append(validImageTypes, imageType.Name)
			}
		}

		return &UnknownImageTypeError{
			unknownImageType: r.ImageType,
			validImageTypes:  validImageTypes,
		}
	}

	return nil
}

func (r *Request) Process() error {
	rh := requestHandler{
		client:  newClient(),
		request: r,
	}

	err := rh.pushBlueprint()
	if err != nil {
		return err
	}
	defer func() {
		err := rh.deleteBlueprint()
		if err != nil {
			log.Printf("cannot delete the blueprint: %v\n", err)
		}
	}()

	err = rh.pushCompose()
	if err != nil {
		return err
	}
	defer func() {
		err := rh.deleteCompose()
		if err != nil {
			log.Printf("cannot delete the compose: %v\n", err)
		}
	}()

	err = rh.waitForFinishedCompose()
	if err != nil {
		return err
	}

	err = rh.writeComposeImage()
	if err != nil {
		return err
	}

	if rh.request.ManifestPath != "" {
		err := rh.writeManifest()
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *requestHandler) pushBlueprint() error {
	blueprintDetail, err := loadBlueprintDetailOrCreateEmpty(h.request.Blueprint)
	if err != nil {
		return err
	}

	h.blueprintName = blueprintDetail.name

	if blueprintDetail.isTOML {
		response, err := client.PostTOMLBlueprintV0(h.client, blueprintDetail.blueprint)
		if err := translateError(response, err); err != nil {
			return &APIError{
				Message: "cannot post a new toml blueprint",
				Cause:   err,
			}
		}
	} else {
		response, err := client.PostJSONBlueprintV0(h.client, blueprintDetail.blueprint)
		if err := translateError(response, err); err != nil {
			return &APIError{
				Message: "cannot post a new json blueprint",
				Cause:   err,
			}
		}
	}

	return nil
}

func (h *requestHandler) pushCompose() error {
	response, err := client.PostComposeV0(h.client, `{"blueprint_name": "`+h.blueprintName+`", "compose_type": "`+h.request.ImageType+`"}`)
	if err := translateError(response, err); err != nil {
		return &APIError{
			Message: "cannot post a new compose",
			Cause:   err,
		}
	}

	composes, response, err := client.GetComposeStatusV0(h.client, "*", h.blueprintName, "", "")
	if err := translateError(response, err); err != nil {
		return &APIError{
			Message: "cannot retrieve a compose status",
			Cause:   err,
		}
	}

	if len(composes) != 1 {
		return fmt.Errorf("expected one compose, got %d", len(composes))
	}

	h.composeId = composes[0].ID

	return nil
}

func (h *requestHandler) waitForFinishedCompose() error {
	for {
		composes, response, err := client.GetComposeStatusV0(h.client, h.composeId.String(), "", "", "")
		if err := translateError(response, err); err != nil {
			return &APIError{
				Message: "cannot retrieve a compose status",
				Cause:   err,
			}
		}

		if len(composes) != 1 {
			panic("wrong number of composes")
		}

		if composes[0].QueueStatus == common.IBFailed {
			return errors.New("compose failed")
		}

		if composes[0].QueueStatus == common.IBFinished {
			break
		}

		time.Sleep(1 * time.Second)
	}

	return nil
}

func (h *requestHandler) writeComposeImage() error {
	response, err := client.WriteComposeImageV0(h.client, h.request.ImageWriter, h.composeId.String())
	if err := translateError(response, err); err != nil {
		return &APIError{
			Message: "cannot download the image",
			Cause:   err,
		}
	}

	return nil
}

func (h *requestHandler) deleteBlueprint() error {
	response, err := client.DeleteBlueprintV0(h.client, h.blueprintName)
	if err := translateError(response, err); err != nil {
		return &APIError{
			Message: "cannot delete the blueprint",
			Cause:   err,
		}
	}

	return nil
}

func (h *requestHandler) deleteCompose() error {
	_, response, err := client.DeleteComposeV0(h.client, h.composeId.String())
	if err := translateError(response, err); err != nil {
		return &APIError{
			Message: "cannot delete the blueprint",
			Cause:   err,
		}
	}

	return nil
}

func (h *requestHandler) writeManifest() error {
	var tarManifestBuffer bytes.Buffer
	response, err := client.WriteComposeMetadataV0(h.client, &tarManifestBuffer, h.composeId.String())

	if err := translateError(response, err); err != nil {
		return &APIError{
			Message: "cannot retrieve the manifest",
			Cause:   err,
		}
	}

	tarReader := tar.NewReader(&tarManifestBuffer)

	manifestHeader, err := tarReader.Next()
	if err != nil {
		return fmt.Errorf("cannot decode the metadata tar: %v", err)
	}

	f, err := os.Create(h.request.ManifestPath)
	if err != nil {
		return fmt.Errorf("cannot created the manifest file: %v", err)
	}

	_, err = io.CopyN(f, tarReader, manifestHeader.Size)
	if err != nil {
		return fmt.Errorf("cannot copy the manifest: %v", err)
	}

	return nil
}

func isImageTypeValid(imageTypeToValidate string, types []weldr.ComposeTypeV0) bool {
	for _, imageType := range types {
		if imageType.Name == imageTypeToValidate {
			return imageType.Enabled
		}
	}

	return false
}

func newClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", "/run/weldr/api.socket")
			},
		},
	}
}

func translateError(response *client.APIResponse, err error) error {
	if err != nil {
		return err
	}

	if response == nil {
		return nil
	}

	if !response.Status {
		var errorString string
		for _, e := range response.Errors {
			errorString = errorString + e.String() + "\n"
		}

		return errors.New("API returned an error: " + errorString)
	}

	return nil
}

type blueprintDetail struct {
	isTOML    bool
	name      string
	blueprint string
}

func loadBlueprintDetailOrCreateEmpty(blueprint []byte) (*blueprintDetail, error) {
	if len(blueprint) == 0 {
		name := uuid.New().String()
		return &blueprintDetail{
			isTOML:    false,
			name:      name,
			blueprint: fmt.Sprintf(`{"name":"%s"}`, name),
		}, nil
	}

	return loadBlueprintDetail(blueprint)
}

func loadBlueprintDetail(rawBlueprint []byte) (*blueprintDetail, error) {
	type blueprintStruct struct {
		Name string `json:"name" toml:"name"`
	}
	var blueprint blueprintStruct
	isTOML := false
	err := json.Unmarshal(rawBlueprint, &blueprint)

	if err != nil {
		err := toml.Unmarshal(rawBlueprint, &blueprint)
		isTOML = true
		if err != nil {
			return nil, fmt.Errorf("cannot unmarshal the blueprint, it's not json nor toml")
		}
	}

	return &blueprintDetail{
		name:      blueprint.Name,
		isTOML:    isTOML,
		blueprint: string(rawBlueprint),
	}, nil
}
