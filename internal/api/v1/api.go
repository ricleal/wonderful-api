package v1

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"wonderful/internal/api/v1/openapi"
	"wonderful/internal/service"
)

// wonderfulAPI is the implementation of the API.
type wonderfulAPI struct {
	userService service.UserService
}

// New returns a new wonderfulAPI.
func New(userService service.UserService) *wonderfulAPI {
	return &wonderfulAPI{
		userService: userService,
	}
}

// validateQueryParams validates the query parameters.
// TODO: check why OpenAPI Validator is not working for min/max boundaries.
func validateQueryParams(params openapi.GetWonderfulsParams) error {
	if params.Limit != nil && (*params.Limit < 1 || *params.Limit > 100) {
		return fmt.Errorf("invalid limit: limit must be between 1 and 100")
	}
	if params.StartingAfter != nil && params.EndingBefore != nil {
		return fmt.Errorf("invalid startingAfter and endingBefore: only one of them can be used")
	}
	return nil
}

// GetWonderfuls returns a list of wonderfuls.
func (c *wonderfulAPI) GetWonderfuls(w http.ResponseWriter, r *http.Request, params openapi.GetWonderfulsParams) {
	ctx := r.Context()

	if err := validateQueryParams(params); err != nil {
		sendAPIError(ctx, w, http.StatusBadRequest, err.Error(), err)
		return
	}

	p, err := service.ConvertParams(params.Limit, params.StartingAfter, params.EndingBefore, params.Email)
	if err != nil {
		sendAPIError(ctx, w, http.StatusBadRequest, "Invalid parameters", err)
		return
	}

	users, err := c.userService.ListUsers(ctx, *p)
	if err != nil {
		sendAPIError(ctx, w, http.StatusInternalServerError, "Error listing users", err)
		return
	}

	openapiUsers := make([]openapi.User, 0, len(users))
	for _, user := range users {
		picLarge := user.Picture["large"]
		picMedium := user.Picture["medium"]
		picThumbnail := user.Picture["thumbnail"]
		cellPhone := user.Cell
		mainPhone := user.Phone
		openapiUsers = append(openapiUsers, openapi.User{
			Email: user.Email,
			Id:    user.ID,
			Name:  user.Name,
			Phone: &struct {
				Cell *string "json:\"cell,omitempty\""
				Main *string "json:\"main,omitempty\""
			}{
				Cell: &cellPhone,
				Main: &mainPhone,
			},
			Picture: &struct {
				Large     *string "json:\"large,omitempty\""
				Medium    *string "json:\"medium,omitempty\""
				Thumbnail *string "json:\"thumbnail,omitempty\""
			}{
				Large:     &picLarge,
				Medium:    &picMedium,
				Thumbnail: &picThumbnail,
			},
			RegistrationDate: user.Registration,
		})
	}
	json.NewEncoder(w).Encode(openapiUsers) //nolint:errcheck //ignore error
}

// PostPopulate populates the database with users.
func (c *wonderfulAPI) PostPopulate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	err := c.userService.Create(ctx)
	if err != nil {
		// let's check if the error is from the RandomUser API
		if errors.Is(err, service.ErrRandomUserAPI) {
			sendAPIError(ctx, w, http.StatusInternalServerError, "Error getting data from RandomUser API", err)
			return
		}
		sendAPIError(ctx, w, http.StatusInternalServerError, "Error populating users", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
