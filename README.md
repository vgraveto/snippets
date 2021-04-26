# snippets
Template Web site and REST API to manage snippets and users.

Folder **cmd/api** includes de *REST API* and folder **cmd/web** has the web site that uses the *API* to get to the *MySQL* database.

In folder **pkg** you will find the common data packages used by both, like data models. The **ui** folder incudes the *html* templates and all the needed static files.

Use the command *''make swagger´´* to build the ***swagger.yaml*** file from the existing metadata that is provided as comments in source code.

## Database
The database ***ca-certificate.crt*** should be added to **certs**  folder and the config file ***snippetsAPI.toml*** should be reviewed to include te correct ''url'', ''database name'' and ''password''.

Use ***CreateSnippetsDatabase.sql*** *sql* file to build your database tables in your MySql server, after creating your database schema.