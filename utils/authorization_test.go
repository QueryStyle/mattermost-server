package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"
	"strings"
	"testing"

	"github.com/mattermost/mattermost-server/model"
)

type RoleState struct {
	RoleName   string `json:"roleName"`
	Permission string `json:"permission"`
	ShouldHave bool   `json:"shouldHave"`
}

func mockConfig() *model.Config {
	config := model.Config{}
	config.SetDefaults()
	return &config
}

func mapping() map[string]map[string][]RoleState {

	policiesRolesMapping := make(map[string]map[string][]RoleState)

	raw, err := ioutil.ReadFile("./policies-roles-mapping.json")
	if err != nil {
		panic(err)
	}

	var f map[string]interface{}
	err = json.Unmarshal(raw, &f)
	if err != nil {
		panic(err)
	}

	for policyName, value := range f {

		capitalizedName := fmt.Sprintf("%v%v", strings.ToUpper(policyName[:1]), policyName[1:])
		policiesRolesMapping[capitalizedName] = make(map[string][]RoleState)

		for policyValue, roleStatesMappings := range value.(map[string]interface{}) {

			var roleStates []RoleState
			for _, roleStateMapping := range roleStatesMappings.([]interface{}) {

				// Marshalling & Unmarshaling again... is this the best way?
				roleStateMappingJSON, _ := json.Marshal(roleStateMapping)
				var roleState RoleState
				_ = json.Unmarshal(roleStateMappingJSON, &roleState)

				roleStates = append(roleStates, roleState)

			}

			policiesRolesMapping[capitalizedName][policyValue] = roleStates

		}

	}

	return policiesRolesMapping
}

func TestSetRolePermissionsFromConfig(t *testing.T) {

	mapping := mapping()

	for policyName, v := range mapping {
		for policyValue, rolesMappings := range v {

			config := mockConfig()
			updateConfig(config, policyName, policyValue)
			roles := model.MakeDefaultRoles()
			updatedRoles := SetRolePermissionsFromConfig(roles, config, true)

			for _, roleMappingItem := range rolesMappings {
				role := updatedRoles[roleMappingItem.RoleName]

				permission := roleMappingItem.Permission
				hasPermission := roleHasPermission(role, permission)

				if roleMappingItem.ShouldHave && !hasPermission {
					t.Errorf("Expected '%v' to have '%v' permission when '%v' is set to '%v'.", role.Name, permission, policyName, policyValue)
				}
				if !roleMappingItem.ShouldHave && hasPermission {
					t.Errorf("Expected '%v' not to have '%v' permission when '%v' is set to '%v'.", role.Name, permission, policyName, policyValue)
				}

			}

		}
	}
}

func updateConfig(config *model.Config, key string, value string) {
	v := reflect.ValueOf(config.ServiceSettings)
	field := v.FieldByName(key)
	if !field.IsValid() {
		v = reflect.ValueOf(config.TeamSettings)
		field = v.FieldByName(key)
	}
	field.Elem().SetString(value)
}

func roleHasPermission(role *model.Role, permission string) bool {
	for _, p := range role.Permissions {
		if p == permission {
			return true
		}
	}
	return false
}
