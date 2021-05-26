### ASSUME

### Context 
This module is used to ease the changing of AWS roles from the command line.

It is a common setup in AWS projects to have to switch between AWS roles frequently. This can be done by setting environment variables
or updating configs. For a project that spans many services, across multiple environments, with roles specified for each one, this service
is designed to ease that process.

The idea is that the tool can be called with human-readable names associated with each role in a config or hardcoded into this codebase. The tool then sets the role profile approriately.

### Usage 
Assume allows you to assume an AWS role temporarily. The command can take an ARN or use a preconfigured list of roles specified in the source code. 

To see the list of roles, use list 
`assume list`

To assume a role from the list, simply specify the name of a profile to use from your .aws/credentials file and the name of the role to assume. The below command will use my profile "main" to assume the "arn:aws:iam::XXX/YYY" role.
`assume -p main`

To use an ARN specify the profile and the arn using the "role" or "r" flag.
`assume -p cf -r arn:aws:iam::XXX/YYY`

The code can easily be extended to incorporate the use of a JSON configuration file, similar to how the previous Python code would. 

Prereqs:
    1. The AWS profile you specify must have the permissions to assume the targeted role.
    2. Your .aws/credentials file must be set up appropriately. An example corresponding to the above commands is below.

[main]
aws_access_key_id = <ACCESS KEY ID>
aws_secret_access_key = <ACCESS KEY>