package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"reflect"

	"github.com/pkg/errors"
	"github.com/vmware/cloud-provider-for-cloud-director/pkg/vcdsdk"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

var cloudinit = `
#cloud-config
manual_cache_clean: True
`

func main() {
	vAppName := "shalin-manual-test"
	site := "https://vcd.hw.ovh-ca.d2iq.cloud"
	org := "konvoy"
	ovdc := "konvoy"
	username := ""
	password := ""
	refreshToken := os.Getenv("VCD_REFRESH_TOKEN")
	catalog := "default"
	template := "konvoy-rhel-84-release-1.26.6-20230718040729-with-password"
	network := "private-0"
	workloadVCDClient, err := vcdsdk.NewVCDClientFromSecrets(
		site,
		org,
		ovdc,
		org,
		username,
		password,
		refreshToken,
		true,  //insecure
		false) // getvdc client
	if err != nil {
		log.Fatalf("error creating VCD client for [%s] app: [%v]", vAppName, err)
	}

	vdcManager, err := vcdsdk.NewVDCManager(workloadVCDClient, workloadVCDClient.ClusterOrgName,
		workloadVCDClient.ClusterOVDCName)
	if err != nil {
		log.Fatalf("unable to create vdc manager: %w", err)
	}
	vApp, err := vdcManager.GetOrCreateVApp(
		vAppName,
		network,
	)
	if err != nil {
		log.Fatalf("unable to create vapp: %w", err)
	}
	fmt.Println("app created", vApp.VApp.Name)
	//create custom VM
	// err = vdcManager.AddNewTkgVM(
	// 	vAppName+"-poweron",
	// 	vAppName,
	// 	1,
	// 	catalog,
	// 	template,
	// 	"",   // placement policy
	// 	"",   // sizing policy
	// 	"",   // storage profile
	// 	true) // poweron

	// if err != nil {
	// 	log.Fatalf("unable to create Power On vm: %w", err)
	// }
	// fmt.Println("Power on VM created")

	// err = vdcManager.AddNewTkgVM(
	// 	vAppName+"-poweroff",
	// 	vAppName,
	// 	1,
	// 	catalog,
	// 	template,
	// 	"",    // placement policy
	// 	"",    // sizing policy
	// 	"",    // storage profile
	// 	false) // poweron

	// if err != nil {
	// 	log.Fatalf("unable to create Power off vm: %w", err)
	// }
	// fmt.Println("Power off VM created")

	// // compose the VM and then power on. similar to capvcd
	// composeVMName := vAppName + "-compose-on"
	// err = vdcManager.AddNewTkgVM(
	// 	composeVMName,
	// 	vAppName,
	// 	1,
	// 	catalog,
	// 	template,
	// 	"",    // placement policy
	// 	"",    // sizing policy
	// 	"",    // storage profile
	// 	false) // poweron

	// if err != nil {
	// 	log.Fatalf("unable to create Power off vm: %w", err)
	// }
	// vm, err := vApp.GetVMByName(composeVMName, true)
	// if err != nil {
	// 	log.Fatalf("unable to get VM object: %w", err)
	// }

	// task, err := vm.PowerOn()
	// if err != nil {
	// 	log.Fatalf("unable to power on the vm: %w", err)
	// }
	// if err = task.WaitTaskCompletion(); err != nil {
	// 	log.Fatalf("error while waiting for vm to powerup: %w", err)
	// }
	// fmt.Println("Composing the VM and powered on the VM completed")

	// // compose the VM and then power on with customization.
	// composeCustomizeVMName := vAppName + "-compose-on-customize"
	// err = vdcManager.AddNewTkgVM(
	// 	composeCustomizeVMName,
	// 	vAppName,
	// 	1,
	// 	catalog,
	// 	template,
	// 	"",    // placement policy
	// 	"",    // sizing policy
	// 	"",    // storage profile
	// 	false) // poweron

	// if err != nil {
	// 	log.Fatalf("unable to create Power off vm for customization: %w", err)
	// }
	// customizationVM, err := vApp.GetVMByName(composeCustomizeVMName, true)
	// if err != nil {
	// 	log.Fatalf("unable to get customization VM object: %w", err)
	// }

	// err = customizationVM.PowerOnAndForceCustomization()
	// if err != nil {
	// 	log.Fatalf("unable to power on the vm with customization: %w", err)
	// }
	// vmStatus, err := customizationVM.GetStatus()
	// if err != nil {
	// 	log.Fatalf("unable to get vm [%s] status after powering on: [%v]", customizationVM.VM.Name, err)
	// }
	// fmt.Println("Composing the VM and powered on the VM completed. VM Status.", vmStatus)

	// compose the VM, set user data then power on with customization.
	composeCustomizeCloudInitVMName := vAppName + "-compose-dhcp"
	err = vdcManager.AddNewTkgVM(
		composeCustomizeCloudInitVMName,
		vAppName,
		1,
		catalog,
		template,
		"",    // placement policy
		"",    // sizing policy
		"",    // storage profile
		false) // poweron

	if err != nil {
		log.Fatalf("unable to create Power off vm for cloudinit customization: %w", err)
	}
	cloudInitVM, err := vApp.GetVMByName(composeCustomizeCloudInitVMName, true)
	if err != nil {
		log.Fatalf("unable to get customization cloudinit VM object: %w", err)
	}
	cloudInitBytes := []byte(cloudinit)
	b64CloudInitScript := base64.StdEncoding.EncodeToString(cloudInitBytes)
	keyVals := map[string]string{
		"guestinfo.userdata":          b64CloudInitScript,
		"guestinfo.userdata.encoding": "base64",
		"disk.enableUUID":             "1",
	}

	err = reconcileVMNetworks(vdcManager,
		vApp,
		cloudInitVM,
		[]string{network})
	if err != nil {
		log.Fatalf("error setting DHCP on the network: %w", err)
	}

	for key, val := range keyVals {
		err = vdcManager.SetVmExtraConfigKeyValue(cloudInitVM, key, val, true)
		if err != nil {
			log.Fatalf("unable to set vm config %s=%s on VM %s. error: %w", key, val, cloudInitVM.VM.Name, err)
		}
	}

	err = cloudInitVM.PowerOnAndForceCustomization()
	if err != nil {
		log.Fatalf("unable to power on the vm with customization: %w", err)
	}
	vmStatus, err := cloudInitVM.GetStatus()
	if err != nil {
		log.Fatalf("unable to get vm [%s] status after powering on: [%v]", cloudInitVM.VM.Name, err)
	}
	fmt.Println("Composing the VM and powered on the VM completed. VM Status.", vmStatus)
	// cloudInitTask, err := cloudInitVM.PowerOn()
	// if err != nil {
	// 	log.Fatalf("unable to power on the cloudinit vm: %w", err)
	// }
	// if err = cloudInitTask.WaitTaskCompletion(); err != nil {
	// 	log.Fatalf("error while waiting for cloudinit vm to powerup: %w", err)
	// }
	fmt.Println("Composing the VM, cloudinit and powered on the VM completed")
}

func reconcileVMNetworks(vdcManager *vcdsdk.VdcManager, vApp *govcd.VApp, vm *govcd.VM, networks []string) error {
	connections, err := vm.GetNetworkConnectionSection()
	if err != nil {
		return errors.Wrapf(err, "Failed to get attached networks to VM")
	}

	desiredConnectionArray := make([]*types.NetworkConnection, len(networks))

	for index, ovdcNetwork := range networks {

		desiredNetworkConnection := getNetworkConnection(connections, ovdcNetwork)
		if desiredNetworkConnection.IPAddressAllocationMode == "POOL" {
			desiredNetworkConnection.IPAddressAllocationMode = "DHCP"
			desiredNetworkConnection.IPAddress = ""
			desiredNetworkConnection.NeedsCustomization = true
		}
		desiredConnectionArray[index] = getNetworkConnection(connections, ovdcNetwork)
	}

	if !containsTheSameElements(connections.NetworkConnection, desiredConnectionArray) {
		connections.NetworkConnection = desiredConnectionArray
		// update connection indexes for deterministic reconcilation
		connections.PrimaryNetworkConnectionIndex = 0
		for index, connection := range connections.NetworkConnection {
			connection.NetworkConnectionIndex = index
		}

		err = vm.UpdateNetworkConnectionSection(connections)
		if err != nil {
			return errors.Wrapf(err, "failed to update networks of VM")
		}
		// update vm.VM object for the rest of the flow, especially for getPrimaryNetwork function
		vm.VM.NetworkConnectionSection = connections
	}

	return nil
}

func getNetworkConnection(connections *types.NetworkConnectionSection, ovdcNetwork string) *types.NetworkConnection {

	for _, existingConnection := range connections.NetworkConnection {
		if existingConnection.Network == ovdcNetwork {
			return existingConnection
		}
	}

	return &types.NetworkConnection{
		Network:                 ovdcNetwork,
		NeedsCustomization:      false,
		IsConnected:             true,
		IPAddressAllocationMode: "POOL",
		NetworkAdapterType:      "VMXNET3",
	}
}

// containsTheSameElements checks all elements in the two array are the same regardless of order
func containsTheSameElements(array1 []*types.NetworkConnection, array2 []*types.NetworkConnection) bool {
	if len(array1) != len(array2) {
		return false
	}

OUTER:
	for _, element1 := range array1 {
		for _, element2 := range array2 {
			if reflect.DeepEqual(element1, element2) {
				continue OUTER
			}
		}

		return false
	}

	return true
}
