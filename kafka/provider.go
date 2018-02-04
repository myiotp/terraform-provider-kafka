package kafka

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"brokers": {
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Required:    true,
				Description: "A list of kafka brokers",
			},
			"timeout": {
				Type:        schema.TypeInt,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("KAFKA_TIMEOUT", 10000),
				Description: "A csv list of kafka brokers",
			},
		},

		ConfigureFunc: providerConfigure,
		ResourcesMap: map[string]*schema.Resource{
			"kafka_topic": kafkaTopicResource(),
		},
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	var brokers *[]string

	if brokersRaw, ok := d.GetOk("brokers"); ok {
		brokerI := brokersRaw.([]interface{})
		log.Printf("[DEBUG] configuring provider with Brokers of size %d", len(brokerI))
		b := make([]string, len(brokerI))
		for i, v := range brokerI {
			b[i] = v.(string)
		}
		log.Printf("[DEBUG] b of size %d", len(b))
		brokers = &b
	} else {
		log.Printf("[ERROR] something wrong? %v , ", d.Get("timeout"))
		return nil, fmt.Errorf("brokers was not set")
	}

	log.Printf("[DEBUG] configuring provider with Brokers @ %v", brokers)

	config := &Config{
		Brokers: brokers,
	}

	log.Printf("[DEBUG] Config @ %v", config)

	return NewClient(config)
}
