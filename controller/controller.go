package controller

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/dbsystel/kibana-config-controller/kibana"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	v1 "k8s.io/api/core/v1"
)

// Controller wrapper for kibana
type Controller struct {
	logger log.Logger
	k      kibana.IAPIClient
}

// Create creates the given configMap
func (c *Controller) Create(obj interface{}) {
	configmapObj := obj.(*v1.ConfigMap)

	id := configmapObj.Annotations["kibana.net/id"]
	kobj := configmapObj.Annotations["kibana.net/savedobject"]

	var err error
	kibanaID, err := strconv.Atoi(id)
	if err != nil {
		//nolint:errcheck
		level.Info(c.logger).Log(
			"msg", "Kibana ID is not an int: "+id,
			"configmap", configmapObj.Name,
			"namespace", configmapObj.Namespace,
		)
		return
	}

	isKibanaObject, err := strconv.ParseBool(kobj)
	if err != nil {
		//nolint:errcheck
		level.Info(c.logger).Log(
			"msg", "Kibana savedObject is not a bool: "+kobj,
			"configmap", configmapObj.Name,
			"namespace", configmapObj.Namespace,
		)
		return
	}

	if kibanaID == c.k.GetID() && isKibanaObject {
		for k, v := range configmapObj.Data {
			objType := c.searchTypeFromJSON(strings.NewReader(v))
			if objType == "" {
				//nolint:errcheck
				level.Info(c.logger).Log("msg", "type not found in JSON body. Can not be created.")
				continue
			}

			//nolint:errcheck
			level.Info(c.logger).Log(
				"msg", "Creating "+objType+": "+k,
				"configmap", configmapObj.Name,
				"namespace", configmapObj.Namespace,
			)

			objID := c.searchIDFromJSON(strings.NewReader(v))
			if objID == "" {
				//nolint:errcheck
				level.Info(c.logger).Log("msg", "id not found in JSON body. Can not be created.")
				continue
			}

			v = c.deleteNotAllowedFields(strings.NewReader(v))

			err = c.k.CreateObject(objType, objID, strings.NewReader(v))

			if err != nil {
				//nolint:errcheck
				level.Info(c.logger).Log(
					"msg", "Failed to create: "+k,
					"configmap", configmapObj.Name,
					"namespace", configmapObj.Namespace,
				)
				//nolint:errcheck
				level.Error(c.logger).Log("err", err.Error())
				return
			}

			//nolint:errcheck
			level.Info(c.logger).Log(
				"msg", "Succeeded: Created: "+k,
				"configmap", configmapObj.Name,
				"namespace", configmapObj.Namespace,
			)
		}
	}

	//nolint:errcheck
	level.Debug(c.logger).Log("msg", "Skipping configmap: "+configmapObj.Name)
}

// Update updates the given configMap
func (c *Controller) Update(oldobj interface{}, newobj interface{}) {
	configmapObj := newobj.(*v1.ConfigMap)
	id := configmapObj.Annotations["kibana.net/id"]            // TODO add error check
	kobj := configmapObj.Annotations["kibana.net/savedobject"] // TODO add error check

	kibanaID, _ := strconv.Atoi(id)
	isKibanaObject, _ := strconv.ParseBool(kobj)

	if noDifference(oldobj.(*v1.ConfigMap), configmapObj) {
		//nolint:errcheck
		level.Debug(c.logger).Log("msg", "Skipping automatically updated configmap:"+configmapObj.Name)
		return
	}

	if kibanaID == c.k.GetID() && isKibanaObject {
		var err error
		for k, v := range configmapObj.Data {
			objType := c.searchTypeFromJSON(strings.NewReader(v))
			if objType == "" {
				//nolint:errcheck
				level.Info(c.logger).Log("msg", "type not found in JSON body. Can not be updated.")
				continue
			}
			//nolint:errcheck
			level.Info(c.logger).Log(
				"msg", "Updating "+objType+": "+k,
				"configmap", configmapObj.Name,
				"namespace", configmapObj.Namespace,
			)

			objID := c.searchIDFromJSON(strings.NewReader(v))
			if objID == "" {
				//nolint:errcheck
				level.Info(c.logger).Log("msg", "id not found in JSON body. Can not be updated.")
				continue
			}

			v = c.deleteNotAllowedFields(strings.NewReader(v))

			err = c.k.UpdateObject(objType, objID, strings.NewReader(v))

			if err != nil {
				//nolint:errcheck
				level.Info(c.logger).Log(
					"msg", "Failed to update: "+k,
					"configmap", configmapObj.Name,
					"namespace", configmapObj.Namespace,
				)
				//nolint:errcheck
				level.Error(c.logger).Log("err", err.Error())
			} else {
				//nolint:errcheck
				level.Info(c.logger).Log(
					"msg", "Succeeded: Updated: "+k,
					"configmap", configmapObj.Name,
					"namespace", configmapObj.Namespace,
				)
			}
		}
	} else {
		//nolint:errcheck
		level.Debug(c.logger).Log("msg", "Skipping configmap: "+configmapObj.Name)
	}
}

// Delete deletes the given configMap
func (c *Controller) Delete(obj interface{}) {
	configmapObj := obj.(*v1.ConfigMap)
	id := configmapObj.Annotations["kibana.net/id"]            // TODO add error check
	kobj := configmapObj.Annotations["kibana.net/savedobject"] // TODO add error check

	kibanaID, _ := strconv.Atoi(id)
	isKibanaObject, _ := strconv.ParseBool(kobj)

	if kibanaID == c.k.GetID() && isKibanaObject {
		var err error
		for k, v := range configmapObj.Data {
			objType := c.searchTypeFromJSON(strings.NewReader(v))
			if objType == "" {
				//nolint:errcheck
				level.Info(c.logger).Log("msg", "type not found in JSON body. Can not be deleted.")
				continue
			}
			//nolint:errcheck
			level.Info(c.logger).Log(
				"msg", "Deleting "+objType+": "+k,
				"configmap", configmapObj.Name,
				"namespace", configmapObj.Namespace,
			)

			objID := c.searchIDFromJSON(strings.NewReader(v))
			if objID == "" {
				//nolint:errcheck
				level.Info(c.logger).Log("msg", "id not found in JSON body. Can not be deleted.")
				continue
			}

			err = c.k.DeleteObject(objType, objID)

			if err != nil {
				//nolint:errcheck
				level.Info(c.logger).Log(
					"msg", "Failed to delete: "+k,
					"configmap", configmapObj.Name,
					"namespace", configmapObj.Namespace,
				)
				//nolint:errcheck
				level.Error(c.logger).Log("err", err.Error())
			} else {
				//nolint:errcheck
				level.Info(c.logger).Log(
					"msg", "Succeeded: Deleted: "+k,
					"configmap", configmapObj.Name,
					"namespace", configmapObj.Namespace,
				)
			}
		}
	} else {
		//nolint:errcheck
		level.Debug(c.logger).Log("msg", "Skipping configmap: "+configmapObj.Name)
	}
}

// searchTypeFromJSON extracts the 'id' propery from the given object
func (c *Controller) searchIDFromJSON(objJSON *strings.Reader) string {
	newObj := make(map[string]interface{})
	err := json.NewDecoder(objJSON).Decode(&newObj)
	if err != nil {
		//nolint:errcheck
		level.Error(c.logger).Log("err", err.Error())
	}

	if newObj["id"] != nil {
		return newObj["id"].(string)
	} else if newObj["_id"] != nil {
		return newObj["_id"].(string)
	}
	return ""
}

// searchTypeFromJSON extracts the 'type' propery from the given object
func (c *Controller) searchTypeFromJSON(objJSON *strings.Reader) string {
	newObj := make(map[string]interface{})
	err := json.NewDecoder(objJSON).Decode(&newObj)
	if err != nil {
		//nolint:errcheck
		level.Error(c.logger).Log("err", err.Error())
	}

	if newObj["type"] != nil {
		return newObj["type"].(string)
	} else if newObj["_type"] != nil {
		return newObj["_type"].(string)
	}
	return ""
}

// deleteNotAllowedFields deletes the fields which are not allowed for kibana api in json body
func (c *Controller) deleteNotAllowedFields(objJSON *strings.Reader) string {
	newObj := make(map[string]interface{})
	err := json.NewDecoder(objJSON).Decode(&newObj)
	if err != nil {
		//nolint:errcheck
		level.Error(c.logger).Log("err", err.Error())
	}

	delete(newObj, "_id")
	delete(newObj, "id")
	delete(newObj, "_type")
	delete(newObj, "type")
	delete(newObj, "_meta")
	delete(newObj, "meta")
	if newObj["_source"] != nil {
		newObj["attributes"] = newObj["_source"]
		delete(newObj, "_source")
	} else if newObj["source"] != nil {
		newObj["attributes"] = newObj["source"]
		delete(newObj, "source")
	}

	jsonString, err := json.Marshal(newObj)
	if err != nil {
		//nolint:errcheck
		level.Error(c.logger).Log("err", err.Error())
	}
	return string(jsonString)
}

// New creates new Controller instance
func New(k kibana.IAPIClient, logger log.Logger) *Controller {
	controller := &Controller{}
	controller.logger = logger
	controller.k = k
	return controller
}

// noDifference checks if two configmaps are equal
func noDifference(newConfigMap *v1.ConfigMap, oldConfigMap *v1.ConfigMap) bool {
	if len(newConfigMap.Data) != len(oldConfigMap.Data) {
		return false
	}
	for k, v := range newConfigMap.Data {
		if v != oldConfigMap.Data[k] {
			return false
		}
	}
	if len(newConfigMap.Annotations) != len(oldConfigMap.Annotations) {
		return false
	}
	for k, v := range newConfigMap.Annotations {
		if v != oldConfigMap.Annotations[k] {
			return false
		}
	}
	return true
}
