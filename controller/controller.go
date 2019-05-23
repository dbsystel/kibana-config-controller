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

type Controller struct {
	logger log.Logger
	k      kibana.APIClient
}

func (c *Controller) Create(obj interface{}) {
	configmapObj := obj.(*v1.ConfigMap)
	id := configmapObj.Annotations["kibana.net/id"]            // TODO add error check
	kobj := configmapObj.Annotations["kibana.net/savedobject"] // TODO add error check

	kibanaID, _ := strconv.Atoi(id)
	isKibanaObject, _ := strconv.ParseBool(kobj)

	if kibanaID == c.k.ID && isKibanaObject {
		var err error
		for k, v := range configmapObj.Data {
			objType := c.searchTypeFromJson(strings.NewReader(v))
			if objType == "" {
				level.Info(c.logger).Log("msg", "type not found in Json body. Can not be created.")
				continue
			}

			level.Info(c.logger).Log(
				"msg", "Creating "+objType+": "+k,
				"configmap", configmapObj.Name,
				"namespace", configmapObj.Namespace,
			)

			objID := c.searchIDFromJson(strings.NewReader(v))
			if objID == "" {
				level.Info(c.logger).Log("msg", "id not found in Json body. Can not be created.")
				continue
			}

			v = c.deleteNotAllowedFields(strings.NewReader(v))

			err = c.k.CreateObject(objType, objID, strings.NewReader(v))

			if err != nil {
				level.Info(c.logger).Log(
					"msg", "Failed to create: "+k,
					"configmap", configmapObj.Name,
					"namespace", configmapObj.Namespace,
				)
				level.Error(c.logger).Log("err", err.Error())
			} else {
				level.Info(c.logger).Log(
					"msg", "Succeeded: Created: "+k,
					"configmap", configmapObj.Name,
					"namespace", configmapObj.Namespace,
				)
			}
		}
	} else {
		level.Debug(c.logger).Log("msg", "Skipping configmap: "+configmapObj.Name)
	}
}

func (c *Controller) Update(oldobj interface{}, newobj interface{}) {
	configmapObj := newobj.(*v1.ConfigMap)
	id := configmapObj.Annotations["kibana.net/id"]            // TODO add error check
	kobj := configmapObj.Annotations["kibana.net/savedobject"] // TODO add error check

	kibanaID, _ := strconv.Atoi(id)
	isKibanaObject, _ := strconv.ParseBool(kobj)

	if noDifference(oldobj.(*v1.ConfigMap), configmapObj) {
		level.Debug(c.logger).Log("msg", "Skipping automatically updated configmap:"+configmapObj.Name)
		return
	}

	if kibanaID == c.k.ID && isKibanaObject {
		var err error
		for k, v := range configmapObj.Data {
			objType := c.searchTypeFromJson(strings.NewReader(v))
			if objType == "" {
				level.Info(c.logger).Log("msg", "type not found in Json body. Can not be updated.")
				continue
			}
			level.Info(c.logger).Log(
				"msg", "Updating "+objType+": "+k,
				"configmap", configmapObj.Name,
				"namespace", configmapObj.Namespace,
			)

			objID := c.searchIDFromJson(strings.NewReader(v))
			if objID == "" {
				level.Info(c.logger).Log("msg", "id not found in Json body. Can not be updated.")
				continue
			}

			v = c.deleteNotAllowedFields(strings.NewReader(v))

			err = c.k.UpdateObject(objType, objID, strings.NewReader(v))

			if err != nil {
				level.Info(c.logger).Log(
					"msg", "Failed to update: "+k,
					"configmap", configmapObj.Name,
					"namespace", configmapObj.Namespace,
				)
				level.Error(c.logger).Log("err", err.Error())
			} else {
				level.Info(c.logger).Log(
					"msg", "Succeeded: Updated: "+k,
					"configmap", configmapObj.Name,
					"namespace", configmapObj.Namespace,
				)
			}
		}
	} else {
		level.Debug(c.logger).Log("msg", "Skipping configmap: "+configmapObj.Name)
	}
}
func (c *Controller) Delete(obj interface{}) {
	configmapObj := obj.(*v1.ConfigMap)
	id := configmapObj.Annotations["kibana.net/id"]            // TODO add error check
	kobj := configmapObj.Annotations["kibana.net/savedobject"] // TODO add error check

	kibanaID, _ := strconv.Atoi(id)
	isKibanaObject, _ := strconv.ParseBool(kobj)

	if kibanaID == c.k.ID && isKibanaObject {
		var err error
		for k, v := range configmapObj.Data {
			objType := c.searchTypeFromJson(strings.NewReader(v))
			if objType == "" {
				level.Info(c.logger).Log("msg", "type not found in Json body. Can not be deleted.")
				continue
			}
			level.Info(c.logger).Log(
				"msg", "Deleting "+objType+": "+k,
				"configmap", configmapObj.Name,
				"namespace", configmapObj.Namespace,
			)

			objID := c.searchIDFromJson(strings.NewReader(v))
			if objID == "" {
				level.Info(c.logger).Log("msg", "id not found in Json body. Can not be deleted.")
				continue
			}

			err = c.k.DeleteObject(objType, objID)

			if err != nil {
				level.Info(c.logger).Log(
					"msg", "Failed to delete: "+k,
					"configmap", configmapObj.Name,
					"namespace", configmapObj.Namespace,
				)
				level.Error(c.logger).Log("err", err.Error())
			} else {
				level.Info(c.logger).Log(
					"msg", "Succeeded: Deleted: "+k,
					"configmap", configmapObj.Name,
					"namespace", configmapObj.Namespace,
				)
			}
		}
	} else {
		level.Debug(c.logger).Log("msg", "Skipping configmap: "+configmapObj.Name)
	}
}

func (c *Controller) searchIDFromJson(objJSON *strings.Reader) string {
	newObj := make(map[string]interface{})
	err := json.NewDecoder(objJSON).Decode(&newObj)
	if err != nil {
		level.Error(c.logger).Log("err", err.Error())
	}

	if newObj["id"] != nil {
		return newObj["id"].(string)
	} else if newObj["_id"] != nil {
		return newObj["_id"].(string)
	}
	return ""
}

func (c *Controller) searchTypeFromJson(objJSON *strings.Reader) string {
	newObj := make(map[string]interface{})
	err := json.NewDecoder(objJSON).Decode(&newObj)
	if err != nil {
		level.Error(c.logger).Log("err", err.Error())
	}

	if newObj["type"] != nil {
		return newObj["type"].(string)
	} else if newObj["_type"] != nil {
		return newObj["_type"].(string)
	}
	return ""
}

// delete the fields which are not allowed for kibana api in json body
func (c *Controller) deleteNotAllowedFields(objJSON *strings.Reader) string {
	newObj := make(map[string]interface{})
	err := json.NewDecoder(objJSON).Decode(&newObj)
	if err != nil {
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
		level.Error(c.logger).Log("err", err.Error())
	}
	return string(jsonString)
}

// create new Controller instance
func New(k kibana.APIClient, logger log.Logger) *Controller {
	controller := &Controller{}
	controller.logger = logger
	controller.k = k
	return controller
}

// are two configmaps same
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
