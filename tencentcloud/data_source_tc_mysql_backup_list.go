package tencentcloud

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceTencentMysqlBackupList() *schema.Resource {

	return &schema.Resource{
		Read: dataSourceTencentMysqlBackupListRead,
		Schema: map[string]*schema.Schema{
			"mysql_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"max_number": {
				Type:         schema.TypeInt,
				ForceNew:     true,
				Optional:     true,
				Default:      10,
				ValidateFunc: validateIntegerInRange(1, 10000),
			},
			"result_output_file": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
			},
			// Computed values
			"list": {Type: schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"time": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"finish_time": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"size": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"backup_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"backup_model": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"intranet_url": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"internet_url": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"creator": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceTencentMysqlBackupListRead(d *schema.ResourceData, meta interface{}) error {
	defer LogElapsed("data_source.tencentcloud_mysql_backup_list.read")()

	logId := GetLogId(nil)
	ctx := context.WithValue(context.TODO(), "logId", logId)

	mysqlService := MysqlService{client: meta.(*TencentCloudClient).apiV3Conn}

	max_number, _ := d.Get("max_number").(int)
	backInfoItems, err := mysqlService.DescribeBackupsByMysqlId(ctx, d.Get("mysql_id").(string), int64(max_number))

	if err != nil {
		return fmt.Errorf("api[DescribeBackups]fail, return %s", err.Error())
	}

	var itemShemas []map[string]interface{}
	var ids = make([]string, len(backInfoItems))

	for index, item := range backInfoItems {
		mapping := map[string]interface{}{
			"time":         *item.Date,
			"finish_time":  *item.FinishTime,
			"size":         *item.Size,
			"backup_id":    *item.BackupId,
			"backup_model": *item.Type,
			"intranet_url": strings.Replace(*item.IntranetUrl, "\u0026", "&", -1),
			"internet_url": strings.Replace(*item.InternetUrl, "\u0026", "&", -1),
			"creator":      *item.Creator,
		}
		ids[index] = fmt.Sprintf("%d", *item.BackupId)
		itemShemas = append(itemShemas, mapping)
	}

	if err := d.Set("list", itemShemas); err != nil {
		log.Printf("[CRITAL]%s provider set itemShemas fail, reason:%s\n ", logId, err.Error())
	}
	d.SetId(dataResourceIdsHash(ids))

	if output, ok := d.GetOk("result_output_file"); ok && output.(string) != "" {

		if err := writeToFile(output.(string), itemShemas); err != nil {
			log.Printf("[CRITAL]%s output file[%s] fail,  reason[%s]\n",
				logId, output.(string), err.Error())
		}

	}
	return nil
}