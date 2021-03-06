/* Copyright © 2017 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: BSD-2-Clause

   Generated by: https://github.com/swagger-api/swagger-codegen.git */

package aaa

// Appliance registration access token
type RegistrationToken struct {

	// List results
	Roles []string `json:"roles"`

	// Access token
	Token string `json:"token"`
}
