package kafka

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func kafkaTopicResource() *schema.Resource {
	return &schema.Resource{
		Create: topicCreate,
		Read:   topicRead,
		Update: topicUpdate,
		Delete: topicDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the topic",
			},
			"partitions": {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "number of partitions",
			},
			"replication_factor": {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "number of replicas",
			},
			"config": {
				Type:        schema.TypeMap,
				Optional:    true,
				ForceNew:    false,
				Description: "A map of string k/v attributes",
			},
		},
	}
}

func topicCreate(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*Client)
	t := metaToTopic(d, meta)

	for i, b := range *c.config.Brokers {
		log.Printf("[DEBUG] Brokers %d , %s", i, b)
	}
	err := c.CreateTopic(t)

	if err != nil {
		return err
	}

	d.SetId(t.Name)
	return nil
}

func topicRefreshFunc(client *Client, topic string, expected Topic) resource.StateRefreshFunc {
	return func() (result interface{}, s string, err error) {
		log.Printf("[DEBUG] waiting for topic to update %s", topic)
		actual, err := client.ReadTopic(topic)
		if err != nil {
			log.Printf("[ERROR] could not read topic %s, %s", topic, err)
			return actual, "Error", err
		}

		if expected.Equal(actual) {
			return actual, "Ready", nil
		}

		return nil, fmt.Sprintf("%v != %v", strPtrMapToStrMap(actual.Config), strPtrMapToStrMap(expected.Config)), nil
	}
}

func topicUpdate(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*Client)
	t := metaToTopic(d, meta)

	err := c.UpdateTopic(t)

	if err != nil {
		return err
	}

	stateConf := &resource.StateChangeConf{
		Pending:      []string{"Updating"},
		Target:       []string{"Ready"},
		Refresh:      topicRefreshFunc(c, d.Id(), t),
		Timeout:      10 * time.Second,
		Delay:        1 * time.Second,
		PollInterval: 1 * time.Second,
		MinTimeout:   2 * time.Second,
	}

	_, err = stateConf.WaitForState()

	if err != nil {
		return fmt.Errorf(
			"Error waiting for topic (%s) to become ready: %s",
			d.Id(), err)
	}

	return err
}

func topicDelete(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*Client)
	t := metaToTopic(d, meta)

	err := c.DeleteTopcic(t.Name)
	if err != nil {
		return err
	}

	stateConf := &resource.StateChangeConf{
		Pending:      []string{"Pending"},
		Target:       []string{"Deleted"},
		Refresh:      topicDeleteFunc(c, d.Id(), t),
		Timeout:      300 * time.Second,
		Delay:        3 * time.Second,
		PollInterval: 2 * time.Second,
		MinTimeout:   20 * time.Second,
	}
	_, err = stateConf.WaitForState()

	if err != nil {
		return fmt.Errorf("Error waiting for topic (%s) to delete: %s", d.Id(), err)
	}

	d.SetId("")
	return err
}

func topicDeleteFunc(client *Client, id string, t Topic) resource.StateRefreshFunc {
	return func() (result interface{}, s string, err error) {
		topic, err := client.ReadTopic(t.Name)

		if err != nil {
			_, ok := err.(TopicMissingError)
			if ok {
				return topic, "Deleted", nil
			}
			return topic, "UNKNOWN", err
		}
		return topic, "Pending", nil
	}

}

func topicRead(d *schema.ResourceData, meta interface{}) error {
	name := d.Id()
	client := meta.(*Client)
	topic, err := client.ReadTopic(name)

	if err != nil {
		log.Printf("[ERROR] Error getting topics %s from Kafka", err)
		_, ok := err.(TopicMissingError)
		if ok {
			d.SetId("")
			return nil
		}

		return err
	}

	log.Printf("[Debug] Setting the state from Kafka %v", topic)
	d.Set("name", topic.Name)
	d.Set("partitions", topic.Partitions)
	d.Set("replication_factor", topic.ReplicationFactor)
	d.Set("config", topic.Config)

	return nil
}
