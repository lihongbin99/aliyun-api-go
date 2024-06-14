package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	ecs20140526 "github.com/alibabacloud-go/ecs-20140526/v4/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
)

var config = &openapi.Config{
	AccessKeyId:     tea.String(os.Getenv("ALIBABA_CLOUD_ACCESS_KEY_ID")),
	AccessKeySecret: tea.String(os.Getenv("ALIBABA_CLOUD_ACCESS_KEY_SECRET")),
	Endpoint:        tea.String("ecs.cn-beijing.aliyuncs.com"),
}

var client *ecs20140526.Client

func init() {
	var err error
	client, err = ecs20140526.NewClient(config)
	if err != nil {
		panic(err)
	}
}

func main() {
	// 创建实例
	systemDisk := &ecs20140526.CreateInstanceRequestSystemDisk{
		Category: tea.String("cloud_essd"), // 类型
		Size:     tea.Int32(40),            // 容量
	}
	// createInstanceRequest := &ecs20140526.CreateInstanceRequest{
	// 	InstanceChargeType:      tea.String("PostPaid"),                  // 付费类型
	// 	RegionId:                tea.String("cn-beijing"),                // 地域
	// 	ZoneId:                  tea.String("cn-beijing-l"),              // 可用区
	// 	VSwitchId:               tea.String("vsw-2ze328sbseqkdvxjcitmr"), // 交换机
	// 	InstanceType:            tea.String("ecs.gn7i-c8g1.2xlarge"),     // 实例规格
	// 	ImageId:                 tea.String("m-2zeh5n4ljxrqlyun8dx9"),    // 镜像
	// 	SystemDisk:              systemDisk,                              // 系统盘
	// 	InternetChargeType:      tea.String("PayByTraffic"),              // 公网带宽计费模式
	// 	InternetMaxBandwidthIn:  tea.Int32(200),                          // 公网下载带宽
	// 	InternetMaxBandwidthOut: tea.Int32(100),                          // 公网上传带宽
	// 	PasswordInherit:         tea.Bool(true),                          // 登录凭证
	// 	SecurityGroupId:         tea.String("sg-2ze3m4b335au4dr7f5kq"),   // 安全组
	// }

	createInstanceRequest := &ecs20140526.CreateInstanceRequest{
		InstanceChargeType:      tea.String("PostPaid"),                                  // 付费类型
		RegionId:                tea.String("cn-beijing"),                                // 地域
		ZoneId:                  tea.String("cn-beijing-l"),                              // 可用区
		VSwitchId:               tea.String("vsw-2ze328sbseqkdvxjcitmr"),                 // 交换机
		InstanceType:            tea.String("ecs.e-c1m1.large"),                          // 实例规格
		ImageId:                 tea.String("ubuntu_22_04_x64_20G_alibase_20240508.vhd"), // 镜像
		SystemDisk:              systemDisk,                                              // 系统盘
		InternetChargeType:      tea.String("PayByTraffic"),                              // 公网带宽计费模式
		InternetMaxBandwidthIn:  tea.Int32(200),                                          // 公网下载带宽
		InternetMaxBandwidthOut: tea.Int32(100),                                          // 公网上传带宽
		Password:                tea.String("hznWdg$N^JDJWQ3T"),                          // 登录凭证
		SecurityGroupId:         tea.String("sg-2ze3m4b335au4dr7f5kq"),                   // 安全组
	}
	fmt.Println("create instance")
	instanceId, err := CreateInstance(createInstanceRequest)
	if err != nil {
		panic(err)
	}
	fmt.Println("instance id: ", instanceId)

	// 查询所有实例
	if instances, err := DescribeInstances("cn-beijing"); err != nil {
		fmt.Println(err)
	} else {
		instance := instances.Body.Instances.Instance
		for i := 0; i < len(instance); i++ {
			fmt.Println("all instance id: ", *instance[i].InstanceId)
		}
	}

	// 分配公网ip
	if ip, err := AllocatePublicIpAddress(instanceId); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("allocate public ip address: ", ip)
	}

	// 启动实例
	time.Sleep(10 * time.Second)
	fmt.Println("start instance...")
	for {
		time.Sleep(1 * time.Second)
		err = StartInstance(instanceId)
		if err != nil {
			if strings.Contains(err.Error(), "does not support this operation.") {
				continue
			}
			fmt.Println(err)
		}
		break
	}
	fmt.Println("start instance success")

	// 停止实例
	time.Sleep(10 * time.Second)
	fmt.Println("stop instance...")
	for {
		time.Sleep(1 * time.Second)
		err = StopInstance(instanceId, true, "StopCharging") // KeepCharging,StopCharging
		if err != nil {
			if strings.Contains(err.Error(), "ecs task is conflicted.") {
				continue
			}
			fmt.Println(err)
		}
		break
	}
	fmt.Println("stop instance success")

	// 删除实例
	time.Sleep(10 * time.Second)
	fmt.Println("delete instance...")
	for {
		time.Sleep(1 * time.Second)
		err = DeleteInstance(instanceId)
		if err != nil {
			if strings.Contains(err.Error(), "does not support this operation.") {
				continue
			}
			fmt.Println(err)
		}
		break
	}
	fmt.Println("delete instance sucess")
}

func DescribeInstances(regionId string) (*ecs20140526.DescribeInstancesResponse, error) {
	describeInstancesRequest := &ecs20140526.DescribeInstancesRequest{
		RegionId: tea.String(regionId),
	}
	runtime := &util.RuntimeOptions{}
	return client.DescribeInstancesWithOptions(describeInstancesRequest, runtime)
}

func CreateInstance(request *ecs20140526.CreateInstanceRequest) (string, error) {
	runtime := &util.RuntimeOptions{}
	createInstanceResponse, err := client.CreateInstanceWithOptions(request, runtime)
	if err != nil {
		return "", err
	}
	return *createInstanceResponse.Body.InstanceId, nil
}

func AllocatePublicIpAddress(instanceId string) (string, error) {
	allocatePublicIpAddressRequest := &ecs20140526.AllocatePublicIpAddressRequest{
		InstanceId: tea.String(instanceId),
	}
	runtime := &util.RuntimeOptions{}
	allocatePublicIpAddressResponse, err := client.AllocatePublicIpAddressWithOptions(allocatePublicIpAddressRequest, runtime)
	if err != nil {
		return "", err
	}
	return *allocatePublicIpAddressResponse.Body.IpAddress, nil
}

func StartInstance(instanceId string) error {
	startInstanceRequest := &ecs20140526.StartInstanceRequest{
		InstanceId: tea.String(instanceId),
	}
	runtime := &util.RuntimeOptions{}
	_, err := client.StartInstanceWithOptions(startInstanceRequest, runtime)
	return err
}

func StopInstance(instanceId string, forceStop bool, stoppedMode string) error {
	stopInstanceRequest := &ecs20140526.StopInstanceRequest{
		InstanceId:  tea.String(instanceId),
		ForceStop:   tea.Bool(forceStop),
		StoppedMode: tea.String(stoppedMode),
	}
	runtime := &util.RuntimeOptions{}
	_, err := client.StopInstanceWithOptions(stopInstanceRequest, runtime)
	return err
}

func DeleteInstance(instanceId string) error {
	deleteInstanceRequest := &ecs20140526.DeleteInstanceRequest{
		InstanceId: tea.String(instanceId), // 实例ID
		Force:      tea.Bool(true),
	}
	runtime := &util.RuntimeOptions{}
	_, err := client.DeleteInstanceWithOptions(deleteInstanceRequest, runtime)
	return err
}
