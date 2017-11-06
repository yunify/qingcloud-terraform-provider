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

func TestAccQingcloudTag_basic(t *testing.T) {
	var tag qc.DescribeTagsOutput
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		IDRefreshName: "qingcloud_tag.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckTagDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccTagConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTagExists("qingcloud_tag.foo", &tag),
					resource.TestCheckResourceAttr(
						"qingcloud_tag.foo", "name", "tag1"),
					resource.TestCheckResourceAttr(
						"qingcloud_tag.foo", "color", "#9f9bb7"),
				),
			},
			resource.TestStep{
				Config: testAccTagConfigTwo,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTagExists("qingcloud_tag.foo", &tag),
					resource.TestCheckResourceAttr(
						"qingcloud_tag.foo", "name", "tag1"),
					resource.TestCheckResourceAttr(
						"qingcloud_tag.foo", "description", "test"),
					resource.TestCheckResourceAttr(
						"qingcloud_tag.foo", "color", "#fff"),
				),
			},
		},
	})
}

func testAccCheckTagExists(n string, tag *qc.DescribeTagsOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EIP ID is set")
		}

		client := testAccProvider.Meta().(*QingCloudClient)
		input := new(qc.DescribeTagsInput)
		input.Tags = []*string{qc.String(rs.Primary.ID)}
		d, err := client.tag.DescribeTags(input)

		log.Printf("[WARN] tag id %#v", rs.Primary.ID)

		if err != nil {
			return err
		}

		if d == nil || qc.StringValue(d.TagSet[0].TagID) == "" {
			return fmt.Errorf("tag not found")
		}

		*tag = *d
		return nil
	}
}

func testAccCheckTagDestroy(s *terraform.State) error {
	return testAccCheckTagDestroyWithProvider(s, testAccProvider)
}

func testAccCheckTagDestroyWithProvider(s *terraform.State, provider *schema.Provider) error {
	client := provider.Meta().(*QingCloudClient)
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "qingcloud_tag" {
			continue
		}

		// Try to find the resource
		input := new(qc.DescribeTagsInput)
		input.Tags = []*string{qc.String(rs.Primary.ID)}
		output, err := client.tag.DescribeTags(input)
		if err == nil && qc.IntValue(output.RetCode) == 0 {
			if len(output.TagSet) != 0 {
				return fmt.Errorf("Found  tag: %s", rs.Primary.ID)
			}
		}
	}
	return nil
}

const testAccTagConfig = `
resource "qingcloud_tag" "foo"{
	name="tag1"
}
`
const testAccTagConfigTwo = `
resource "qingcloud_tag" "foo"{
	name="tag1"
	description="test"
	color = "#fff"
}
`
