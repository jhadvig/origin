package config

import (
	"encoding/json"
	"fmt"

	"github.com/GoogleCloudPlatform/kubernetes/pkg/api/errors"
	kubeclient "github.com/GoogleCloudPlatform/kubernetes/pkg/client"
	kubeapi "github.com/GoogleCloudPlatform/kubernetes/pkg/api"

	clientapi "github.com/openshift/origin/pkg/cmd/client/api"
)

type ApplyResult struct {
	Errors     errors.ErrorList
	ItemStatus string
}

// Apply creates and manages resources defined in the Config. The create process wont
// stop on error, but it will finish the job and then return error and for each item
// in the config a list of errors and status string.
func Apply(data []byte, storage clientapi.ClientMappings) (result []ApplyResult, err error) {
	// Unmarshal the Config JSON manually instead of using runtime.Decode()
	conf := struct {
		Items []json.RawMessage `json:"items" yaml:"items"`
	}{}

	if err := json.Unmarshal(data, &conf); err != nil {
		return nil, fmt.Errorf("Unable to parse Config: %v", err)
	}

	if len(conf.Items) == 0 {
		return nil, fmt.Errorf("Config.items is empty")
	}

	for i, item := range conf.Items {
		itemResult := ApplyResult{}

		if item == nil || (len(item) == 4 && string(item) == "null") {
			itemResult.Errors = append(itemResult.Errors, fmt.Errorf("Config.items[%v] is null", i))
			continue
		}

		itemBase := kubeapi.JSONBase{}

		err = json.Unmarshal(item, &itemBase)
		if err != nil {
			itemResult.Errors = append(itemResult.Errors, fmt.Errorf("Unable to parse Config item: %v", err))
			continue
		}

		if itemBase.Kind == "" {
			itemResult.Errors = append(itemResult.Errors, fmt.Errorf("Config.items[%v] has an empty 'kind'", i))
			continue
		}

		if itemBase.ID == "" {
			itemResult.Errors = append(itemResult.Errors, fmt.Errorf("Config.items[%v] has an empty 'id'", i))
			continue
		}

		client, path, err := getClientAndPath(itemBase.Kind, storage)
		if err != nil {
			itemResult.Errors = append(itemResult.Errors, fmt.Errorf("Config.items[%v]: %v", i, err))
			continue
		}
		if client == nil {
			itemResult.Errors = append(itemResult.Errors, fmt.Errorf("Config.items[%v]: Invalid client for 'kind=%v'", i, itemBase.Kind))
			continue
		}

		jsonResource, err := item.MarshalJSON()
		if err != nil {
			itemResult.Errors = append(itemResult.Errors, err)
			continue
		}
		request := client.Verb("POST").Path(path).Body(jsonResource)
		_, err = request.Do().Raw()

		if err != nil {
			itemResult.Errors = append(itemResult.Errors, fmt.Errorf("Failed to create Config.items[%v] of 'kind=%v': %v", i, itemBase.Kind, err))
			if statusErr, ok := err.(*kubeclient.StatusErr); ok {
				itemMessage := statusErr.Status.Message
				itemResult.ItemStatus = fmt.Sprintf("Creation failed for %v with id=%v' : %v", itemBase.Kind, itemBase.ID, itemMessage)
			}
		} else {
			itemResult.ItemStatus = fmt.Sprintf("Creation succeeded for %v with id=%v", itemBase.Kind, itemBase.ID)
		}
		result = append(result, itemResult)
	}
	return
}

// getClientAndPath returns the RESTClient and path defined for a given
// resource kind. Returns an error when no RESTClient is found.
func getClientAndPath(kind string, mappings clientapi.ClientMappings) (clientapi.RESTClient, string, error) {
	for k, m := range mappings {
		if m.Kind == kind {
			return m.Client, k, nil
		}
	}
	return nil, "", fmt.Errorf("No client found for 'kind=%v'", kind)
}

// reportError provides a human-readable error message that include the Config
// item JSON representation.
func reportError(item interface{}, message string) error {
	itemJSON, _ := json.Marshal(item)
	return fmt.Errorf(message+": %s", string(itemJSON))
}
