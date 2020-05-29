package weldr_image

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/ondrejbudai/osbuild-image/internal/osbuild-composer/client"
	"github.com/ondrejbudai/osbuild-image/internal/osbuild-composer/common"
	"github.com/ondrejbudai/osbuild-image/internal/osbuild-composer/weldr"
)

type Request struct {
	ImageType   string
	ImageWriter io.Writer
}

type APIError struct {
	Message string
	Cause   error
}

func (e *APIError) Error() string {
	return fmt.Sprintf("%s\n caused by: %v", e.Message, e.Cause)
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
		return errors.New("the image type is not valid")
	}

	return nil
}

func (r *Request) Process() error {
	c := newClient()

	id := uuid.New()
	response, err := client.PostJSONBlueprintV0(c, `{"name": "`+blueprintName(id)+`"}`)

	if err := translateError(response, err); err != nil {
		return &APIError{
			Message: "cannot post a new blueprint",
			Cause:   err,
		}
	}

	response, err = client.PostComposeV0(c, `{"blueprint_name": "`+blueprintName(id)+`", "compose_type": "`+r.ImageType+`"}`)
	if err := translateError(response, err); err != nil {
		return &APIError{
			Message: "cannot post a new compose",
			Cause:   err,
		}
	}

	var composeId uuid.UUID
	for {
		composes, response, err := client.GetComposeStatusV0(c, "*", blueprintName(id), "", "")
		if err := translateError(response, err); err != nil {
			return &APIError{
				Message: "cannot retrieve a compose status",
				Cause:   err,
			}
		}
		if len(composes) != 1 {
			return errors.New("wrong number of composes")
		}

		if composes[0].QueueStatus == common.IBFailed {
			return errors.New("compose failed")
		}

		if composes[0].QueueStatus == common.IBFinished {
			composeId = composes[0].ID
			break
		}

		time.Sleep(1 * time.Second)
	}

	response, err = client.WriteComposeImageV0(c, r.ImageWriter, composeId.String())
	if err := translateError(response, err); err != nil {
		return &APIError{
			Message: "canoot download the image",
			Cause:   err,
		}
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

func blueprintName(id uuid.UUID) string {
	return "bp-" + id.String()
}

func translateError(response *client.APIResponse, err error) error {
	if err != nil {
		return err
	}

	if response == nil {
		return nil
	}

	if !response.Status {
		return errors.New("API returned response with false status")
	}

	return nil
}
