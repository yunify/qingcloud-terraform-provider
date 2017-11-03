package qingcloud

import (
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	qc "github.com/yunify/qingcloud-sdk-go/service"
)

func TestAccQingcloudEIP_basic(t *testing.T) {
	var eip qc.DescribeEIPsOutput

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},

		// module name
		IDRefreshName: "qingcloud_eip.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckEIPDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccEIPConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(
						"qingcloud_eip.foo", &eip),
					resource.TestCheckResourceAttr(
						"qingcloud_eip.foo", "bandwidth", "2"),
					resource.TestCheckResourceAttr(
						"qingcloud_eip.foo", "billing_mode", "traffic"),
					resource.TestCheckResourceAttr(
						"qingcloud_eip.foo", "description", "first"),
					resource.TestCheckResourceAttr(
						"qingcloud_eip.foo", "name", "first_eip"),
					resource.TestCheckResourceAttr(
						"qingcloud_eip.foo", "need_icp", "0"),
				),
			},
			resource.TestStep{
				Config: testAccEIPConfigTwo,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(
						"qingcloud_eip.foo", &eip),
					resource.TestCheckResourceAttr(
						"qingcloud_eip.foo", "bandwidth", "4"),
					resource.TestCheckResourceAttr(
						"qingcloud_eip.foo", "billing_mode", "bandwidth"),
					resource.TestCheckResourceAttr(
						"qingcloud_eip.foo", "description", "eip"),
					resource.TestCheckResourceAttr(
						"qingcloud_eip.foo", "name", "eip"),
					resource.TestCheckResourceAttr(
						"qingcloud_eip.foo", "need_icp", "0"),
				),
			},
		},
	})

}
func TestAccQingcloudEIP_tag(t *testing.T) {
	var eip qc.DescribeEIPsOutput

	testTagNameValue := func(names ...string) resource.TestCheckFunc {
		return func(state *terraform.State) error {
			tags := eip.EIPSet[0].Tags
			same_count := 0
			for _, tag := range tags {
				for _, name := range names {
					if qc.StringValue(tag.TagName) == name {
						same_count++
					}
					if same_count == len(eip.EIPSet[0].Tags) {
						return nil
					}
				}
			}
			return fmt.Errorf("tag name error %#v", names)
		}
	}
	testTagDetach := func() resource.TestCheckFunc {
		return func(state *terraform.State) error {
			if len(eip.EIPSet[0].Tags) != 0 {
				return fmt.Errorf("tag not detach ")
			}
			return nil
		}
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		IDRefreshName: "qingcloud_eip.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckEIPDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccEipConfigTag,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(
						"qingcloud_eip.foo", &eip),
					testTagNameValue("11", "22"),
				),
			},
			resource.TestStep{
				Config: testAccEipConfigTagTwo,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEIPExists(
						"qingcloud_eip.foo", &eip),
					testTagDetach(),
				),
			},
		},
	})

}

func testAccCheckEIPExists(n string, eip *qc.DescribeEIPsOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EIP ID is set")
		}

		client := testAccProvider.Meta().(*QingCloudClient)
		input := new(qc.DescribeEIPsInput)
		input.EIPs = []*string{qc.String(rs.Primary.ID)}
		d, err := client.eip.DescribeEIPs(input)

		log.Printf("[WARN] eip id %#v", rs.Primary.ID)

		if err != nil {
			return err
		}

		if d == nil || qc.StringValue(d.EIPSet[0].EIPAddr) == "" {
			return fmt.Errorf("EIP not found")
		}

		*eip = *d
		return nil
	}
}
func testAccCheckEIPDestroy(s *terraform.State) error {
	return testAccCheckEIPDestroyWithProvider(s, testAccProvider)
}

func testAccCheckEIPDestroyWithProvider(s *terraform.State, provider *schema.Provider) error {
	client := provider.Meta().(*QingCloudClient)
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "qingcloud_eip" {
			continue
		}

		// Try to find the resource
		input := new(qc.DescribeEIPsInput)
		input.EIPs = []*string{qc.String(rs.Primary.ID)}
		output, err := client.eip.DescribeEIPs(input)
		if err == nil && qc.IntValue(output.RetCode) == 0 {
			if len(output.EIPSet) != 0 && qc.StringValue(output.EIPSet[0].Status) != "released" {
				return fmt.Errorf("Found  EIP: %s", rs.Primary.ID)
			}
		}
	}
	return nil
}

const testAccEIPConfig = `
resource "qingcloud_eip" "foo" {
    name = "first_eip"
    description = "first"
    billing_mode = "traffic"
    bandwidth = 2
    need_icp = 0
} `
const testAccEIPConfigTwo = `
resource "qingcloud_eip" "foo" {
    name = "eip"
    description = "eip"
    billing_mode = "bandwidth"
    bandwidth = 4
    need_icp = 0
} `

const testAccEipConfigTag = `

resource "qingcloud_eip" "foo" {
    name = "eip"
    description = "eip"
    billing_mode = "bandwidth"
    bandwidth = 4
    need_icp = 0
	tag_ids = ["${qingcloud_tag.test.id}",
				"${qingcloud_tag.test2.id}"]
}
resource "qingcloud_tag" "test"{
	name="11"
}
resource "qingcloud_tag" "test2"{
	name="22"
}
`
const testAccEipConfigTagTwo = `

resource "qingcloud_eip" "foo" {
    name = "eip"
    description = "eip"
    billing_mode = "bandwidth"
    bandwidth = 4
    need_icp = 0
}
resource "qingcloud_tag" "test"{
	name="11"
}
resource "qingcloud_tag" "test2"{
	name="22"
}
`
