package kssh

import (
	"bufio"
	utils "clitool/utils"
	"clitool/utils/CmdRegistry"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

var KsshCmd = CmdRegistry.Cmd{
	Name:    "kssh",
	RunCmd:  func() { RunMssh(false) },
	FlagSet: KsshFlagSet,
}

var KsftpCmd = CmdRegistry.Cmd{
	Name:    "ksftp",
	RunCmd:  func() { RunMssh(true) },
	FlagSet: KsshFlagSet,
}

var KsshFlagSet flag.FlagSet
var TargetName string
var env string
var app string

const (
	defaultEnv    = "DEV"
	envUsage      = "Specify the environment to query about."
	defaultTarget = "ExampleTarget"
	TargetUsage   = "Specify the target data to query about."
	defaultApp    = "ExampleApp"
	appUsage      = "Application to query about. (Frontend, database, etc.)"
	moduleUsage   = "The KSSH/KSFTP command will execute the MSSH or MSFTP for the Ubuntu user against the Instance ID specified by the command arguments."
)

func init() {
	KsshFlagSet = *flag.NewFlagSet("kssh", flag.ContinueOnError)
	KsshFlagSet.Usage = func() { fmt.Print(moduleUsage) }
	KsshFlagSet.StringVar(&TargetName, "target", defaultTarget, TargetUsage)
	KsshFlagSet.StringVar(&TargetName, "t", defaultTarget, "Shorthand -Target")
	KsshFlagSet.StringVar(&env, "env", defaultEnv, envUsage)
	KsshFlagSet.StringVar(&env, "e", defaultEnv, "Shorthand -env")
	KsshFlagSet.StringVar(&app, "app", defaultApp, appUsage)
	KsshFlagSet.StringVar(&app, "a", defaultApp, "Shorthand -app")
	CmdRegistry.RegisterCmd(KsshCmd)
	CmdRegistry.RegisterCmd(KsftpCmd)
	CmdRegistry.RegisterFlagSet(KsshFlagSet)
}

func validateArgsAndFlags() int {
	if KsshFlagSet.Arg(0) == "help" {
		KsshFlagSet.PrintDefaults()
		return 1
	}

	if TargetName == "" {
		fmt.Println("ERROR: Target name flag not set.")
		return 1
	}

	return 0
}

func RunMssh(withSftp bool) {
	KsshFlagSet.Parse(CmdRegistry.CmdArgs())
	if validateArgsAndFlags() != 0 {
		return
	}

	fmt.Printf("Getting instance ID for %v %v in %v\n", TargetName, app, env)
	describeFilter := []*ec2.Filter{
		{Name: aws.String("instance-state-name"), Values: []*string{aws.String("running")}},
		{Name: aws.String("tag:Target"), Values: []*string{aws.String(strings.Title(strings.ToLower(TargetName)))}},
		{Name: aws.String("tag:AppName"), Values: []*string{aws.String(app)}},
		{Name: aws.String("tag:Environment"), Values: []*string{aws.String(strings.ToUpper(env))}},
	}
	descOutput, descErr := utils.GetInstances(describeFilter)
	if descErr != nil {
		fmt.Println("Error getting instance information!", descErr)
		return
	}

	reservationList := descOutput.Reservations
	var iid string
	resLen := len(reservationList)
	if resLen > 1 {
		iid = getUserInput(reservationList)
	} else if resLen == 1 {
		iid = aws.StringValue(reservationList[0].Instances[0].InstanceId)
	} else if resLen == 0 {
		fmt.Println("No instances found!")
		return
	} else {
		fmt.Println("Failed to get an instance ID.")
		return
	}

	fmt.Printf("Instance ID: \n%v\n", iid)
	cmdString := fmt.Sprintf("ubuntu@%v", iid) //Execute mssh command using Instance ID from previous step
	var cmd *exec.Cmd
	if withSftp == true {
		fmt.Println("Executing msftp...")
		cmd = exec.Command("msftp", cmdString)
	} else {
		fmt.Println("Executing mssh...")
		cmd = exec.Command("mssh", cmdString)
	}
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Run()
}

func getUserInput(reservationList []*ec2.Reservation) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Multiple instance IDs found. Please select one.")
	for index, value := range reservationList {
		fmt.Printf("\n[%v]: %v", index, aws.StringValue(value.Instances[0].InstanceId))
	}
	fmt.Printf("\n->")

	input, _, err := reader.ReadRune()
	if err != nil {
		fmt.Println("Error processing input!", err)
	}

	return aws.StringValue(reservationList[int(input-'0')].Instances[0].InstanceId)
}
