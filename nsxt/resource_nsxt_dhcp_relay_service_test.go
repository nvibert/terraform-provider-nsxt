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

func TestNSXDhcpRelayServiceBasic(t *testing.T) {

	prfName := fmt.Sprintf("test-nsx-dhcp-relay-service")
	updatePrfName := fmt.Sprintf("%s-update", prfName)
	testResourceName := "nsxt_dhcp_relay_service.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXDhcpRelayServiceCheckDestroy(state, prfName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXDhcpRelayServiceCreateTemplate(prfName),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXDhcpRelayServiceExists(prfName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", prfName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "tags.#", "1"),
				),
			},
			{
				Config: testAccNSXDhcpRelayServiceUpdateTemplate(updatePrfName),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXDhcpRelayServiceExists(updatePrfName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updatePrfName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test Update"),
					resource.TestCheckResourceAttr(testResourceName, "tags.#", "2"),
				),
			},
		},
	})
}

func testAccNSXDhcpRelayServiceExists(display_name string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		nsxClient := testAccProvider.Meta().(*nsxt.APIClient)

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Dhcp Relay Service resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Dhcp Relay Service resource ID not set in resources ")
		}

		service, responseCode, err := nsxClient.LogicalRoutingAndServicesApi.ReadDhcpRelay(nsxClient.Context, resourceID)
		if err != nil {
			return fmt.Errorf("Error while retrieving Dhcp Relay Service ID %s. Error: %v", resourceID, err)
		}

		if responseCode.StatusCode != http.StatusOK {
			return fmt.Errorf("Error while checking if Dhcp Relay Service %s exists. HTTP return code was %d", resourceID, responseCode)
		}

		if display_name == service.DisplayName {
			return nil
		}
		return fmt.Errorf("Dhcp Relay Service %s wasn't found", display_name)
	}
}

func testAccNSXDhcpRelayServiceCheckDestroy(state *terraform.State, display_name string) error {

	nsxClient := testAccProvider.Meta().(*nsxt.APIClient)

	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_logical_port" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		service, responseCode, err := nsxClient.LogicalRoutingAndServicesApi.ReadDhcpRelay(nsxClient.Context, resourceID)
		if err != nil {
			if responseCode.StatusCode != http.StatusOK {
				return nil
			}
			return fmt.Errorf("Error while retrieving Dhcp Relay Service ID %s. Error: %v", resourceID, err)
		}

		if display_name == service.DisplayName {
			return fmt.Errorf("Dhcp Relay Service %s still exists", display_name)
		}
	}
	return nil
}

func testAccNSXDhcpRelayServiceCreateTemplate(name string) string {
	return fmt.Sprintf(`
resource "nsxt_dhcp_relay_profile" "test" {
	display_name = "prf"
	server_addresses = ["1.1.1.1"]
}

resource "nsxt_dhcp_relay_service" "test" {
	display_name = "%s"
	description = "Acceptance Test"
	dhcp_relay_profile_id = "${nsxt_dhcp_relay_profile.test.id}"
	tags = [{scope = "scope1"
	    	 tag = "tag1"}
	]
}`, name)
}

func testAccNSXDhcpRelayServiceUpdateTemplate(updated_name string) string {
	return fmt.Sprintf(`
resource "nsxt_dhcp_relay_profile" "test" {
	display_name = "prf"
	server_addresses = ["1.1.1.1"]
}

resource "nsxt_dhcp_relay_service" "test" {
	display_name = "%s"
	description = "Acceptance Test Update"
	dhcp_relay_profile_id = "${nsxt_dhcp_relay_profile.test.id}"
	tags = [{scope = "scope1"
	         tag = "tag1"}, 
	        {scope = "scope2"
	    	 tag = "tag2"}
	]
}`, updated_name)
}
