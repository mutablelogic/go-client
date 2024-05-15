package homeassistant

import (
	"encoding/json"
	"strings"

	// Packages
	"github.com/mutablelogic/go-client"
	"golang.org/x/exp/maps"

	// Namespace imports
	. "github.com/djthorpe/go-errors"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type Domain struct {
	Domain   string             `json:"domain"`
	Services map[string]Service `json:"services,omitempty"`
}

type Service struct {
	Name        string           `json:"name"`
	Description string           `json:"description,omitempty,wrap"`
	Fields      map[string]Field `json:"fields,omitempty,wrap"`
}

type Field struct {
	Required bool                `json:"required,omitempty"`
	Example  any                 `json:"example,omitempty"`
	Selector map[string]Selector `json:"selector,omitempty"`
}

type Selector struct {
	Text              string `json:"text,omitempty"`
	Mode              string `json:"mode,omitempty"`
	Min               int    `json:"min,omitempty"`
	Max               int    `json:"max,omitempty"`
	UnitOfMeasurement string `json:"unit_of_measurement,omitempty"`
}

type reqCall struct {
	Entity string `json:"entity_id"`
}

///////////////////////////////////////////////////////////////////////////////
// API CALLS

// Domains returns all domains and their associated service objects
func (c *Client) Domains() ([]Domain, error) {
	var response []Domain
	if err := c.Do(nil, &response, client.OptPath("services")); err != nil {
		return nil, err
	}

	// Return success
	return response, nil
}

// Return callable services for a domain
func (c *Client) Services(domain string) ([]Service, error) {
	var response []Domain
	if err := c.Do(nil, &response, client.OptPath("services")); err != nil {
		return nil, err
	}
	for _, v := range response {
		if v.Domain != domain {
			continue
		}
		if len(v.Services) == 0 {
			// No services found
			return []Service{}, nil
		} else {
			return maps.Values(v.Services), nil
		}
	}
	// Return not found
	return nil, ErrNotFound.Withf("domain not found: %q", domain)
}

// Call a service for an entity. Returns a list of states that have
// changed while the service was being executed.
// TODO: This is a placeholder implementation, and requires fields to
// be passed in the request
func (c *Client) Call(service, entity string) ([]State, error) {
	domain := domainForEntity(entity)
	if domain == "" {
		return nil, ErrBadParameter.Withf("Invalid entity: %q", entity)
	}

	// Call the service
	var response []State
	if payload, err := client.NewJSONRequest(reqCall{
		Entity: entity,
	}); err != nil {
		return nil, err
	} else if err := c.Do(payload, &response, client.OptPath("services", domain, service)); err != nil {
		return nil, err
	}

	// Return success
	return response, nil
}

///////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (v Domain) String() string {
	data, _ := json.MarshalIndent(v, "", "  ")
	return string(data)
}

func (v Service) String() string {
	data, _ := json.MarshalIndent(v, "", "  ")
	return string(data)
}

func (v Field) String() string {
	data, _ := json.MarshalIndent(v, "", "  ")
	return string(data)
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

func domainForEntity(entity string) string {
	parts := strings.SplitN(entity, ".", 2)
	if len(parts) == 2 {
		return parts[0]
	} else {
		return ""
	}
}
