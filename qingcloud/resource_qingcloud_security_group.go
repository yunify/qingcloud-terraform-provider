package qingcloud

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"

	qc "github.com/yunify/qingcloud-sdk-go/service"
)

func resourceQingcloudSecurityGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceQingcloudSecurityGroupCreate,
		Read:   resourceQingcloudSecurityGroupRead,
		Update: resourceQingcloudSecurityGroupUpdate,
		Delete: resourceQingcloudSecurityGroupDelete,
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of SecurityGroup ",
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The description of SecurityGroup",
			},
			"tag_ids": &schema.Schema{
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Description: "tag ids , SecurityGroup wants to use",
			},
			"tag_names": &schema.Schema{
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
		},
	}
}

func resourceQingcloudSecurityGroupCreate(d *schema.ResourceData, meta interface{}) error {
	clt := meta.(*QingCloudClient).securitygroup
	input := new(qc.CreateSecurityGroupInput)
	input.SecurityGroupName = qc.String(d.Get("name").(string))
	var output *qc.CreateSecurityGroupOutput
	var err error
	simpleRetry(func() error {
		output, err = clt.CreateSecurityGroup(input)
		if err == nil {
			if output.RetCode != nil && IsServerBusy(*output.RetCode) {
				return fmt.Errorf("Server Busy")
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("Error create security group: %s", err)
	}
	if output.RetCode != nil && qc.IntValue(output.RetCode) != 0 {
		return fmt.Errorf("Error create security group: %s", *output.Message)
	}
	d.SetId(qc.StringValue(output.SecurityGroupID))
	return resourceQingcloudSecurityGroupUpdate(d, meta)
}

func resourceQingcloudSecurityGroupRead(d *schema.ResourceData, meta interface{}) error {
	clt := meta.(*QingCloudClient).securitygroup
	input := new(qc.DescribeSecurityGroupsInput)
	input.SecurityGroups = []*string{qc.String(d.Id())}
	var output *qc.DescribeSecurityGroupsOutput
	var err error
	simpleRetry(func() error {
		output, err = clt.DescribeSecurityGroups(input)
		if err == nil {
			if output.RetCode != nil && IsServerBusy(*output.RetCode) {
				return fmt.Errorf("Server Busy")
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("Error describe security group: %s", err)
	}
	if output.RetCode != nil && qc.IntValue(output.RetCode) != 0 {
		return fmt.Errorf("Error describe security group: %s", *output.Message)
	}
	if len(output.SecurityGroupSet) == 0 {
		d.SetId("")
		return nil
	}
	sg := output.SecurityGroupSet[0]
	d.Set("name", qc.StringValue(sg.SecurityGroupName))
	d.Set("description", qc.StringValue(sg.Description))
	resourceSetTag(d, sg.Tags)
	return nil
}
func resourceQingcloudSecurityGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	d.Partial(true)
	err := modifySecurityGroupAttributes(d, meta)
	if err != nil {
		return err
	}
	d.SetPartial("description")
	d.SetPartial("name")
	if err := resourceUpdateTag(d, meta, qingcloudResourceTypeSecurityGroup); err != nil {
		return err
	}
	d.SetPartial("tag_ids")
	d.Partial(false)
	return resourceQingcloudSecurityGroupRead(d, meta)
}

func resourceQingcloudSecurityGroupDelete(d *schema.ResourceData, meta interface{}) error {
	clt := meta.(*QingCloudClient).securitygroup
	describeSecurityGroupInput := new(qc.DescribeSecurityGroupsInput)
	describeSecurityGroupInput.SecurityGroups = []*string{qc.String(d.Id())}
	describeSecurityGroupInput.Verbose = qc.Int(1)
	var describeSecurityGroupOutput *qc.DescribeSecurityGroupsOutput
	var err error
	simpleRetry(func() error {
		describeSecurityGroupOutput, err = clt.DescribeSecurityGroups(describeSecurityGroupInput)
		if err == nil {
			if describeSecurityGroupOutput.RetCode != nil && IsServerBusy(*describeSecurityGroupOutput.RetCode) {
				return fmt.Errorf("Server Busy")
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("Error describe security group: %s", err)
	}
	if describeSecurityGroupOutput.RetCode != nil && qc.IntValue(describeSecurityGroupOutput.RetCode) != 0 {
		return fmt.Errorf("Error describe security group: %s", err)
	}
	if len(describeSecurityGroupOutput.SecurityGroupSet[0].Resources) > 0 {
		return fmt.Errorf("Error security group %s is using, can't delete", d.Id())
	}
	input := new(qc.DeleteSecurityGroupsInput)
	input.SecurityGroups = []*string{qc.String(d.Id())}
	var output *qc.DeleteSecurityGroupsOutput
	simpleRetry(func() error {
		output, err = clt.DeleteSecurityGroups(input)
		if err == nil {
			if output.RetCode != nil && IsServerBusy(*output.RetCode) {
				return fmt.Errorf("Server Busy")
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("Error delete security group: %s", err)
	}
	if output.RetCode != nil && qc.IntValue(output.RetCode) != 0 {
		return fmt.Errorf("Error delete security group: %s", *output.Message)
	}
	d.SetId("")
	return nil
}
