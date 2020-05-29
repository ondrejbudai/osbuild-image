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
		return errors.New("the image type is not valid")
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

	err = rh.pushCompose()
	if err != nil {
		return err
	}

	err = rh.waitForFinishedCompose()
	if err != nil {
		return err
	}

	err = rh.writeComposeImage()
	if err != nil {

	}
	return nil
}

func (h *requestHandler) pushBlueprint() error {
	blueprintName := uuid.New().String()
	response, err := client.PostJSONBlueprintV0(h.client, `{"name": "`+blueprintName+`"}`)

	if err := translateError(response, err); err != nil {
		return &APIError{
			Message: "cannot post a new blueprint",
			Cause:   err,
		}
	}

	h.blueprintName = blueprintName

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
