// Copyright 2019 Smart-Edge.com, Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	cce "github.com/smartedgemec/controller-ce"
	"github.com/smartedgemec/controller-ce/uuid"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("/dns_vm_app_aliases", func() {
	var (
		vmAppID string
	)

	BeforeEach(func() {
		vmAppID = postVMApps()
	})

	Describe("POST /dns_vm_app_aliases", func() {
		DescribeTable("201 Created",
			func(req string) {
				By("Sending a POST /dns_vm_app_aliases request")
				resp, err := http.Post(
					"http://127.0.0.1:8080/dns_vm_app_aliases",
					"application/json",
					strings.NewReader(fmt.Sprintf(req, vmAppID)))
				Expect(err).ToNot(HaveOccurred())

				By("Verifying a 201 response")
				Expect(resp.StatusCode).To(Equal(http.StatusCreated))

				By("Reading the response body")
				body, err := ioutil.ReadAll(resp.Body)
				Expect(err).ToNot(HaveOccurred())

				var respBody struct {
					ID string
				}

				By("Unmarshalling the response")
				Expect(json.Unmarshal(body, &respBody)).To(Succeed())

				By("Verifying a UUID was returned")
				Expect(uuid.IsValid(respBody.ID)).To(BeTrue())
			},
			Entry(
				"POST /dns_vm_app_aliases",
				`
				{
					"name": "dns vm app alias 123",
					"description": "description 1",
					"vm_app_id": "%s"
				}`),
		)

		DescribeTable("400 Bad Request",
			func(req, expectedResp string) {
				By("Sending a POST /dns_vm_app_aliases request")
				resp, err := http.Post(
					"http://127.0.0.1:8080/dns_vm_app_aliases",
					"application/json",
					strings.NewReader(req))
				Expect(err).ToNot(HaveOccurred())

				By("Verifying a 400 Bad Request response")
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))

				By("Reading the response body")
				body, err := ioutil.ReadAll(resp.Body)
				Expect(err).ToNot(HaveOccurred())

				By("Verifying the response body")
				Expect(string(body)).To(Equal(expectedResp))
			},
			Entry(
				"POST /dns_vm_app_aliases with id",
				`
				{
					"id": "123"
				}`,
				"Validation failed: id cannot be specified in POST request"),
			Entry(
				"POST /dns_vm_app_aliases without name",
				`
				{
					"description": "description 1",
					"vm_app_id": "123"
				}`,
				"Validation failed: name cannot be empty"),
			Entry(
				"POST /dns_vm_app_aliases without description",
				`
				{
					"name": "dns vm app alias 123",
					"vm_app_id": "123"
				}`,
				"Validation failed: description cannot be empty"),
			Entry(
				"POST /dns_vm_app_aliases without vm_app_id",
				`
				{
					"name": "dns vm app alias 123",
					"description": "description 1"
				}`,
				"Validation failed: vm_app_id not a valid uuid"),
		)
	})

	Describe("GET /dns_vm_app_aliases", func() {
		var (
			dnsVMAppAliasID  string
			dnsVMAppAlias2ID string
		)

		BeforeEach(func() {
			dnsVMAppAliasID = postDNSVMAppAliases(vmAppID)
			dnsVMAppAlias2ID = postDNSVMAppAliases(vmAppID)
		})

		DescribeTable("200 OK",
			func() {
				By("Sending a GET /dns_vm_app_aliases request")
				resp, err := http.Get(
					"http://127.0.0.1:8080/dns_vm_app_aliases")

				By("Verifying a 200 OK response")
				Expect(err).ToNot(HaveOccurred())
				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				By("Reading the response body")
				body, err := ioutil.ReadAll(resp.Body)
				Expect(err).ToNot(HaveOccurred())

				var dnsVMAppAliases []cce.DNSVMAppAlias

				By("Unmarshalling the response")
				Expect(json.Unmarshal(body, &dnsVMAppAliases)).
					To(Succeed())

				By("Verifying the 2 created DNS vm app aliases were returned") //nolint:lll
				Expect(dnsVMAppAliases).To(ContainElement(
					cce.DNSVMAppAlias{
						ID:          dnsVMAppAliasID,
						Name:        "dns vm app alias 123",
						Description: "description 1",
						VMAppID:     vmAppID,
					}))
				Expect(dnsVMAppAliases).To(ContainElement(
					cce.DNSVMAppAlias{
						ID:          dnsVMAppAlias2ID,
						Name:        "dns vm app alias 123",
						Description: "description 1",
						VMAppID:     vmAppID,
					}))
			},
			Entry("GET /dns_vm_app_aliases"),
		)
	})

	Describe("GET /dns_vm_app_aliases/{id}", func() {
		var (
			dnsVMAppAliasID string
		)

		BeforeEach(func() {
			dnsVMAppAliasID = postDNSVMAppAliases(vmAppID)
		})

		DescribeTable("200 OK",
			func() {
				dnsVMAppAlias := getDNSVMAppAlias(dnsVMAppAliasID)

				By("Verifying the created DNS vm app alias was returned")
				Expect(dnsVMAppAlias).To(Equal(
					&cce.DNSVMAppAlias{
						ID:          dnsVMAppAliasID,
						Name:        "dns vm app alias 123",
						Description: "description 1",
						VMAppID:     vmAppID,
					},
				))
			},
			Entry("GET /dns_vm_app_aliases/{id}"),
		)

		DescribeTable("404 Not Found",
			func() {
				By("Sending a GET /dns_vm_app_aliases/{id} request")
				resp, err := http.Get(
					fmt.Sprintf(
						"http://127.0.0.1:8080/dns_vm_app_aliases/%s",
						uuid.New()))

				By("Verifying a 404 Not Found response")
				Expect(err).ToNot(HaveOccurred())
				Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
			},
			Entry("GET /dns_vm_app_aliases/{id} with nonexistent ID"),
		)
	})

	Describe("PATCH /dns_vm_app_aliases", func() {
		var (
			dnsVMAppAliasID string
		)

		BeforeEach(func() {
			dnsVMAppAliasID = postDNSVMAppAliases(vmAppID)
		})

		DescribeTable("204 No Content",
			func(reqStr string, expectedAlias *cce.DNSVMAppAlias) {
				By("Sending a PATCH /dns_vm_app_aliases request")
				req, err := http.NewRequest(
					http.MethodPatch,
					"http://127.0.0.1:8080/dns_vm_app_aliases",
					strings.NewReader(fmt.Sprintf(reqStr,
						dnsVMAppAliasID, vmAppID)))
				Expect(err).ToNot(HaveOccurred())

				c := http.Client{}
				resp, err := c.Do(req)
				Expect(err).ToNot(HaveOccurred())

				By("Verifying a 204 No Content response")
				Expect(err).ToNot(HaveOccurred())
				Expect(resp.StatusCode).To(Equal(http.StatusNoContent))

				By("Getting the updated application")
				updatedAlias := getDNSVMAppAlias(dnsVMAppAliasID)

				By("Verifying the DNS vm app alias was updated")
				expectedAlias.SetID(dnsVMAppAliasID)
				expectedAlias.VMAppID = vmAppID
				Expect(updatedAlias).To(Equal(expectedAlias))
			},
			Entry(
				"PATCH /dns_vm_app_aliases",
				`
				[
					{
						"id": "%s",
						"name": "dns vm app alias 123456",
						"description": "description 1",
						"vm_app_id": "%s"
					}
				]`,
				&cce.DNSVMAppAlias{
					Name:        "dns vm app alias 123456",
					Description: "description 1",
				}),
		)

		DescribeTable("400 Bad Request",
			func(reqStr string, expectedResp string) {
				By("Sending a PATCH /dns_vm_app_aliases request")
				if strings.Contains(reqStr, "%s") {
					reqStr = fmt.Sprintf(reqStr, dnsVMAppAliasID)
				}
				req, err := http.NewRequest(
					http.MethodPatch,
					"http://127.0.0.1:8080/dns_vm_app_aliases",
					strings.NewReader(reqStr))
				Expect(err).ToNot(HaveOccurred())

				c := http.Client{}
				resp, err := c.Do(req)

				By("Verifying a 400 Bad Request")
				Expect(err).ToNot(HaveOccurred())
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))

				By("Reading the response body")
				body, err := ioutil.ReadAll(resp.Body)
				Expect(err).ToNot(HaveOccurred())

				By("Verifying the response body")
				Expect(string(body)).To(Equal(expectedResp))
			},
			Entry(
				"PATCH /dns_vm_app_aliases without id",
				`
				[{}]`,
				"Validation failed: id not a valid uuid"),
			Entry(
				"PATCH /dns_vm_app_aliases without name",
				`
				[
					{
						"id": "%s",
						"description": "description 1",
						"vm_app_id": "123"
					}
				]`,
				"Validation failed: name cannot be empty"),
			Entry(
				"PATCH /dns_vm_app_aliases without description",
				`
				[
					{
						"id": "%s",
						"name": "dns vm app alias 123",
						"vm_app_id": "123"
					}
				]`,
				"Validation failed: description cannot be empty"),
			Entry(
				"PATCH /dns_vm_app_aliases without vm_app_id",
				`
				[
					{
						"id": "%s",
						"name": "dns vm app alias 123",
						"description": "description 1"
					}
				]`,
				"Validation failed: vm_app_id not a valid uuid"),
		)
	})

	Describe("DELETE /dns_vm_app_aliases/{id}", func() {
		var (
			dnsVMAppAliasID string
		)

		BeforeEach(func() {
			dnsVMAppAliasID = postDNSVMAppAliases(vmAppID)
		})

		DescribeTable("200 OK",
			func() {
				By("Sending a DELETE /dns_vm_app_aliases/{id} request")
				req, err := http.NewRequest(
					http.MethodDelete,
					fmt.Sprintf(
						"http://127.0.0.1:8080/dns_vm_app_aliases/%s",
						dnsVMAppAliasID),
					nil)
				Expect(err).ToNot(HaveOccurred())

				c := http.Client{}
				resp, err := c.Do(req)
				Expect(err).ToNot(HaveOccurred())

				By("Verifying a 200 OK response")
				Expect(err).ToNot(HaveOccurred())
				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				By("Verifying the DNS vm app alias was deleted")

				By("Sending a GET /dns_vm_app_aliases/{id} request")
				resp, err = http.Get(
					fmt.Sprintf(
						"http://127.0.0.1:8080/dns_vm_app_aliases/%s",
						dnsVMAppAliasID))

				By("Verifying a 404 Not Found response")
				Expect(err).ToNot(HaveOccurred())
				Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
			},
			Entry("DELETE /dns_vm_app_aliases/{id}"),
		)

		DescribeTable("404 Not Found",
			func(id string) {
				By("Sending a DELETE /dns_vm_app_aliases/{id} request")
				req, err := http.NewRequest(
					http.MethodDelete,
					fmt.Sprintf(
						"http://127.0.0.1:8080/dns_vm_app_aliases/%s",
						id),
					nil)
				Expect(err).ToNot(HaveOccurred())

				c := http.Client{}
				resp, err := c.Do(req)
				Expect(err).ToNot(HaveOccurred())

				By("Verifying a 404 Not Found response")
				Expect(err).ToNot(HaveOccurred())
				Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
			},
			Entry(
				"DELETE /dns_vm_app_aliases/{id} with nonexistent ID",
				uuid.New()),
		)

		DescribeTable("422 Unprocessable Entity",
			func() {
				postDNSConfigsDNSVMAppAliases(
					postDNSConfigs(),
					dnsVMAppAliasID)

				By("Sending a DELETE /dns_vm_app_aliases/{id} request")
				req, err := http.NewRequest(
					http.MethodDelete,
					fmt.Sprintf(
						"http://127.0.0.1:8080/dns_vm_app_aliases/%s",
						dnsVMAppAliasID),
					nil)
				Expect(err).ToNot(HaveOccurred())

				c := http.Client{}
				resp, err := c.Do(req)
				Expect(err).ToNot(HaveOccurred())

				By("Verifying a 422 response")
				Expect(resp.StatusCode).To(Equal(
					http.StatusUnprocessableEntity))

				By("Reading the response body")
				body, err := ioutil.ReadAll(resp.Body)
				Expect(err).ToNot(HaveOccurred())

				By("Verifying the response body")
				Expect(string(body)).To(Equal(fmt.Sprintf(
					"cannot delete dns_vm_app_alias_id %s: record in use in "+
						"dns_configs_dns_vm_app_aliases",
					dnsVMAppAliasID)))
			},
			Entry("DELETE /dns_vm_app_aliases/{id} with dns_configs_dns_vm_app_aliases record"), //nolint:lll
		)
	})
})
