/* Copyright © 2019 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/bindings"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/data"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
	"testing"
)

func TestAccDataSourceNsxtPolicyLBAppProfile_basic(t *testing.T) {
	name := "terraform_test"
	testResourceName := "data.nsxt_policy_lb_app_profile.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccSkipIfIsGlobalManager(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccDataSourceNsxtPolicyLBAppProfileDeleteByName(name)
		},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					if err := testAccDataSourceNsxtPolicyLBAppProfileCreate(name); err != nil {
						panic(err)
					}
				},
				Config: testAccNsxtPolicyLBAppProfileReadTemplate(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", name),
					resource.TestCheckResourceAttr(testResourceName, "type", "TCP"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
				),
			},
			{
				/* fetch test profile by name only */
				Config: testAccNsxtPolicyLBAppProfileNameOnlyTemplate(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", name),
					resource.TestCheckResourceAttr(testResourceName, "type", "TCP"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
				),
			},
			{
				/* fetch default HTTP profile */
				Config: testAccNsxtPolicyLBAppProfileTypeOnlyTemplate("HTTP"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testResourceName, "display_name"),
					resource.TestCheckResourceAttr(testResourceName, "type", "HTTP"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
				),
			},
			{
				Config: testAccNsxtPolicyEmptyTemplate(),
			},
		},
	})
}

func testAccDataSourceNsxtPolicyLBAppProfileCreate(name string) error {
	connector, err := testAccGetPolicyConnector()
	if err != nil {
		return fmt.Errorf("Error during test client initialization: %v", err)
	}
	client := infra.NewDefaultLbAppProfilesClient(connector)
	converter := bindings.NewTypeConverter()
	converter.SetMode(bindings.REST)

	displayName := name
	description := name
	profileType := "LBFastTcpProfile"
	obj := model.LBFastTcpProfile{
		Description:  &description,
		DisplayName:  &displayName,
		ResourceType: profileType,
	}

	dataValue, errs := converter.ConvertToVapi(obj, model.LBFastTcpProfileBindingType())
	if errs != nil {
		return fmt.Errorf("Error during conversion of LBFastTcpProfile: %v", errs[0])
	}

	// Generate a random ID for the resource
	id := newUUID()

	err = client.Patch(id, dataValue.(*data.StructValue))
	if err != nil {
		return handleCreateError("LBAppProfile", id, err)
	}
	return nil
}

func testAccDataSourceNsxtPolicyLBAppProfileDeleteByName(name string) error {
	connector, err := testAccGetPolicyConnector()
	if err != nil {
		return fmt.Errorf("Error during test client initialization: %v", err)
	}
	client := infra.NewDefaultLbAppProfilesClient(connector)

	// Find the object by name
	objList, err := client.List(nil, nil, nil, nil, nil, nil)
	if err != nil {
		return handleListError("LBAppProfile", err)
	}
	force := true
	for _, objInList := range objList.Results {
		result, err := policyLbAppProfileConvert(objInList, "ANY")
		if err != nil {
			return fmt.Errorf("Error during LBAppProfile conversion: %v", err)
		}
		if result != nil && *result.DisplayName == name {
			err := client.Delete(*result.Id, &force)
			if err != nil {
				return handleDeleteError("LBAppProfile", *result.Id, err)
			}
			return nil
		}
	}
	return fmt.Errorf("Error while deleting LBAppProfile '%s': resource not found", name)
}

func testAccNsxtPolicyLBAppProfileReadTemplate(name string) string {
	return fmt.Sprintf(`
data "nsxt_policy_lb_app_profile" "test" {
  type = "TCP"
  display_name = "%s"
}`, name)
}

func testAccNsxtPolicyLBAppProfileTypeOnlyTemplate(pType string) string {
	return fmt.Sprintf(`
data "nsxt_policy_lb_app_profile" "test" {
  type = "%s"
}`, pType)
}

func testAccNsxtPolicyLBAppProfileNameOnlyTemplate(name string) string {
	return fmt.Sprintf(`
data "nsxt_policy_lb_app_profile" "test" {
  display_name = "%s"
}`, name)
}
