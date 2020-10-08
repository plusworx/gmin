![Go](https://github.com/plusworx/gmin/workflows/Go/badge.svg)  [![Go Report Card](https://goreportcard.com/badge/github.com/plusworx/gmin)](https://goreportcard.com/report/github.com/plusworx/gmin)

# gmin
gmin is a Google Workspace administration CLI (command line interface) written in Go (aka Golang). gmin performs actions by using credentials of a service account that has delegated domain-wide authority and the email address of a user with administrative authority in the Workspace domain.

There is a gmin [documentation website](https://gmin.plusworx.uk) that is continually updated.

## Installing gmin
See [Installing gmin](https://gmin.plusworx.uk/#/getting_started?id=installing-gmin) page.

gmin is intended to run on Linux, Windows and MacOS for which pre-compiled binaries are [available](https://github.com/plusworx/gmin/releases). gmin may run on other operating systems, but it is up to the user to test.

At present, there is no automated installation. Users should obtain a pre-compiled binary for their particular operating system or compile it themselves if they have a Go environment set up.

## Quick start
Commands usually take the form of `gmin <verb> <object> <flags>`. Although there may be the odd exception like `gmin whoami`.

Command flags can have short names like `-x` and always have a a long name like `--longnameforx`. Flags that relate to boolean values (true/false) can be set like this -

`--changepassword=true`

When setting a boolean value to false the syntax above must be used, however, when setting a value to true, the value is optional. The above flag could be set to true by simply using the flag name like this -

`--changepassword`

Type `gmin <command> -h or --help` to get help about a particular command. For example, to get help about creating a new user you could enter `gmin create user -h`. Help text will be improved over time.

### Some usage examples

**Create a new user**

`gmin create user new.user@mydomain.com --firstname New --lastname User --password MyStrongPassword`

using flag short names the above command would look like this -

`gmin crt user new.user@mydomain.com -f New -l User -p MyStrongPassword`

You can set any user object attribute via the attributes (-a or --attributes) flag by providing a JSON string.

Therefore the user above could be created by using the following command -

`gmin create user new.user@domain.com -a '{"name":{"givenName":"New","familyName":"User"},"password":"MyStrongPassword"}'`

When creating objects, values provided by the attributes flag are overriden by attributes provided by other flags. For example -

`gmin crt user d.williams@mycompany.org -f Danny -l Williams -p SuperSecretPwd -a '{"name":{"givenName":"Douglas"}}'`

would result in a user whose first name is Danny not Douglas.

**Get a user**

`gmin get user new.user@mydomain.com`

This command results in information for one user being returned. If I only want their addresses then I would say -

`gmin get user new.user@mydomain.com -a addresses`

If I only want a particular part of the address, because address is made up of other attributes, I could say something like -

`gmin get user new.user@mydomain.com -a 'addresses(formatted)'`

**List users**

`gmin list users`

This command returns user information about all users. If I want to restrict the amount of information returned then I can specify some attributes -

`gmin list users -a primaryemail~name~addresses`

and only the primaryEmail, name and addresses attributes will be returned for each user. If an attribute is empty (either doesn't exist or is blank/false), then it is NOT displayed.

If I want to filter the results still further I can provide a query with or without attributes -

`gmin list users -a primaryemail~name~addresses -q orgunitpath=/Sales`

and only information about users in the Sales organisation unit will be returned.

**Delete a user**

`gmin delete user new.user@mydomain.com`

**Undelete a user**

`gmin undelete user new.user@mydomain.com`

This allows you to restore a user that has been deleted.

### Attributes Flag

**Get and list commands**

These commands have an attribute flag (-a or --attributes) where you can specify particular object attributes that you want to see in the results. If the attribute is not present (including false or empty) then it will not be displayed.

**Create and update commands**

Where an attribute flag (-a or --attributes) is provided for create or update commands, the value can be any valid JSON string that provides attribute values. Please note that if you are providing empty or false values you will need to use the --force flag with the field names separated by '~' otherwise those fields will be ignored.

### Query Flag

Some commands have a query flag (-q or --query) where you can specify query clauses to filter the returned results by. For instance, you might want to retrieve a list of users that have certain attributes -

`gmin list users -q isadmin=true`

This command will return a list of users that have super admin privileges. You could limit the amount of output by specifying the attributes that you want to see for each user, like this -

`gmin list users -q isadmin=true -a primaryemail`

This command will only return the primary email address for each user that satisfies the query.

Similarly to the attributes flag above, query clauses are separated by the tilde (~) character and quotation marks may need to be used. This command returns a list of users whose last name is Smith and have an address in London -

`gmin list users -q lastname=Smith~addressLocality=London`

## Why am I writing gmin

* I want to write something non-trivial in Go
* I want to gain a deep understanding of Google APIs, particularly those useful to Google Workspace admins
* I want to write something that would help me in my daily Google Workspace admin work 
* I want to create something that can be distributed as a single executable
* I want a Google Workspace admin tool that is more intuitive and easier to work with (at least for me) than existing tools
* I think that other Google Workplace admins might find gmin useful

## Project Status
gmin is young, but maturing by the day. The command syntax is unlikely to change much over time, but there may be some breaking changes which I will notify you about in the release notes. Functionality will be added when it is ready.

I have published a [development roadmap](https://https://gmin.plusworx.uk/#/dev_roadmap) and welcome any suggestions as to the most important features to add.

## Community

Google Group: https://groups.google.com/a/plusworx.uk/d/forum/gmin

Reddit: https://www.reddit.com/r/gmin/

## Code of Conduct
Please see the code of conduct document - https://github.com/plusworx/gmin/blob/master/CODE_OF_CONDUCT.md

## Acknowledgements

Thank you to Jay Lee for writing GAM (https://github.com/jay0lee/GAM) which inspired me to have a go at gmin.

Thanks to the Go team for creating a programming language that is really enjoyable to use.

Thanks to the Google Workspace team for building a world-class collaboration platform.

Last, but not least, thanks to all of the creators and maintainers of tools and libraries that I use to create and maintain gmin. Without your work I would not be able to do anything.

## License
This software is made available under the MIT license.
