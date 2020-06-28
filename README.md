# gmin
Administration CLI for Google G Suite written in Go (aka Golang). gmin performs actions by using credentials of a service account that has delegated domain-wide authority and the email address of a user with administrative authority.

## Installing gmin
gmin is intended to run on Linux, Windows and MacOS for which pre-compiled binaries will be made available. gmin may run on other operating systems, but it is up to the user to test.

At present, there is no automated installation. Users should obtain a pre-compiled binary for their particular operating system or compile it themselves if they have a Go environment set up.

The following steps should get you a working gmin installation -

1. Create a Google Cloud Platform project.
2. Enable the Admin SDK API for the project. At the moment, the Admin SDK is all that is needed to use gmin. Additional APIs might be required when future functionality is added.
3. Create a service account, create and download a JSON key for the service account and enable G Suite domain-wide delegation.
4. In the G Suite admin console of the domain with which you want to use gmin, go to Security > API Permissions > Manage Domain-wide Delegation and add a new record (or update an existing one). The scopes that you will need to add to the record for full functionality are as follows -
```
Readonly scopes are needed for get and list functions. The other scopes are needed for create, delete, update
and undelete functions.

https://www.googleapis.com/auth/admin.directory.group
https://www.googleapis.com/auth/admin.directory.group.member.readonly
https://www.googleapis.com/auth/admin.directory.orgunit
https://www.googleapis.com/auth/admin.directory.orgunit.readonly
https://www.googleapis.com/auth/admin.directory.user
https://www.googleapis.com/auth/admin.directory.user.readonly
https://www.googleapis.com/auth/admin.directory.group.readonly
https://www.googleapis.com/auth/admin.directory.group.member
```

5. Copy/move the gmin binary to a convenient directory/folder and rename the JSON key file, downloaded earlier, to gmin_credentials and place in the same directory/folder as the gmin binary.
6. Run the command `gmin init` inputting the email address of the admin whose privileges will be used and optionally, the path where you would like the config file, .gmin.yaml, to be written. By default the config file is written to the current user's home directory. If you choose a different installation path then that path will need to be given with each gmin command by using the --config flag.
7. To see the version number of your gmin binary, run the command `gmin -v` or `gmin --version`.
8. To get help from gmin itself, enter `gmin -h` or `gmin --help` and go from there.

If you already use GAM (https://github.com/jay0lee/GAM) then you will already have a service account and JSON credentials file. In this case you could use the same service account by copying the GAM credentials.json, rename the copy to gmin_credentials and place it in the same directory/folder as the gmin executable.

## Usage
Commands usually take the form of `gmin <verb> <object> <flags>`. Although there may be the odd exception like `gmin whoami`.

Command flags look like -x or --longnameforx. Flags that relate to boolean values (true/false) just need the flag to be provided and no value. For example, the `-c or --changepassword` flag when creating a user.

Type `gmin <command> -h or --help` to get help about a particular command. For example, to get help about creating a new user you could enter `gmin create user -h`. Help text will be improved over time.

**Create a new user**

`gmin create user new.user@mydomain.com --firstname New --lastname User --password MyStrongPassword`

with abbreviations the command would look like this -

`gmin crt user new.user@mydomain.com -f New -l User -p MyStrongPassword`

**Get a user**

`gmin get user new.user@mydomain.com`

This command results in all user information being returned. If I only want their addresses then I would say -

`gmin get user new.user@mydomain.com -a addresses`

If I only want a particular part of the address, because address is made up of other attributes, I could say something like -

`gmin get user new.user@mydomain.com -a addresses(formatted)`

**List users**

`gmin list users`

This command returns all user information about all users. If I want to restrict the amount of information returned then I can specify some attributes -

`gmin list users -a primaryemail~name~addresses`

If I want to filter the results still further I can provide a query with or without attributes -

`gmin list users -q orgunitpath=/Sales`

**Delete a user**

`gmin delete user new.user@mydomain.com`

There are no warnings issued by the delete command so use it carefully.

**Undelete a user**

`gmin undelete user new.user@mydomain.com`

This allows you to restore a user that has been deleted.

**Attributes Flag**

Some commands have an attribute flag (-a or --attributes) where you can specify object attributes that you want to provide or that you want to see. For instance, where you are creating a user you can provide all or some of the user attributes that you want to create the user with via the attribute flag.

Attributes are written in a particular format similar to but simpler than JSON. The rules are as follows -

* Attribute names and values are separated by the colon character (:)
* Attributes are separated by the tilde character (~)
* If any attributes contain spaces then the whole attribute string after the -a or --attributes will need to be quoted and you will also need to use quotes around values that contain spaces
* Composite attributes, those that have other attributes contained within them, have their values contained within {} braces

Therefore the user above could be created by using the following command -

`gmin create user new.user@domain.com -a name:{firstname:New~lastname:User}~password:MyStrongPassword`

When creating objects, values provided by the attributes flag override attributes provided by other flags. For example -

`gmin crt user d.williams@mycompany.org -f Danny -l Williams -p SuperSecretPwd -a name:{firstname:Douglas}`

would result in a user whose first name is Douglas not Danny.

https://developers.google.com/admin-sdk/directory/v1/reference is a useful resource for looking up valid attribute names and values.

**Query Flag**

Some commands have a query flag (-q or --query) where you can specify query clauses to filter the returned results by. For instance, you might want to retrieve a list of users that have certain attributes -

`gmin list users -q isadmin=true`

This command will return a list of users that have super admin privileges. You could limit the amount of output by specifying the attributes that you want to see for each user, like this -

`gmin list users -q isadmin=true -a primaryemail`

This command will only return the primary email address for each user that satisfies the query.

Similarly to the attributes flag above, query clauses are separated by the tilde (~) character and quotation marks may need to be used. This command returns a list of users whose last name is Smith and has an address in London -

`gmin list users -q lastname=Smith~addressLocality=London`

https://developers.google.com/admin-sdk/directory/v1/get-start/getting-started is a useful resource for looking up query parameters. There are 'Search for' links for different objects like Users and Groups.

## Why am I writing gmin

* I want to write something non-trivial in Go
* I want to gain a deep understanding of Google APIs, particularly those useful to G Suite admins
* I want to write something that would help me in my daily G Suite admin work 
* I want to create something that can be distributed as a single executable
* I want a G Suite admin tool that is more intuitive and easier to work with (at least for me) than existing tools
* Maybe there are other people who might benefit from gmin

## Project Status
gmin is pretty young which means that it is liable to rapid change, although the command syntax is unlikely to change much (if at all) over time. The functionality is currently limited to users, groups, members and organisation units but additional functionality will be added when it is ready.

All output is in JSON format apart from informational and error messages. Output in other formats such as CSV is on the roadmap, however, I have found the use of the jq utility (https://stedolan.github.io/jq/) can be a great help in working with JSON.

I will be publishing a roadmap and welcome any suggestions as to the most important features to add. A Wiki is on the list of tasks on the upcoming roadmap.

## Community

Google Group: https://groups.google.com/a/plusworx.uk/d/forum/gmin

## Acknowledgements

Thank you to Jay Lee for writing GAM (https://github.com/jay0lee/GAM) which inspired me to have a go at gmin.

Thanks to the Go team for creating a programming language that is really enjoyable to use.

Thanks to the G Suite team for building a world-class collaboration platform.


## License
This software is made available under the MIT license.
