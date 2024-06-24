# Discord Docker Manager
A discord bot that facilitates managing docker containers within an environment. Utilises the docker socket to control the docker engine

[Docker Hub Image](https://hub.docker.com/repository/docker/firzen23/discord_docker_manager/general)

## Setup Requirements

1. Create a new application in the discord developor portal [Discord Dev](https://discord.com/developers/applications)
2. Click on the bot section and give it a username. Ensure you click the reset token button to generate a unique token. Make sure you keep this somewhere as it's required to run the application
3. As this bot uses message contant make sure you tick the `Message Content Intent` option
4. Select the OAuth2 option and tick the `bot` scope and copy the generated url into the browser and then select which server you want to add the bot into
5. If you have developer mode enabled in Discord, you can get the server ID by right-clicking on the server name and clicking the Copy ID button.

## Settings

### Required Settings

- `DISCORD_TOKEN` Please set this value to the token you generated after creating the discord bot
- `-v /var/run/docker.sock:/var/run/docker.sock` This passes the docker 


### Optional Settings

- `GUILD_ID` This variable is the server id that you want to deploy the application to. If this is not provided the bot will still work but is published globally instead of to a single server causing changes and deployments to take long
- `DOCKER_FILTER` This variable allows setting a [Docker Filter](https://docs.docker.com/config/filter/) that will scope down which containers are allowed to be managed. The filter provided is required to follow the same syntax as docker and multiple values can be provided seperated by a comma `,`.

    Pass a single filter that matches containers that have a label with key `color` and value `blue`
    `DOCKER_FILTER=label=color=blue`

    Pass multiple filters that matches containers that have a name of `foo` and status `running`
    `DOCKER_FILTER=name=foo,status=running`

## Running Bot Application

The application can be run in docker, docker compose or natively on a system, I have included example configurations and steps for each below.

### Docker

```
docker run --rm --name docker-manager \
-v /var/run/docker.sock:/var/run/docker.sock \
-e DISCORD_TOKEN=<your_discord_token> \
-e GUILD_ID=<your_server_id> \
-e DOCKER_FILTER=<your_docker_filter>
firzen23/discord_docker_manager:latest
```

### Docker Compsoe

```
version: '3'

services:
  docker-manager:
    container_name: docker-manager
    image: firzen23/discord_docker_manager:latest
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock #required
    environment:
      - DISCORD_TOKEN=<your_discord_token> #required
      - GUILD_ID=<your_server_id> #optional
      - DOCKER_FILTER=<your_docker_filter> #required

```

### Running natively

1. Clone the repository locally
2. Get dependencies `go get ./...`
3. Build binary `go build ./cmd -o ./docker-manager`
4. Run with your desired args `./docker-manager -token <your_discord_token> -guid <your_server_id> -filter <your_docker_filter>`

## How to use
Once the application has been setup and confirmed running you can enter a `/` to get a list of commands. Commands will automatically be generated for each server that is returned as well as the following actions for each:
1. `start` Starts the server
2. `stop` Stops the server gracefully
3. `status` Provides the current status of the server
4. `restart` Stops the server and then starts it again
