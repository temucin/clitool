/**
**/

package assume

import (
	"clitool/utils"
	"clitool/utils/CmdRegistry"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/bigkevmcd/go-configparser"
)

var AssumeFlagSet flag.FlagSet
var role string
var profile string
var secretAccessKey string
var unassume bool
var roleName string
var homeDir, _ = os.UserHomeDir()
var workingDir, _ = os.Getwd()

//Flag constants
const (
	moduleUsage                 = "Assumes an AWS role and updates the users credentials file with the session token for that role. Use help to see what flags to use."
	defaultRole                 = ""
	roleUsage                   = "Specify which ARN role to assume"
	defaultProfile              = ""
	profileUsage                = "Specifies which profile in your ~/.aws/credentials file to use when requesting the role. Also used to retrieve the role ARN from the JSON config if thee role or roleName flag is unspecified."
	roleNameUsage               = "Specifies which role name to use from the hardcoded RoleArns in the CLI. Use \"list\" command to see these roles."
	credsFileAwsAccessKeyId     = "aws_access_key_id"
	credsFileAwsSecretAccessKey = "aws_secret_access_key"
	credsFileAwsSessionToken    = "aws_session_token"
)

var roleArns = map[string]string{
	"main": "arn:aws:iam::XXX/YYY",
}

var AssumeCmd = CmdRegistry.Cmd{
	Name:    "assume",
	RunCmd:  runAssume,
	FlagSet: AssumeFlagSet,
}

func init() {
	//Set up flags
	AssumeFlagSet = *flag.NewFlagSet("assume", flag.ContinueOnError)
	AssumeFlagSet.Usage = func() { fmt.Print(moduleUsage) }
	AssumeFlagSet.StringVar(&role, "role", defaultRole, roleUsage)
	AssumeFlagSet.StringVar(&role, "r", defaultRole, "Shortcut for role")
	AssumeFlagSet.StringVar(&profile, "profile", defaultProfile, profileUsage)
	AssumeFlagSet.StringVar(&profile, "p", defaultProfile, "Shortcut for profile")
	AssumeFlagSet.BoolVar(&unassume, "unassume", false, "Removes session token and resets default values to selected profile keys")
	AssumeFlagSet.StringVar(&roleName, "roleName", "", roleNameUsage)
	AssumeFlagSet.StringVar(&roleName, "n", "", "Shortcut for roleName")

	//Register command
	CmdRegistry.RegisterCmd(AssumeCmd)
	CmdRegistry.RegisterFlagSet(AssumeFlagSet)
}

func runAssume() {
	AssumeFlagSet.Parse(CmdRegistry.CmdArgs())

	switch CmdRegistry.CmdArgs()[0] {
	case "help":
		AssumeFlagSet.PrintDefaults()
	case "list":
		printRoleArns()
	default:
		if validateArgsAndFlags() != 0 {
			return
		}
		assumeRole()
	}

	cleanUp()
}

func printRoleArns() {
	fmt.Println("The following roles are available.")
	fmt.Println("[ Name to use : Corresponding Role ARN ]")
	for key, val := range roleArns {
		fmt.Println("[", key, ": ", val, "]")
	}
}

func cleanUp() {
	role = ""
	profile = ""
	secretAccessKey = ""
	unassume = false
	roleName = ""
}

func validateArgsAndFlags() int {
	if profile == "" {
		fmt.Println("Error! You need to specify a profile from your AWS Credentials file to use when assuming a role.")
		return 1
	}

	return 0
}

func getProfile(profile string, credsFile *os.File) (string, string) {
	credsFileName := credsFile.Name()
	credsProvider := credentials.SharedCredentialsProvider{
		Filename: credsFileName,
		Profile:  profile,
	}
	profileValue, err := credsProvider.Retrieve()
	if err != nil {
		fmt.Println("Error getting profile from %s", credsFileName)
	}
	return profileValue.AccessKeyID, profileValue.SecretAccessKey
}

func getRoleArn(profile string) string {
	configFile, err := os.Open(workingDir + "config.json")
	if err != nil {
		fmt.Println("Error reading config file. Please make sure that \"config.json\" exists and is readable.")
		log.Fatal(err)
	}
	defer configFile.Close()
	byteValue, _ := ioutil.ReadAll(configFile)
	var configJson map[string]map[string]string
	json.Unmarshal([]byte(byteValue), &configJson)
	roleArn := configJson[profile]["role_arn"]
	fmt.Printf("RoleArn to assume %v\n", roleArn)
	return roleArn
}

func updateCreds(credsFile *os.File, keyID string, secretKey string, sessToken string) {
	defer credsFile.Close()
	config, err := configparser.NewConfigParserFromFile(credsFile.Name())
	if err != nil {
		fmt.Println("Error reading credentials file!", err)
	}

	if !config.HasSection("default") {
		fmt.Println("Creating default section in AWS credentials file and updating section.")
		config.AddSection("default")
	}

	err = config.Set("default", credsFileAwsAccessKeyId, keyID)
	if err != nil {
		fmt.Println("Error updating access key id in credentials file.")
	}
	err = config.Set("default", credsFileAwsSecretAccessKey, secretKey)
	if err != nil {
		fmt.Println("Error updating secret access key in credentials file.")
	}

	if sessToken != "" {
		err = config.Set("default", credsFileAwsSessionToken, sessToken)
		if err != nil {
			fmt.Println("Error updating session token in credentials file.")
		}
	} else {
		err = config.RemoveOption("default", credsFileAwsSessionToken)
	}

	config.SaveWithDelimiter(credsFile.Name(), "=")
}

func getCredsFile() *os.File {
	credsFile, err := os.Open(homeDir + "/.aws/credentials")
	if err != nil {
		fmt.Println("Error reading config file. Please make sure that \"" + homeDir + "/.aws/credentials\" exists and is readable.")
	}

	backupCredsFile, err := os.Open(homeDir + "/.aws/credentials.bkp")
	if err != nil {
		backupCredsFile, err := os.Create(homeDir + "/.aws/credentials.bkp")
		fmt.Println("Creating backup credentials file.")
		_, err = io.Copy(backupCredsFile, credsFile)
		if err != nil {
			fmt.Println("Error copying to backup credentials file.", backupCredsFile)
		}
		backupCredsFile.Sync()
	}
	defer backupCredsFile.Close()

	return credsFile

}

func assumeRole() {
	if validateArgsAndFlags() != 0 { //Validate input
		return
	}
	credsFile := getCredsFile()
	profileKeyId, profileSecretKey := getProfile(profile, credsFile) //Reads credentials file to get access key based on profile input

	//If unassume flag is used, we simply update the default key values to "reset" the role
	if unassume {
		fmt.Println("Resetting default credentials.")
		if profile == "" {
			fmt.Println("Please specify the AWS profile to reset the default credentials with.")
			return
		}
		updateCreds(credsFile, profileKeyId, profileSecretKey, "")
		fmt.Println("Default credentials updated with", profile, "profile.")
	} else {
		if role == "" && roleName == "" {
			fmt.Println("Using config.json to determine role to assume.")
			role = getRoleArn(profile) //Get role arn from swap-profile config.json if no role ARN is specified
		} else if roleName != "" {
			role = roleArns[roleName]
			if role == "" {
				fmt.Println("Error! The role name you specified does not exist.")
				return
			}
			fmt.Println("Assuming ", role)
		} else if role != "" {
			fmt.Println("Assuming ", role)
		}

		assumeResults, err := utils.AssumeRole(role, profile, profileKeyId) //execute sts assume-role command
		if err != nil {
			fmt.Println("Error assuming role!", err)
		}

		fmt.Println("Role assumed!", assumeResults.AssumedRoleUser)
		// expirationTime := assumeResults.Credentials.AccessKeyId
		keyID := assumeResults.Credentials.AccessKeyId
		secretKey := assumeResults.Credentials.SecretAccessKey
		sessToken := assumeResults.Credentials.SessionToken
		fmt.Printf("Expires at %v\n", *assumeResults.Credentials.Expiration)
		// fmt.Printf("%v\n%v\n%v\n", *keyID, *secretKey, *sessToken)
		updateCreds(credsFile, *keyID, *secretKey, *sessToken) //update credentials file or env var
	}

}
