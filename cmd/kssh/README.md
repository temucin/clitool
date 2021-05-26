### KSSH
KSSH is a utility for using MSSH and MSFTP. It will allow you to SSH into an EC2 instance by specifying the instance tags. This makes it easier to manage the movement between multiple servers by only having to remember the specific characteristics (things like the environment, server usage, etc.)

When multiple instance IDs are returned, the application will ask you to select one to use.

KSFTP is combined with this command as it uses the exact same logic, but execute the MSFTP command instead.

Prereqs:
    1. You must have assumed a role that allows you to execute the describe-instances call.


*Note:* The K stands for Khan, the authors last name :) 