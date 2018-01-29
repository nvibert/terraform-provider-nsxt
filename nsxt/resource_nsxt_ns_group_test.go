/* Copyright © 2018 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/vmware/go-vmware-nsxt"
	"net/http"
	"testing"
)

func TestNSXNSGroupBasic(t *testing.T) {

	grpName := fmt.Sprintf("test-nsx-ns-group")
	updateGrpName := fmt.Sprintf("%s-update", grpName)
	testResourceName := "nsxt_ns_group.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXNSGroupCheckDestroy(state, grpName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXNSGroupCreateTemplate(grpName),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXNSGroupExists(grpName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", grpName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "tags.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "members.#", "0"),
				),
			},
			{
				Config: testAccNSXNSGroupUpdateTemplate(updateGrpName),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXNSGroupExists(updateGrpName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updateGrpName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test Update"),
					resource.TestCheckResourceAttr(testResourceName, "tags.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "members.#", "0"),
				),
			},
		},
	})
}

func TestNSXNSGroupNested(t *testing.T) {

	grpName := fmt.Sprintf("test-nsx-ns-group")
	updateGrpName := fmt.Sprintf("%s-update", grpName)
	testResourceName := "nsxt_ns_group.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXNSGroupCheckDestroy(state, grpName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXNSGroupNestedCreateTemplate(grpName),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXNSGroupExists(grpName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", grpName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "members.#", "1"),
				),
			},
			{
				Config: testAccNSXNSGroupNestedUpdateTemplate(updateGrpName),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXNSGroupExists(updateGrpName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updateGrpName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test Update"),
					resource.TestCheckResourceAttr(testResourceName, "members.#", "2"),
				),
			},
		},
	})
}

func testAccNSXNSGroupExists(display_name string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		nsxClient := testAccProvider.Meta().(*nsxt.APIClient)

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("NS Group resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("NS Group resource ID not set in resources ")
		}

        localVarOptionals := make(map[string]interface{})
		group, responseCode, err := nsxClient.GroupingObjectsApi.ReadNSGroup(nsxClient.Context, resourceID, localVarOptionals)
		if err != nil {
			return fmt.Errorf("Error while retrieving NS Group ID %s. Error: %v", resourceID, err)
		}

		if responseCode.StatusCode != http.StatusOK {
			return fmt.Errorf("Error while checking if NS Group %s exists. HTTP return code was %d", resourceID, responseCode)
		}

		if display_name == group.DisplayName {
			return nil
		}
		return fmt.Errorf("NS Group %s wasn't found", display_name)
	}
}

func testAccNSXNSGroupCheckDestroy(state *terraform.State, display_name string) error {

	nsxClient := testAccProvider.Meta().(*nsxt.APIClient)

	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_logical_port" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
        localVarOptionals := make(map[string]interface{})
		group, responseCode, err := nsxClient.GroupingObjectsApi.ReadNSGroup(nsxClient.Context, resourceID, localVarOptionals)
		if err != nil {
			if responseCode.StatusCode != http.StatusOK {
				return nil
			}
			return fmt.Errorf("Error while retrieving NS Group ID %s. Error: %v", resourceID, err)
		}

		if display_name == group.DisplayName {
			return fmt.Errorf("NS Group %s still exists", display_name)
		}
	}
	return nil
}

func testAccNSXNSGroupCreateTemplate(name string) string {
	return fmt.Sprintf(`
resource "nsxt_ns_group" "test" {
	display_name = "%s"
	description = "Acceptance Test"
	tags = [{scope = "scope1"
	    	 tag = "tag1"}
	]
}`, name)
}

func testAccNSXNSGroupUpdateTemplate(updatedName string) string {
	return fmt.Sprintf(`
resource "nsxt_ns_group" "test" {
	display_name = "%s"
	description = "Acceptance Test Update"
	tags = [{scope = "scope1"
             tag = "tag1"}, 
            {scope = "scope2"
    	     tag = "tag2"}
	]
}`, updatedName)
}

func testAccNSXNSGroupNestedCreateTemplate(name string) string {
	return fmt.Sprintf(`
resource "nsxt_ns_group" "GRP1" {
	display_name = "grp1"
}
resource "nsxt_ns_group" "test" {
	display_name = "%s"
	description = "Acceptance Test"
	members = [{target_type = "NSGroup"
	            value = "${nsxt_ns_group.GRP1.id}"}]
}`, name)
}

func testAccNSXNSGroupNestedUpdateTemplate(updatedName string) string {
	return fmt.Sprintf(`
resource "nsxt_ns_group" "GRP1" {
	display_name = "grp1"
}
resource "nsxt_ns_group" "GRP2" {
	display_name = "grp2"
}
resource "nsxt_ns_group" "test" {
	display_name = "%s"
	description = "Acceptance Test Update"
	members = [{target_type = "NSGroup"
	            value = "${nsxt_ns_group.GRP1.id}"},
	           {target_type = "NSGroup"
	            value = "${nsxt_ns_group.GRP2.id}"}]
}`, updatedName)
}