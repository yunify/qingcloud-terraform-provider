package qingcloud

import (
	"github.com/hashicorp/terraform/helper/schema"
	qc "github.com/yunify/qingcloud-sdk-go/service"
)

const (
	resourceSecurityGroupRuleSecurityGroupID = "security_group_id"
	resourceSecurityGroupRuleProtocol        = "protocol"
	resourceSecurityGroupRulePriority        = "priority"
	resourceSecurityGroupRuleAction          = "action"
	resourceSecurityGroupRuleDirection       = "direction"
	resourceSecurityGroupRuleFromPort        = "from_port"
	resourceSecurityGroupRuleToPort          = "to_port"
	resourceSecurityGroupCidrBlock           = "cidr_block"
)

func resourceQingcloudSecurityGroupRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceQingcloudSecurityGroupRuleCreate,
		Read:   resourceQingcloudSecurityGroupRuleRead,
		Update: resourceQingcloudSecurityGroupRuleUpdate,
		Delete: resourceQingcloudSecurityGroupRuleDelete,
		Schema: map[string]*schema.Schema{
			resourceSecurityGroupRuleSecurityGroupID: {
				Type:     schema.TypeString,
				Required: true,
			},
			resourceName: {
				Type:     schema.TypeString,
				Optional: true,
			},
			resourceSecurityGroupCidrBlock: {
				Type:     schema.TypeString,
				Optional: true,
			},
			resourceSecurityGroupRuleProtocol: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: withinArrayString("tcp", "udp", "icmp", "gre", "esp", "ah", "ipip"),
			},
			resourceSecurityGroupRulePriority: {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: withinArrayIntRange(0, 100),
				Default:      0,
			},
			resourceSecurityGroupRuleAction: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: withinArrayString("accept", "drop"),
			},
			resourceSecurityGroupRuleDirection: {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: withinArrayInt(0, 1),
				Default:      0,
			},
			resourceSecurityGroupRuleFromPort: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validatePortString,
			},
			resourceSecurityGroupRuleToPort: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validatePortString,
			},
		},
	}
}

func resourceQingcloudSecurityGroupRuleCreate(d *schema.ResourceData, meta interface{}) error {
	clt := meta.(*QingCloudClient).securitygroup
	input := new(qc.AddSecurityGroupRulesInput)
	sgID := d.Get(resourceSecurityGroupRuleSecurityGroupID).(string)
	input.SecurityGroup = qc.String(sgID)
	rule := new(qc.SecurityGroupRule)
	rule.Priority = qc.Int(d.Get(resourceSecurityGroupRulePriority).(int))
	rule.Protocol = qc.String(d.Get(resourceSecurityGroupRuleProtocol).(string))
	rule.Action = qc.String(d.Get(resourceSecurityGroupRuleAction).(string))
	rule.Direction = qc.Int(d.Get(resourceSecurityGroupRuleDirection).(int))
	rule.SecurityGroupRuleName, _ = getNamePointer(d)
	rule.Val1 = getSetStringPointer(d, resourceSecurityGroupRuleFromPort)
	rule.Val2 = getSetStringPointer(d, resourceSecurityGroupRuleToPort)
	rule.Val3 = getSetStringPointer(d, resourceSecurityGroupCidrBlock)
	input.Rules = []*qc.SecurityGroupRule{rule}
	var output *qc.AddSecurityGroupRulesOutput
	var err error
	simpleRetry(func() error {
		output, err = clt.AddSecurityGroupRules(input)
		return isServerBusy(err)
	})
	if err != nil {
		return err
	}
	d.SetId(qc.StringValue(output.SecurityGroupRules[0]))
	if err := applySecurityGroupRule(qc.String(d.Get(resourceSecurityGroupRuleSecurityGroupID).(string)), meta); err != nil {
		return nil
	}
	return resourceQingcloudSecurityGroupRuleRead(d, meta)
}

func resourceQingcloudSecurityGroupRuleRead(d *schema.ResourceData, meta interface{}) error {
	clt := meta.(*QingCloudClient).securitygroup
	input := new(qc.DescribeSecurityGroupRulesInput)
	input.SecurityGroupRules = []*string{qc.String(d.Id())}
	var output *qc.DescribeSecurityGroupRulesOutput
	var err error
	simpleRetry(func() error {
		output, err = clt.DescribeSecurityGroupRules(input)
		return isServerBusy(err)
	})
	if err != nil {
		return err
	}
	if len(output.SecurityGroupRuleSet) == 0 {
		d.SetId("")
		return nil
	}
	sgRule := output.SecurityGroupRuleSet[0]
	d.Set(resourceSecurityGroupRuleSecurityGroupID, qc.StringValue(sgRule.SecurityGroupID))
	d.Set(resourceSecurityGroupRuleProtocol, qc.StringValue(sgRule.Protocol))
	d.Set(resourceSecurityGroupRulePriority, qc.IntValue(sgRule.Priority))
	d.Set(resourceSecurityGroupRuleAction, qc.StringValue(sgRule.Action))
	d.Set(resourceSecurityGroupRuleFromPort, qc.StringValue(sgRule.Val1))
	d.Set(resourceSecurityGroupRuleToPort, qc.StringValue(sgRule.Val2))
	d.Set(resourceSecurityGroupCidrBlock, qc.StringValue(sgRule.Val3))
	d.Set(resourceName, qc.StringValue(sgRule.SecurityGroupRuleName))
	return nil
}

func resourceQingcloudSecurityGroupRuleUpdate(d *schema.ResourceData, meta interface{}) error {
	if err := ModifySecurityGroupRuleAttributes(d, meta); err != nil {
		return err
	}
	if err := applySecurityGroupRule(qc.String(d.Get(resourceSecurityGroupRuleSecurityGroupID).(string)), meta); err != nil {
		return err
	}
	return resourceQingcloudSecurityGroupRuleRead(d, meta)
}

func resourceQingcloudSecurityGroupRuleDelete(d *schema.ResourceData, meta interface{}) error {
	clt := meta.(*QingCloudClient).securitygroup
	input := new(qc.DeleteSecurityGroupRulesInput)
	input.SecurityGroupRules = []*string{qc.String(d.Id())}
	var err error
	simpleRetry(func() error {
		_, err = clt.DeleteSecurityGroupRules(input)
		return isServerBusy(err)
	})
	if err != nil {
		return err
	}
	if err := applySecurityGroupRule(qc.String(d.Get(resourceSecurityGroupRuleSecurityGroupID).(string)), meta); err != nil {
		return nil
	}
	d.SetId("")
	return nil
}
